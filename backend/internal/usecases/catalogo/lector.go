package catalogo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

// VarianteVendible es el contrato del checkout y del inventario.
type VarianteVendible struct {
	ID                      uuid.UUID
	ProductoID              uuid.UUID
	Codigo                  string
	Nombre                  string
	PrecioMinoristaCentavos int64
	Activa                  bool
}

type Lector struct{ db *gorm.DB }

func NuevoLector(db *gorm.DB) *Lector { return &Lector{db: db} }

// VariantePorID siempre obtiene el precio del servidor. Las variantes inactivas no son vendibles.
func (l *Lector) VariantePorID(ctx context.Context, id uuid.UUID) (VarianteVendible, error) {
	var row struct {
		ID                      uuid.UUID
		ProductoID              uuid.UUID
		Codigo                  string
		Nombre                  string
		PrecioMinoristaCentavos int64
		Estado                  string
	}
	err := l.db.WithContext(ctx).Table("variante").Select("id, producto_id, codigo, nombre, precio_minorista_centavos, estado").Where("id = ? AND estado = ?", id, "ACTIVA").Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return VarianteVendible{}, apierr.ErrNoEncontrado
	}
	if err != nil {
		return VarianteVendible{}, err
	}
	return VarianteVendible{ID: row.ID, ProductoID: row.ProductoID, Codigo: row.Codigo, Nombre: row.Nombre, PrecioMinoristaCentavos: row.PrecioMinoristaCentavos, Activa: true}, nil
}
