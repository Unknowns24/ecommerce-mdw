// Package pagos mounts the Mercado Pago endpoints: iniciar un pago, recibir
// el webhook, y consultar el estado del último intento de un pedido.
package pagos

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Unknowns24/ecommerce-mdw/config"
	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/pagos"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/external/mp"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	usecases "github.com/Unknowns24/ecommerce-mdw/internal/usecases/inventario"
)

type handler struct {
	db            *gorm.DB
	intentos      *repo.RepositorioIntentos
	stock         *usecases.ServicioStock
	mp            *mp.Cliente
	webhookSecret string
}

func nuevoHandler(db *gorm.DB, cfg config.Config) *handler {
	return &handler{
		db:            db,
		intentos:      repo.NuevoRepositorioIntentos(db),
		stock:         usecases.NuevoServicioStock(db),
		mp:            mp.NuevoCliente(cfg.MercadoPagoAccessToken, cfg.PublicBaseURL),
		webhookSecret: cfg.MercadoPagoWebhookSecret,
	}
}

// pedidoPorID lee del pedido (módulo de Nicolás) solo las columnas que las
// reglas de pago necesitan: no hay un lector congelado para pedidos, así que
// se consulta la tabla directamente, igual que inventario hace con variante y
// producto. Con bloquear, toma la fila con FOR UPDATE: el webhook y el
// vencimiento automático no pueden decidir los dos sobre el mismo pedido a la
// vez.
func pedidoPorID(ctx context.Context, db *gorm.DB, id uuid.UUID, bloquear bool) (domain.Pedido, bool, error) {
	var p domain.Pedido
	consulta := db.WithContext(ctx).
		Select("id, usuario_id, estado_pedido, estado_pago, medio_pago, envio_centavos, total_centavos, token_acceso, vence_en").
		Where("id = ?", id)
	if bloquear {
		consulta = consulta.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := consulta.Take(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Pedido{}, false, nil
	}
	if err != nil {
		return domain.Pedido{}, false, err
	}
	return p, true, nil
}

type lineaPedido struct {
	Codigo                 string
	Nombre                 string
	Unidades               int
	PrecioUnitarioCentavos int64
}

func lineasDePedido(ctx context.Context, db *gorm.DB, pedidoID uuid.UUID) ([]lineaPedido, error) {
	var lineas []lineaPedido
	err := db.WithContext(ctx).
		Table("detalle_pedido").
		Select("codigo, nombre, unidades, precio_unitario_centavos").
		Where("pedido_id = ?", pedidoID).
		Scan(&lineas).Error
	return lineas, err
}

// confirmarPedido marca el pedido como CONFIRMADO con pago APROBADO. Es
// idempotente: el WHERE solo toma pedidos que todavía esperan el pago, así que
// una segunda notificación no pisa nada, y un pedido vencido o cancelado
// nunca vuelve a confirmarse por este camino.
func confirmarPedido(ctx context.Context, tx *gorm.DB, pedidoID uuid.UUID) error {
	return tx.WithContext(ctx).Exec(
		`UPDATE pedido SET estado_pedido = 'CONFIRMADO', estado_pago = 'APROBADO', actualizado_en = ?
		 WHERE id = ? AND estado_pedido = 'PENDIENTE_DE_PAGO' AND estado_pago <> 'APROBADO'`,
		time.Now(), pedidoID,
	).Error
}
