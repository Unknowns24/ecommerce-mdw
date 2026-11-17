package catalogo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

// arbol es un catálogo de categorías en memoria: cada categoría guarda un solo
// padre (spec, sección 3).
type arbol map[uuid.UUID]domain.Categoria

func (a arbol) agregar(padre *uuid.UUID) uuid.UUID {
	id := uuid.New()
	a[id] = domain.Categoria{ID: id, Nombre: id.String(), PadreID: padre, Estado: "ACTIVA"}
	return id
}

func (a arbol) leer(id uuid.UUID) (domain.Categoria, error) {
	c, ok := a[id]
	if !ok {
		return domain.Categoria{}, apierr.ErrNoEncontrado
	}
	return c, nil
}

// esDescendiente sube desde posiblePadre por la cadena de padres: si llega a
// categoria, posiblePadre está debajo de ella (o es ella misma).
func (a arbol) esDescendiente(categoria, posiblePadre uuid.UUID) (bool, error) {
	actual := &posiblePadre
	for actual != nil {
		if *actual == categoria {
			return true, nil
		}
		actual = a[*actual].PadreID
	}
	return false, nil
}

func campoPadreID(t *testing.T, err error) string {
	t.Helper()
	var v apierr.ErrValidacion
	if !errors.As(err, &v) {
		t.Fatalf("se esperaba ErrValidacion en padreId, llegó %v", err)
	}
	msg, ok := v.Campos["padreId"]
	if !ok {
		t.Fatalf("ErrValidacion sin el campo padreId: %v", v.Campos)
	}
	return msg
}

func TestCategoriaConUnSoloPadreYSinCiclos(t *testing.T) {
	// raiz ── hija ── nieta
	// otra
	a := arbol{}
	raiz := a.agregar(nil)
	hija := a.agregar(&raiz)
	nieta := a.agregar(&hija)
	otra := a.agregar(nil)

	casos := []struct {
		nombre  string
		id      uuid.UUID // uuid.Nil al crear
		padreID *uuid.UUID
		rechaza bool
	}{
		{"crear sin padre es una raíz válida", uuid.Nil, nil, false},
		{"crear debajo de una categoría existente", uuid.Nil, &hija, false},
		{"crear debajo de una categoría que no existe", uuid.Nil, ptr(uuid.New()), true},
		{"una categoría no puede ser su propio padre", raiz, &raiz, true},
		{"mover una categoría debajo de su hija crea un ciclo", raiz, &hija, true},
		{"mover una categoría debajo de su nieta crea un ciclo", raiz, &nieta, true},
		{"mover una categoría a otra rama es válido", hija, &otra, false},
		{"mover una hoja debajo de la raíz es válido", nieta, &raiz, false},
		{"quitarle el padre la vuelve raíz", nieta, nil, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := validarPadreCategoria(c.id, c.padreID, a.leer, a.esDescendiente)
			if c.rechaza {
				campoPadreID(t, err)
				return
			}
			if err != nil {
				t.Fatalf("se esperaba aceptar el padre, llegó %v", err)
			}
		})
	}
}

func TestCategoriaRechazaElCicloConElMensaje(t *testing.T) {
	a := arbol{}
	raiz := a.agregar(nil)
	hija := a.agregar(&raiz)

	err := validarPadreCategoria(raiz, &hija, a.leer, a.esDescendiente)
	if msg := campoPadreID(t, err); msg != "crea un ciclo" {
		t.Fatalf("padreId = %q, se esperaba %q", msg, "crea un ciclo")
	}
}

func TestVarianteInactivaNoSeVende(t *testing.T) {
	casos := []struct {
		nombre   string
		estado   string
		vendible bool
	}{
		// Sin stock, la variante sigue en el catálogo como «SIN STOCK»: el
		// stock lo decide inventario, no el estado.
		{"activa, aunque esté agotada", estadoVarianteActiva, true},
		// Desactivar no equivale a agotar: no se ofrece aunque tenga stock.
		{"inactiva, aunque conserve existencias", estadoVarianteInactiva, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := esVendible(c.estado); got != c.vendible {
				t.Fatalf("esVendible(%q) = %v, se esperaba %v", c.estado, got, c.vendible)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
