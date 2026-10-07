// Package pagos holds the payments module's PostgreSQL repository.
package pagos

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type RepositorioIntentos struct{ db *gorm.DB }

func NuevoRepositorioIntentos(db *gorm.DB) *RepositorioIntentos {
	return &RepositorioIntentos{db: db}
}

func (r *RepositorioIntentos) Crear(ctx context.Context, intento *domain.IntentoPago) error {
	return r.db.WithContext(ctx).Create(intento).Error
}

// PorReferenciaExterna busca un intento ya registrado para esa notificación
// del proveedor. El índice único de (proveedor, referencia_externa) es lo
// que garantiza que esta consulta nunca encuentre dos filas.
func (r *RepositorioIntentos) PorReferenciaExterna(ctx context.Context, proveedor domain.ProveedorPago, referenciaExterna string) (domain.IntentoPago, bool, error) {
	var intento domain.IntentoPago
	err := r.db.WithContext(ctx).
		Where("proveedor = ? AND referencia_externa = ?", proveedor, referenciaExterna).
		First(&intento).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.IntentoPago{}, false, nil
	}
	if err != nil {
		return domain.IntentoPago{}, false, err
	}
	return intento, true, nil
}

// UltimoPorPedido devuelve el intento más reciente de un pedido: es lo que
// GET /api/pedidos/{id}/pago reporta.
func (r *RepositorioIntentos) UltimoPorPedido(ctx context.Context, pedidoID uuid.UUID) (domain.IntentoPago, error) {
	var intento domain.IntentoPago
	err := r.db.WithContext(ctx).
		Where("pedido_id = ?", pedidoID).
		Order("creado_en DESC").
		First(&intento).Error
	return intento, err
}

func (r *RepositorioIntentos) CambiarEstado(ctx context.Context, id uuid.UUID, nuevo domain.EstadoIntentoPago) error {
	return r.db.WithContext(ctx).
		Model(&domain.IntentoPago{}).
		Where("id = ?", id).
		Update("estado", nuevo).Error
}
