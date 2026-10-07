// Package inventario holds the inventory module's PostgreSQL repositories.
// A repository constructed with a transaction (*gorm.DB from db.Transaction)
// runs inside it; constructed with the plain db handle, it runs standalone.
package inventario

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type RepositorioLotes struct{ db *gorm.DB }

func NuevoRepositorioLotes(db *gorm.DB) *RepositorioLotes { return &RepositorioLotes{db: db} }

func (r *RepositorioLotes) Crear(ctx context.Context, lote *domain.Lote) error {
	return r.db.WithContext(ctx).Create(lote).Error
}

func (r *RepositorioLotes) PorID(ctx context.Context, id uuid.UUID) (domain.Lote, error) {
	var lote domain.Lote
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&lote).Error
	return lote, err
}

// PorIDBloqueando lee un lote con SELECT ... FOR UPDATE, para usar dentro de
// la transacción de un ajuste de stock.
func (r *RepositorioLotes) PorIDBloqueando(ctx context.Context, id uuid.UUID) (domain.Lote, error) {
	var lote domain.Lote
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&lote).Error
	return lote, err
}

// PorVariante devuelve los lotes de una variante ordenados por fecha_ingreso:
// ascendente para FIFO, descendente para LIFO.
func (r *RepositorioLotes) PorVariante(ctx context.Context, varianteID uuid.UUID, metodologia string) ([]domain.Lote, error) {
	var lotes []domain.Lote
	err := r.db.WithContext(ctx).
		Where("variante_id = ?", varianteID).
		Order(ordenPorMetodologia(metodologia)).
		Find(&lotes).Error
	return lotes, err
}

// PorVarianteBloqueando es la misma consulta que PorVariante pero con
// SELECT ... FOR UPDATE, para usar dentro de la transacción del checkout:
// evita que dos reservas simultáneas lean la misma última unidad disponible.
func (r *RepositorioLotes) PorVarianteBloqueando(ctx context.Context, varianteID uuid.UUID, metodologia string) ([]domain.Lote, error) {
	var lotes []domain.Lote
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("variante_id = ?", varianteID).
		Order(ordenPorMetodologia(metodologia)).
		Find(&lotes).Error
	return lotes, err
}

func ordenPorMetodologia(metodologia string) string {
	if metodologia == "LIFO" {
		return "fecha_ingreso DESC"
	}
	return "fecha_ingreso ASC"
}
