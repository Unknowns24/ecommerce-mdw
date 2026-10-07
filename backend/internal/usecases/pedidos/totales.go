// Package pedidos contiene las reglas del carrito, el checkout y el ciclo de
// vida del pedido. Estos archivos son reglas puras: no importan GORM ni
// net/http y nunca llaman a time.Now() (todo entra por parámetro), así se
// pueden testear sin base, sin servidor y sin esperar a que pase el tiempo.
package pedidos

import (
	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// ItemPedido es lo único que el comprador decide de cada línea: qué variante
// y cuántas unidades. El precio nunca viene de acá.
type ItemPedido struct {
	VarianteID uuid.UUID
	Unidades   int
}

// LineaCalculada es una línea ya valorizada con los precios del servidor.
type LineaCalculada struct {
	VarianteID             uuid.UUID
	Codigo, Nombre         string
	Unidades               int
	PrecioUnitarioCentavos int64
	SubtotalCentavos       int64
}

// CalcularLineas valoriza los items con los precios que devolvió el catálogo.
// Si una variante se repite, suma sus unidades en una sola línea. Devuelve
// ErrValidacion si el pedido está vacío o las unidades no son positivas, y
// ErrNoEncontrado si una variante no existe o está inactiva.
func CalcularLineas(items []ItemPedido, precios map[uuid.UUID]out.VarianteVendible) ([]LineaCalculada, error) {
	if len(items) == 0 {
		return nil, apierr.ErrValidacion{Campos: map[string]string{"items": "Agregá al menos un producto"}}
	}

	posicion := make(map[uuid.UUID]int, len(items))
	lineas := make([]LineaCalculada, 0, len(items))

	for _, it := range items {
		if it.Unidades <= 0 {
			return nil, apierr.ErrValidacion{Campos: map[string]string{"unidades": "Las unidades deben ser un entero positivo"}}
		}
		v, ok := precios[it.VarianteID]
		if !ok || !v.Activa {
			return nil, apierr.ErrNoEncontrado
		}

		if i, repetida := posicion[it.VarianteID]; repetida {
			lineas[i].Unidades += it.Unidades
			lineas[i].SubtotalCentavos = int64(lineas[i].Unidades) * lineas[i].PrecioUnitarioCentavos
			continue
		}

		posicion[it.VarianteID] = len(lineas)
		lineas = append(lineas, LineaCalculada{
			VarianteID:             v.ID,
			Codigo:                 v.Codigo,
			Nombre:                 v.Nombre,
			Unidades:               it.Unidades,
			PrecioUnitarioCentavos: v.PrecioMinoristaCentavos,
			SubtotalCentavos:       int64(it.Unidades) * v.PrecioMinoristaCentavos,
		})
	}
	return lineas, nil
}

// CalcularTotal suma las líneas y le agrega el envío. Todo en enteros de
// centavos: sin promociones ni cupones (quedan fuera del Parcial I).
func CalcularTotal(lineas []LineaCalculada, envioCentavos int64) (subtotal, total int64) {
	for _, l := range lineas {
		subtotal += l.SubtotalCentavos
	}
	return subtotal, subtotal + envioCentavos
}
