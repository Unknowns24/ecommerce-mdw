package pedidos

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

func variante(precio int64, activa bool) out.VarianteVendible {
	id := uuid.New()
	return out.VarianteVendible{ID: id, Codigo: "COD-" + id.String()[:4], Nombre: "Perfume", PrecioMinoristaCentavos: precio, Activa: activa}
}

func TestCalcularLineas_UsaLosPreciosDelServidor(t *testing.T) {
	a, b := variante(1500050, true), variante(299900, true)
	precios := map[uuid.UUID]out.VarianteVendible{a.ID: a, b.ID: b}

	lineas, err := CalcularLineas([]ItemPedido{{a.ID, 2}, {b.ID, 1}}, precios)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(lineas) != 2 {
		t.Fatalf("esperaba 2 líneas, hubo %d", len(lineas))
	}
	if lineas[0].PrecioUnitarioCentavos != 1500050 || lineas[0].SubtotalCentavos != 3000100 {
		t.Errorf("línea 0 mal valorizada: %+v", lineas[0])
	}
	if lineas[0].Codigo != a.Codigo || lineas[0].Nombre != a.Nombre {
		t.Errorf("la copia histórica de código/nombre no coincide: %+v", lineas[0])
	}
}

func TestCalcularLineas_UnaSolaUnidad(t *testing.T) { // borde
	a := variante(100, true)
	lineas, err := CalcularLineas([]ItemPedido{{a.ID, 1}}, map[uuid.UUID]out.VarianteVendible{a.ID: a})
	if err != nil || lineas[0].SubtotalCentavos != 100 {
		t.Fatalf("una unidad debe costar el precio unitario: %+v, %v", lineas, err)
	}
}

func TestCalcularLineas_SumaVariantesRepetidas(t *testing.T) {
	a := variante(250, true)
	lineas, err := CalcularLineas([]ItemPedido{{a.ID, 2}, {a.ID, 3}}, map[uuid.UUID]out.VarianteVendible{a.ID: a})
	if err != nil {
		t.Fatal(err)
	}
	if len(lineas) != 1 || lineas[0].Unidades != 5 || lineas[0].SubtotalCentavos != 1250 {
		t.Fatalf("debía haber una línea de 5 unidades por 1250: %+v", lineas)
	}
}

func TestCalcularLineas_Errores(t *testing.T) {
	activa, inactiva := variante(100, true), variante(100, false)
	precios := map[uuid.UUID]out.VarianteVendible{activa.ID: activa, inactiva.ID: inactiva}

	casos := []struct {
		nombre   string
		items    []ItemPedido
		esperado func(error) bool
	}{
		{"carrito vacío", nil, esValidacion},
		{"cero unidades", []ItemPedido{{activa.ID, 0}}, esValidacion},
		{"unidades negativas", []ItemPedido{{activa.ID, -1}}, esValidacion},
		{"variante inexistente", []ItemPedido{{uuid.New(), 1}}, esNoEncontrado},
		{"variante inactiva", []ItemPedido{{inactiva.ID, 1}}, esNoEncontrado},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if _, err := CalcularLineas(c.items, precios); !c.esperado(err) {
				t.Fatalf("error inesperado: %v", err)
			}
		})
	}
}

func TestCalcularTotal(t *testing.T) {
	lineas := []LineaCalculada{{SubtotalCentavos: 3000100}, {SubtotalCentavos: 299900}}

	if sub, tot := CalcularTotal(lineas, 50000); sub != 3300000 || tot != 3350000 {
		t.Errorf("con envío: subtotal=%d total=%d", sub, tot)
	}
	if sub, tot := CalcularTotal(lineas, 0); sub != tot { // retiro: sin cargo
		t.Errorf("sin envío el total debe igualar al subtotal: %d vs %d", sub, tot)
	}
	if sub, tot := CalcularTotal(nil, 0); sub != 0 || tot != 0 { // borde
		t.Errorf("sin líneas todo es cero: %d %d", sub, tot)
	}
}

func esValidacion(err error) bool {
	var e apierr.ErrValidacion
	return errors.As(err, &e)
}

func esNoEncontrado(err error) bool { return errors.Is(err, apierr.ErrNoEncontrado) }

func esRegla(err error) bool {
	var e apierr.ErrRegla
	return errors.As(err, &e)
}
