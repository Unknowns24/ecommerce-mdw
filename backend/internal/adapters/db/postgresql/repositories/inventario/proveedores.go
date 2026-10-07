package inventario

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type RepositorioProveedores struct{ db *gorm.DB }

func NuevoRepositorioProveedores(db *gorm.DB) *RepositorioProveedores {
	return &RepositorioProveedores{db: db}
}

func (r *RepositorioProveedores) Crear(ctx context.Context, proveedor *domain.Proveedor) error {
	return r.db.WithContext(ctx).Create(proveedor).Error
}

func (r *RepositorioProveedores) Listar(ctx context.Context, limit, offset int) ([]domain.Proveedor, error) {
	var proveedores []domain.Proveedor
	err := r.db.WithContext(ctx).
		Select("id", "nombre", "correo", "telefono", "web", "creado_en").
		Order("nombre ASC").
		Limit(limit).
		Offset(offset).
		Find(&proveedores).Error
	return proveedores, err
}

func (r *RepositorioProveedores) PorID(ctx context.Context, id uuid.UUID) (domain.Proveedor, error) {
	var proveedor domain.Proveedor
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&proveedor).Error
	return proveedor, err
}
