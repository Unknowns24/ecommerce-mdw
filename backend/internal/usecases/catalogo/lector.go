package catalogo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/catalogo"
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

type Lector struct{ vendibles *repo.Vendibles }

func NuevoLector(db *gorm.DB) *Lector { return &Lector{vendibles: repo.NuevoVendibles(db)} }

// VariantePorID siempre obtiene el precio del servidor. Las variantes inactivas no son vendibles.
func (l *Lector) VariantePorID(ctx context.Context, id uuid.UUID) (VarianteVendible, error) {
	row, err := l.vendibles.PorID(ctx, id)
	if err != nil {
		return VarianteVendible{}, err
	}
	if !puedeVenderse(row.EstadoVariante, row.EstadoProducto, row.EstadoMarca) {
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

func puedeVenderse(variante, producto, marca string) bool {
	return esVendible(variante) && producto == "ACTIVO" && marca == "ACTIVA"
}
