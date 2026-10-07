package identidad

import (
	"context"
	"errors"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrYaInicializado = errors.New("el sistema ya fue inicializado")

type Repositorio struct{ DB *gorm.DB }

func Nuevo(db *gorm.DB) *Repositorio                  { return &Repositorio{DB: db} }
func (r *Repositorio) ConTx(tx *gorm.DB) *Repositorio { return &Repositorio{DB: tx} }

func (r *Repositorio) Crear(ctx context.Context, u *domain.Usuario, rol string) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		return tx.Exec(`INSERT INTO usuario_rol (usuario_id,rol_id) SELECT ?,id FROM rol WHERE nombre = ?`, u.ID, rol).Error
	})
}
func (r *Repositorio) PorCorreo(ctx context.Context, correo string) (domain.Usuario, error) {
	var u domain.Usuario
	err := r.DB.WithContext(ctx).Where("lower(correo) = lower(?)", correo).First(&u).Error
	return u, err
}
func (r *Repositorio) PorID(ctx context.Context, id uuid.UUID) (domain.Usuario, error) {
	var u domain.Usuario
	err := r.DB.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}
func (r *Repositorio) Listar(ctx context.Context, limit, offset int, estado string) ([]domain.Usuario, error) {
	var items []domain.Usuario
	q := r.DB.WithContext(ctx).Model(&domain.Usuario{}).Select("id", "nombre", "apellido", "correo", "estado", "origen", "es_dueno_inicial", "creado_en", "actualizado_en")
	if estado != "" {
		q = q.Where("estado = ?", estado)
	}
	err := q.Order("creado_en DESC").Limit(limit).Offset(offset).Find(&items).Error
	return items, err
}
func (r *Repositorio) ActualizarDatos(ctx context.Context, id uuid.UUID, nombre, apellido string, telefono *string) error {
	res := r.DB.WithContext(ctx).Model(&domain.Usuario{}).Where("id = ?", id).Updates(map[string]any{"nombre": nombre, "apellido": apellido, "telefono": telefono})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *Repositorio) CambiarEstado(ctx context.Context, id uuid.UUID, estado string) error {
	res := r.DB.WithContext(ctx).Model(&domain.Usuario{}).Where("id = ?", id).Update("estado", estado)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *Repositorio) ExisteDueno(ctx context.Context) (bool, error) {
	var n int64
	err := r.DB.WithContext(ctx).Table("usuario_rol ur").Joins("JOIN rol r ON r.id = ur.rol_id").Where("r.nombre = ?", "dueno").Count(&n).Error
	return n > 0, err
}
func EsNoEncontrado(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

func (r *Repositorio) Bootstrap(ctx context.Context, u *domain.Usuario) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Este bloqueo serializa las solicitudes de bootstrap incluso con distintos correos.
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(2026100401)`).Error; err != nil {
			return err
		}
		var n int64
		if err := tx.Table("usuario_rol ur").Joins("JOIN rol r ON r.id = ur.rol_id").Where("r.nombre = ?", "dueno").Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return ErrYaInicializado
		}
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		res := tx.Exec(`INSERT INTO usuario_rol (usuario_id,rol_id) SELECT ?,id FROM rol WHERE nombre = ?`, u.ID, "dueno")
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
