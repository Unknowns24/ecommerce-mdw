package inventario

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type RepositorioMovimientos struct{ db *gorm.DB }

func NuevoRepositorioMovimientos(db *gorm.DB) *RepositorioMovimientos {
	return &RepositorioMovimientos{db: db}
}

func (r *RepositorioMovimientos) Crear(ctx context.Context, movimiento *domain.MovimientoStock) error {
	return r.db.WithContext(ctx).Create(movimiento).Error
}

// PorVariante pagina el historial de movimientos de una variante, más
// recientes primero.
func (r *RepositorioMovimientos) PorVariante(ctx context.Context, varianteID uuid.UUID, limit, offset int) ([]domain.MovimientoStock, error) {
	var movimientos []domain.MovimientoStock
	err := r.db.WithContext(ctx).
		Joins("JOIN lote ON lote.id = movimiento_stock.lote_id").
		Where("lote.variante_id = ?", varianteID).
		Order("movimiento_stock.creado_en DESC").
		Limit(limit).
		Offset(offset).
		Find(&movimientos).Error
	return movimientos, err
}

// NetoPorLotes agrega, en una sola consulta, cuánto movieron los ajustes y
// salidas de cada lote desde su ingreso: AJUSTE suma su delta con signo,
// SALIDA resta. El INGRESO no se vuelve a sumar acá porque ya está contado
// en lote.unidades_ingresadas; el movimiento INGRESO que lo acompaña es
// sólo auditoría.
func (r *RepositorioMovimientos) NetoPorLotes(ctx context.Context, loteIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	neto := make(map[uuid.UUID]int, len(loteIDs))
	if len(loteIDs) == 0 {
		return neto, nil
	}

	var filas []struct {
		LoteID uuid.UUID
		Neto   int
	}
	err := r.db.WithContext(ctx).
		Model(&domain.MovimientoStock{}).
		Select("lote_id, SUM(CASE WHEN tipo = 'AJUSTE' THEN unidades WHEN tipo = 'SALIDA' THEN -unidades ELSE 0 END) AS neto").
		Where("lote_id IN ?", loteIDs).
		Group("lote_id").
		Find(&filas).Error
	if err != nil {
		return nil, err
	}
	for _, fila := range filas {
		neto[fila.LoteID] = fila.Neto
	}
	return neto, nil
}
