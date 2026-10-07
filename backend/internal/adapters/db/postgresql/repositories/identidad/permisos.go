package identidad

import (
	"context"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/google/uuid"
)

func (r *Repositorio) ListarPermisos(ctx context.Context, limit, offset int) ([]domain.Permiso, error) {
	var x []domain.Permiso
	err := r.DB.WithContext(ctx).Select("id", "accion", "descripcion").Order("accion").Limit(limit).Offset(offset).Find(&x).Error
	return x, err
}
func (r *Repositorio) PermisosDeUsuario(ctx context.Context, id uuid.UUID) ([]string, error) {
	var x []string
	err := r.DB.WithContext(ctx).Table("permiso p").Select("DISTINCT p.accion").Joins("JOIN rol_permiso rp ON rp.permiso_id = p.id").Joins("JOIN usuario_rol ur ON ur.rol_id = rp.rol_id").Where("ur.usuario_id = ?", id).Order("p.accion").Pluck("p.accion", &x).Error
	return x, err
}
func (r *Repositorio) PermisosDeRol(ctx context.Context, id uuid.UUID) ([]string, error) {
	var x []string
	err := r.DB.WithContext(ctx).Table("permiso p").Select("p.accion").Joins("JOIN rol_permiso rp ON rp.permiso_id = p.id").Where("rp.rol_id = ?", id).Order("p.accion").Pluck("p.accion", &x).Error
	return x, err
}
