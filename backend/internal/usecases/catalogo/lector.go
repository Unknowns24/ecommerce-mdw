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
	err := l.db.WithContext(ctx).Table("variante").Select("id, producto_id, codigo, nombre, precio_minorista_centavos, estado").Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return VarianteVendible{}, apierr.ErrNoEncontrado
	}
	if err != nil {
		return VarianteVendible{}, err
	}
	if !esVendible(row.Estado) {
		return VarianteVendible{}, apierr.ErrNoEncontrado
	}
	return VarianteVendible{ID: row.ID, ProductoID: row.ProductoID, Codigo: row.Codigo, Nombre: row.Nombre, PrecioMinoristaCentavos: row.PrecioMinoristaCentavos, Activa: true}, nil
}

const (
	estadoVarianteActiva   = "ACTIVA"
	estadoVarianteInactiva = "INACTIVA"
)

// esVendible aplica la regla de la spec (sección 6): una variante inactiva no
// se ofrece aunque conserve existencias. Decide sólo por el estado: el stock
// lo resuelve inventario, y una variante activa sin stock sigue siendo del
// catálogo (se muestra «SIN STOCK»). Desactivar no equivale a agotar.
func esVendible(estado string) bool { return estado == estadoVarianteActiva }
