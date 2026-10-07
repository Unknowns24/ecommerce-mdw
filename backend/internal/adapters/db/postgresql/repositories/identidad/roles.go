package identidad

import (
	"context"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *Repositorio) CrearRol(ctx context.Context, rol *domain.Rol, permisos []string) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rol).Error; err != nil {
			return err
		}
		for _, p := range permisos {
			if err := tx.Exec(`INSERT INTO rol_permiso (rol_id,permiso_id) SELECT ?,id FROM permiso WHERE accion = ?`, rol.ID, p).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *Repositorio) ListarRoles(ctx context.Context, limit, offset int) ([]domain.Rol, error) {
	var x []domain.Rol
	err := r.DB.WithContext(ctx).Select("id", "nombre", "creado_en").Order("nombre").Limit(limit).Offset(offset).Find(&x).Error
	return x, err
}
func (r *Repositorio) PorNombre(ctx context.Context, nombre string) (domain.Rol, error) {
	var x domain.Rol
	err := r.DB.WithContext(ctx).Where("nombre = ?", nombre).First(&x).Error
	return x, err
}
func (r *Repositorio) AsignarAUsuario(ctx context.Context, usuarioID, rolID uuid.UUID) error {
	return r.DB.WithContext(ctx).Create(&domain.UsuarioRol{UsuarioID: usuarioID, RolID: rolID}).Error
}
func (r *Repositorio) QuitarDeUsuario(ctx context.Context, usuarioID, rolID uuid.UUID) error {
	return r.DB.WithContext(ctx).Where("usuario_id = ? AND rol_id = ?", usuarioID, rolID).Delete(&domain.UsuarioRol{}).Error
}
func (r *Repositorio) RolesDeUsuario(ctx context.Context, id uuid.UUID) ([]string, error) {
	var nombres []string
	err := r.DB.WithContext(ctx).Table("rol r").Select("r.nombre").Joins("JOIN usuario_rol ur ON ur.rol_id = r.id").Where("ur.usuario_id = ?", id).Order("r.nombre").Pluck("r.nombre", &nombres).Error
	return nombres, err
}
func (r *Repositorio) ReemplazarRoles(ctx context.Context, id uuid.UUID, roles []string) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("usuario_id = ?", id).Delete(&domain.UsuarioRol{}).Error; err != nil {
			return err
		}
		for _, nombre := range roles {
			res := tx.Exec(`INSERT INTO usuario_rol (usuario_id,rol_id) SELECT ?,id FROM rol WHERE nombre = ?`, id, nombre)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
		}
		return nil
	})
}
