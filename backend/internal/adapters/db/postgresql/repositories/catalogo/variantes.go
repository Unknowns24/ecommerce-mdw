package catalogo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type Variantes struct{ db *gorm.DB }

func NuevoVariantes(db *gorm.DB) *Variantes { return &Variantes{db: db} }

const columnasVariante = "id, producto_id, codigo, nombre, descripcion, precio_minorista_centavos, precio_mayorista_centavos, estado, creado_en, actualizado_en"

func (r *Variantes) Crear(ctx context.Context, v *domain.Variante) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *Variantes) PorID(ctx context.Context, id uuid.UUID) (domain.Variante, error) {
	var v domain.Variante
	err := r.db.WithContext(ctx).Table("variante").Select(columnasVariante).Where("id = ?", id).Take(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return v, apierr.ErrNoEncontrado
	}
	return v, err
}

func (r *Variantes) PorCodigo(ctx context.Context, codigo string) (domain.Variante, error) {
	var v domain.Variante
	err := r.db.WithContext(ctx).Table("variante").Select(columnasVariante).Where("codigo = ?", codigo).Take(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return v, apierr.ErrNoEncontrado
	}
	return v, err
}

func (r *Variantes) Listar(ctx context.Context, limit, offset int) ([]domain.Variante, int64, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, 0, apierr.ErrValidacion{Campos: map[string]string{"paginacion": "fuera de rango"}}
	}
	var total int64
	if err := r.db.WithContext(ctx).Table("variante").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []domain.Variante
	err := r.db.WithContext(ctx).Table("variante").Select(columnasVariante).Order("creado_en DESC, id ASC").Limit(limit).Offset(offset).Scan(&out).Error
	if out == nil {
		out = []domain.Variante{}
	}
	return out, total, err
}

func (r *Variantes) Actualizar(ctx context.Context, id uuid.UUID, changes map[string]any) (domain.Variante, error) {
	result := r.db.WithContext(ctx).Table("variante").Where("id = ?", id).Updates(changes)
	if result.Error != nil {
		return domain.Variante{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Variante{}, apierr.ErrNoEncontrado
	}
	return r.PorID(ctx, id)
}

func (r *Variantes) CambiarEstado(ctx context.Context, id uuid.UUID, estado string) error {
	result := r.db.WithContext(ctx).Table("variante").Where("id = ?", id).Update("estado", estado)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apierr.ErrNoEncontrado
	}
	return nil
}

func (r *Variantes) ProductoActivo(ctx context.Context, id uuid.UUID) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("producto").Where("id = ? AND estado = ?", id, "ACTIVO").Count(&total).Error
	return total == 1, err
}
