package catalogo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type Imagenes struct{ db *gorm.DB }

func NuevoImagenes(db *gorm.DB) *Imagenes { return &Imagenes{db: db} }

func (r *Imagenes) Crear(ctx context.Context, imagen *domain.ImagenVariante) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := bloquearVariante(tx, imagen.VarianteID); err != nil {
			return err
		}
		var count int64
		if err := tx.Table("imagen_variante").Where("variante_id = ?", imagen.VarianteID).Count(&count).Error; err != nil {
			return err
		}
		imagen.EsPrincipal = count == 0
		return tx.Create(imagen).Error
	})
}

func (r *Imagenes) PorID(ctx context.Context, id uuid.UUID) (domain.ImagenVariante, error) {
	var imagen domain.ImagenVariante
	err := r.db.WithContext(ctx).Table("imagen_variante").Select("id, variante_id, referencia_imagen, es_principal, orden").Where("id = ?", id).Take(&imagen).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return imagen, apierr.ErrNoEncontrado
	}
	return imagen, err
}

func (r *Imagenes) Eliminar(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var objetivo domain.ImagenVariante
		if err := tx.Table("imagen_variante").Select("variante_id").Where("id = ?", id).Take(&objetivo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apierr.ErrNoEncontrado
			}
			return err
		}
		if err := bloquearVariante(tx, objetivo.VarianteID); err != nil {
			return err
		}
		var imagen domain.ImagenVariante
		if err := tx.Table("imagen_variante").Select("id, variante_id, es_principal").Where("id = ?", id).Take(&imagen).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apierr.ErrNoEncontrado
			}
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&domain.ImagenVariante{}).Error; err != nil {
			return err
		}
		if imagen.EsPrincipal {
			var siguiente domain.ImagenVariante
			err := tx.Table("imagen_variante").Select("id").Where("variante_id = ?", imagen.VarianteID).Order("orden ASC, id ASC").Take(&siguiente).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
			return tx.Table("imagen_variante").Where("id = ?", siguiente.ID).Update("es_principal", true).Error
		}
		return nil
	})
}

// DefinirPrincipal clears the old flag and sets the selected image in one transaction.
func (r *Imagenes) DefinirPrincipal(ctx context.Context, varianteID, imagenID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := bloquearVariante(tx, varianteID); err != nil {
			return err
		}
		var imagen domain.ImagenVariante
		if err := tx.Table("imagen_variante").Select("id, variante_id").Where("id = ? AND variante_id = ?", imagenID, varianteID).Take(&imagen).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apierr.ErrNoEncontrado
			}
			return err
		}
		if err := tx.Table("imagen_variante").Where("variante_id = ? AND es_principal = true", varianteID).Update("es_principal", false).Error; err != nil {
			return err
		}
		return tx.Table("imagen_variante").Where("id = ?", imagenID).Update("es_principal", true).Error
	})
}

func bloquearVariante(tx *gorm.DB, id uuid.UUID) error {
	var row struct{ ID uuid.UUID }
	err := tx.Table("variante").Select("id").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apierr.ErrNoEncontrado
	}
	return err
}
