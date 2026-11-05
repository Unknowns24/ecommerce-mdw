package pagos

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/external/mp"
	usecases "github.com/Unknowns24/ecommerce-mdw/internal/usecases/inventario"
	reglas "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pagos"
)

const (
	intervaloVencimiento = time.Minute
	loteVencimiento      = 50
	motivoVencimiento    = "Vencimiento automático: no se acreditó el pago de Mercado Pago dentro del plazo de la reserva"
)

// consultaPagos es lo que el vencimiento necesita de Mercado Pago. Es una
// interfaz para que la prueba no salga a la red.
type consultaPagos interface {
	EstadosDePagosDelPedido(ctx context.Context, pedidoID string) ([]string, error)
}

// IniciarVencimiento cancela, cada minuto, los pedidos de Mercado Pago cuya
// reserva de 15 minutos se agotó, y libera su stock (H14). main.go lo llama
// una vez al levantar el servidor; corre hasta que ctx se cancela.
func IniciarVencimiento(ctx context.Context, db *gorm.DB, cfg config.Config) {
	cliente := mp.NuevoCliente(cfg.MercadoPagoAccessToken, cfg.PublicBaseURL)
	stock := usecases.NuevoServicioStock(db)
	go func() {
		ticker := time.NewTicker(intervaloVencimiento)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if n, err := vencerPendientes(ctx, db, stock, cliente, time.Now()); err != nil {
					log.Printf("vencimiento: %v", err)
				} else if n > 0 {
					log.Printf("vencimiento: %d pedidos cancelados por reserva agotada", n)
				}
			}
		}
	}()
}

// vencerPendientes cancela los pedidos vencidos y devuelve cuántos canceló.
// Un temporizador local no basta para asumir que no hubo pago (spec, sección
// 6): antes de cancelar se le pregunta a Mercado Pago, y si hay un pago que
// todavía puede acreditarse —o el proveedor no responde— el pedido se deja
// para la próxima vuelta.
func vencerPendientes(ctx context.Context, db *gorm.DB, stock *usecases.ServicioStock, pagos consultaPagos, ahora time.Time) (int, error) {
	var ids []uuid.UUID
	err := db.WithContext(ctx).Table("pedido").
		Where("medio_pago = 'MERCADO_PAGO' AND estado_pedido = 'PENDIENTE_DE_PAGO' AND vence_en <= ?", ahora).
		Order("vence_en").Limit(loteVencimiento).
		Pluck("id", &ids).Error
	if err != nil {
		return 0, err
	}

	cancelados := 0
	for _, id := range ids {
		estados, err := pagos.EstadosDePagosDelPedido(ctx, id.String())
		if err != nil {
			log.Printf("vencimiento: no se pudo consultar Mercado Pago para el pedido %s, se reintenta: %v", id, err)
			continue
		}
		if reglas.HayPagoActivo(estados) {
			continue
		}

		cancelado, err := vencerPedido(ctx, db, stock, id, ahora)
		if err != nil {
			log.Printf("vencimiento: pedido %s: %v", id, err)
			continue
		}
		if cancelado {
			cancelados++
		}
	}
	return cancelados, nil
}

// vencerPedido cancela un pedido y libera su stock en una transacción, con la
// fila bloqueada: si el webhook lo confirmó mientras tanto, no hace nada.
func vencerPedido(ctx context.Context, db *gorm.DB, stock *usecases.ServicioStock, id uuid.UUID, ahora time.Time) (bool, error) {
	cancelado := false
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pedido, existe, err := pedidoPorID(ctx, tx, id, true)
		if err != nil || !existe {
			return err
		}
		if pedido.EstadoPedido != "PENDIENTE_DE_PAGO" || !reglas.Vencido(pedido, ahora) {
			return nil
		}
		// cancelado_por queda nulo: lo canceló el sistema, no una persona.
		res := tx.Exec(
			`UPDATE pedido SET estado_pedido = 'CANCELADO', cancelado_en = ?, motivo_cancelacion = ?, actualizado_en = ?
			 WHERE id = ? AND estado_pedido = 'PENDIENTE_DE_PAGO'`,
			ahora, motivoVencimiento, ahora, id,
		)
		if res.Error != nil || res.RowsAffected == 0 {
			return res.Error
		}
		if err := stock.Liberar(ctx, tx, id); err != nil {
			return err
		}
		cancelado = true
		return nil
	})
	return cancelado, err
}
