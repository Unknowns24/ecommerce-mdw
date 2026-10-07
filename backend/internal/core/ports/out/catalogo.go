// Package out declara lo que el núcleo necesita del mundo exterior, sin decir
// cómo se resuelve. Los adaptadores (base de datos, catálogo de otro módulo)
// lo implementan.
package out

import (
	"context"

	"github.com/google/uuid"
)

// VarianteVendible es lo que el pedido necesita saber de una variante para
// valorizarla. Replica el contrato congelado del catálogo de Genaro
// (usecases/catalogo.VarianteVendible): los precios salen de acá, nunca del
// body del request.
type VarianteVendible struct {
	ID                      uuid.UUID
	ProductoID              uuid.UUID
	Codigo                  string
	Nombre                  string
	PrecioMinoristaCentavos int64
	Activa                  bool
}

// LectorCatalogo devuelve la variante o apierr.ErrNoEncontrado si no existe
// o está inactiva. Lo implementa el Lector de Genaro; mientras no esté en
// develop, los tests usan un stub en memoria.
type LectorCatalogo interface {
	VariantePorID(ctx context.Context, id uuid.UUID) (VarianteVendible, error)
}
