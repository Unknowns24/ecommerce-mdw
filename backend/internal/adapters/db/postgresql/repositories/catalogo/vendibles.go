package catalogo

import (
	"context"
	"errors"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VarianteParaVenta struct {
	ID, ProductoID                              uuid.UUID
	Codigo, Nombre                              string
	PrecioMinoristaCentavos                     int64
	EstadoVariante, EstadoProducto, EstadoMarca string
}

type Vendibles struct{ db *gorm.DB }

func NuevoVendibles(db *gorm.DB) *Vendibles { return &Vendibles{db: db} }

func (r *Vendibles) PorID(ctx context.Context, id uuid.UUID) (VarianteParaVenta, error) {
	var v VarianteParaVenta
	err := r.db.WithContext(ctx).Table("variante v").
		Select("v.id, v.producto_id, v.codigo, v.nombre, v.precio_minorista_centavos, v.estado AS estado_variante, p.estado AS estado_producto, m.estado AS estado_marca").
		Joins("JOIN producto p ON p.id = v.producto_id").Joins("JOIN marca m ON m.id = p.marca_id").
		Where("v.id = ?", id).Take(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return v, apierr.ErrNoEncontrado
	}
	return v, err
}
