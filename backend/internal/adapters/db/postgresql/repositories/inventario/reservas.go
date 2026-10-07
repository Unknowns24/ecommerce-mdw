package inventario

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type RepositorioReservas struct{ db *gorm.DB }

func NuevoRepositorioReservas(db *gorm.DB) *RepositorioReservas {
	return &RepositorioReservas{db: db}
}

func (r *RepositorioReservas) Crear(ctx context.Context, reserva *domain.ReservaStock) error {
	return r.db.WithContext(ctx).Create(reserva).Error
}

func (r *RepositorioReservas) ActivasPorPedido(ctx context.Context, pedidoID uuid.UUID) ([]domain.ReservaStock, error) {
	var reservas []domain.ReservaStock
	err := r.db.WithContext(ctx).
		Where("pedido_id = ? AND estado = ?", pedidoID, domain.ReservaActiva).
		Find(&reservas).Error
	return reservas, err
}

func (r *RepositorioReservas) ActivasPorLote(ctx context.Context, loteID uuid.UUID) ([]domain.ReservaStock, error) {
	var reservas []domain.ReservaStock
	err := r.db.WithContext(ctx).
		Where("lote_id = ? AND estado = ?", loteID, domain.ReservaActiva).
		Find(&reservas).Error
	return reservas, err
}

// ActivasPorLotes suma, por lote, las unidades de sus reservas ACTIVA. Es lo
// que Reservar necesita para saber cuánto de cada lote ya está comprometido
// antes de asignar unidades nuevas.
func (r *RepositorioReservas) ActivasPorLotes(ctx context.Context, loteIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	activas := make(map[uuid.UUID]int, len(loteIDs))
	if len(loteIDs) == 0 {
		return activas, nil
	}

	var filas []struct {
		LoteID uuid.UUID
		Total  int
	}
	err := r.db.WithContext(ctx).
		Model(&domain.ReservaStock{}).
		Select("lote_id, SUM(unidades) AS total").
		Where("lote_id IN ? AND estado = ?", loteIDs, domain.ReservaActiva).
		Group("lote_id").
		Find(&filas).Error
	if err != nil {
		return nil, err
	}
	for _, fila := range filas {
		activas[fila.LoteID] = fila.Total
	}
	return activas, nil
}

func (r *RepositorioReservas) CambiarEstado(ctx context.Context, id uuid.UUID, nuevo domain.EstadoReserva) error {
	return r.db.WithContext(ctx).
		Model(&domain.ReservaStock{}).
		Where("id = ?", id).
		Update("estado", nuevo).Error
}
