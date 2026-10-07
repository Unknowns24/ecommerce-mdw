// Package pagos mounts the Mercado Pago endpoints: iniciar un pago, recibir
// el webhook, y consultar el estado del último intento de un pedido.
package pagos

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/pagos"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/external/mp"
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

// pedidoParaPago es lo mínimo del pedido (módulo de Nicolás) que este
// módulo necesita para abrir una preferencia: no hay un lector congelado
// para pedidos todavía, así que se consulta la tabla directamente, igual
// que inventario hace con variante y producto.
type pedidoParaPago struct {
	ID            uuid.UUID
	EstadoPago    string
	MedioPago     string
	TotalCentavos int64
}

func pedidoPorID(ctx context.Context, db *gorm.DB, id uuid.UUID) (pedidoParaPago, bool, error) {
	var p pedidoParaPago
	err := db.WithContext(ctx).
		Table("pedido").
		Select("id, estado_pago, medio_pago, total_centavos").
		Where("id = ?", id).
		Scan(&p).Error
	if err != nil {
		return pedidoParaPago{}, false, err
	}
	if p.ID == uuid.Nil {
		return pedidoParaPago{}, false, nil
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
// idempotente: el WHERE excluye los pedidos que ya están aprobados, así que
// una segunda notificación no pisa nada ni dispara un UPDATE de más.
func confirmarPedido(ctx context.Context, tx *gorm.DB, pedidoID uuid.UUID) error {
	return tx.WithContext(ctx).Exec(
		`UPDATE pedido SET estado_pedido = 'CONFIRMADO', estado_pago = 'APROBADO', actualizado_en = ?
		 WHERE id = ? AND estado_pago <> 'APROBADO'`,
		time.Now(), pedidoID,
	).Error
}
