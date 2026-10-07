package catalogo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type MarcasCategorias struct{ db *gorm.DB }

func NuevoMarcasCategorias(db *gorm.DB) *MarcasCategorias { return &MarcasCategorias{db: db} }

func (r *MarcasCategorias) CrearMarca(ctx context.Context, m *domain.Marca) error {
	return r.db.WithContext(ctx).Create(m).Error
}
func (r *MarcasCategorias) Marca(ctx context.Context, id uuid.UUID) (domain.Marca, error) {
	var m domain.Marca
	err := r.db.WithContext(ctx).Table("marca").Select("id, nombre, estado, creado_en").Where("id = ?", id).Take(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m, apierr.ErrNoEncontrado
	}
	return m, err
}
func (r *MarcasCategorias) ListarMarcas(ctx context.Context, limit, offset int) ([]domain.Marca, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Table("marca").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []domain.Marca
	err := r.db.WithContext(ctx).Table("marca").Select("id, nombre, estado, creado_en").Order("nombre ASC, id ASC").Limit(limit).Offset(offset).Scan(&out).Error
	if out == nil {
		out = []domain.Marca{}
	}
	return out, total, err
}
func (r *MarcasCategorias) EditarMarca(ctx context.Context, id uuid.UUID, changes map[string]any) (domain.Marca, error) {
	result := r.db.WithContext(ctx).Table("marca").Where("id = ?", id).Updates(changes)
	if result.Error != nil {
		return domain.Marca{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Marca{}, apierr.ErrNoEncontrado
	}
	return r.Marca(ctx, id)
}
func (r *MarcasCategorias) DesactivarMarca(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Table("marca").Where("id = ?", id).Update("estado", "INACTIVA")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apierr.ErrNoEncontrado
	}
	return nil
}

func (r *MarcasCategorias) CrearCategoria(ctx context.Context, c *domain.Categoria) error {
	return r.db.WithContext(ctx).Create(c).Error
}
func (r *MarcasCategorias) Categoria(ctx context.Context, id uuid.UUID) (domain.Categoria, error) {
	var c domain.Categoria
	err := r.db.WithContext(ctx).Table("categoria").Select("id, nombre, padre_id, estado, creado_en").Where("id = ?", id).Take(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c, apierr.ErrNoEncontrado
	}
	return c, err
}
func (r *MarcasCategorias) ListarCategorias(ctx context.Context, limit, offset int) ([]domain.Categoria, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Table("categoria").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []domain.Categoria
	err := r.db.WithContext(ctx).Table("categoria").Select("id, nombre, padre_id, estado, creado_en").Order("nombre ASC, id ASC").Limit(limit).Offset(offset).Scan(&out).Error
	if out == nil {
		out = []domain.Categoria{}
	}
	return out, total, err
}
func (r *MarcasCategorias) EditarCategoria(ctx context.Context, id uuid.UUID, changes map[string]any) (domain.Categoria, error) {
	result := r.db.WithContext(ctx).Table("categoria").Where("id = ?", id).Updates(changes)
	if result.Error != nil {
		return domain.Categoria{}, result.Error
	}
	if result.RowsAffected == 0 {
		return domain.Categoria{}, apierr.ErrNoEncontrado
	}
	return r.Categoria(ctx, id)
}
func (r *MarcasCategorias) DesactivarCategoria(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Table("categoria").Where("id = ?", id).Update("estado", "INACTIVA")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apierr.ErrNoEncontrado
	}
	return nil
}

func (r *MarcasCategorias) EsDescendiente(ctx context.Context, categoriaID, posiblePadre uuid.UUID) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Raw(`WITH RECURSIVE d AS (SELECT id FROM categoria WHERE id = ? UNION ALL SELECT c.id FROM categoria c JOIN d ON c.padre_id = d.id) SELECT count(*) FROM d WHERE id = ?`, categoriaID, posiblePadre).Scan(&total).Error
	return total > 0, err
}
