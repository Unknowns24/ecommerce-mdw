package catalogo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type Productos struct{ db *gorm.DB }

func NuevoProductos(db *gorm.DB) *Productos { return &Productos{db: db} }

const columnasProducto = "id, marca_id, nombre, descripcion, estado, metodologia_rotacion, creado_en, actualizado_en"

func (r *Productos) Crear(ctx context.Context, p *domain.Producto, categorias []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(p).Error; err != nil {
			return err
		}
		return reemplazarCategorias(tx, p.ID, categorias)
	})
}

func (r *Productos) PorID(ctx context.Context, id uuid.UUID) (domain.Producto, error) {
	var p domain.Producto
	err := r.db.WithContext(ctx).Table("producto").Select(columnasProducto).Where("id = ?", id).Take(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return p, apierr.ErrNoEncontrado
	}
	return p, err
}

func (r *Productos) Listar(ctx context.Context, limit, offset int) ([]domain.Producto, int64, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, 0, apierr.ErrValidacion{Campos: map[string]string{"paginacion": "fuera de rango"}}
	}
	var total int64
	if err := r.db.WithContext(ctx).Table("producto").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	out := make([]domain.Producto, 0)
	err := r.db.WithContext(ctx).Table("producto").Select(columnasProducto).Order("creado_en DESC, id ASC").Limit(limit).Offset(offset).Scan(&out).Error
	return out, total, err
}

func (r *Productos) Actualizar(ctx context.Context, id uuid.UUID, changes map[string]any, categorias *[]uuid.UUID) (domain.Producto, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exists struct{ ID uuid.UUID }
		if err := tx.Table("producto").Select("id").Where("id = ?", id).Take(&exists).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apierr.ErrNoEncontrado
			}
			return err
		}
		if len(changes) > 0 {
			if err := tx.Table("producto").Where("id = ?", id).Updates(changes).Error; err != nil {
				return err
			}
		}
		if categorias != nil {
			return reemplazarCategorias(tx, id, *categorias)
		}
		return nil
	})
	if err != nil {
		return domain.Producto{}, err
	}
	return r.PorID(ctx, id)
}

func (r *Productos) CambiarEstado(ctx context.Context, id uuid.UUID, estado string) error {
	result := r.db.WithContext(ctx).Table("producto").Where("id = ?", id).Update("estado", estado)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apierr.ErrNoEncontrado
	}
	return nil
}

func (r *Productos) ReemplazarCategorias(ctx context.Context, id uuid.UUID, categorias []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return reemplazarCategorias(tx, id, categorias) })
}

func reemplazarCategorias(tx *gorm.DB, id uuid.UUID, categorias []uuid.UUID) error {
	if err := tx.Where("producto_id = ?", id).Delete(&domain.ProductoCategoria{}).Error; err != nil {
		return err
	}
	seen := map[uuid.UUID]bool{}
	for _, categoriaID := range categorias {
		if seen[categoriaID] {
			continue
		}
		seen[categoriaID] = true
		if err := tx.Create(&domain.ProductoCategoria{ProductoID: id, CategoriaID: categoriaID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Productos) Categorias(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	var rows []struct{ CategoriaID uuid.UUID }
	err := r.db.WithContext(ctx).Table("producto_categoria").Select("categoria_id").Where("producto_id = ?", id).Limit(100).Scan(&rows).Error
	out := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.CategoriaID)
	}
	return out, err
}

func (r *Productos) Variantes(ctx context.Context, id uuid.UUID) ([]domain.Variante, error) {
	var out []domain.Variante
	err := r.db.WithContext(ctx).Table("variante").Select("id, producto_id, codigo, nombre, descripcion, precio_minorista_centavos, precio_mayorista_centavos, estado, creado_en, actualizado_en").Where("producto_id = ?", id).Order("nombre ASC, id ASC").Limit(100).Scan(&out).Error
	return out, err
}

func (r *Productos) MarcaActiva(ctx context.Context, id uuid.UUID) (bool, error) {
	var total int64
	err := r.db.WithContext(ctx).Table("marca").Where("id = ? AND estado = ?", id, "ACTIVA").Count(&total).Error
	return total == 1, err
}

func (r *Productos) CategoriasActivas(ctx context.Context, ids []uuid.UUID) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	var total int64
	err := r.db.WithContext(ctx).Table("categoria").Where("id IN ? AND estado = ?", ids, "ACTIVA").Count(&total).Error
	return total == int64(len(ids)), err
}
