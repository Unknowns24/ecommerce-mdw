package pedidos

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// carritoFalso guarda carritos en memoria. Respeta lo que importa del
// repositorio real: AgregarItem suma si la variante ya estaba (upsert) y las
// operaciones sobre una línea llevan el id del carrito.
type carritoFalso struct {
	carritos map[uuid.UUID]*domain.Carrito // por usuario
}

func (f *carritoFalso) PorUsuario(_ context.Context, u uuid.UUID) (domain.Carrito, error) {
	c, ok := f.carritos[u]
	if !ok {
		return domain.Carrito{}, apierr.ErrNoEncontrado
	}
	copia := *c
	copia.Items = append([]domain.ItemCarrito(nil), c.Items...)
	return copia, nil
}

func (f *carritoFalso) CrearSiNoExiste(ctx context.Context, u uuid.UUID) (domain.Carrito, error) {
	if _, ok := f.carritos[u]; !ok {
		f.carritos[u] = &domain.Carrito{ID: uuid.New(), UsuarioID: &u}
	}
	return f.PorUsuario(ctx, u)
}

func (f *carritoFalso) buscar(carritoID uuid.UUID) *domain.Carrito {
	for _, c := range f.carritos {
		if c.ID == carritoID {
			return c
		}
	}
	return nil
}

func (f *carritoFalso) AgregarItem(_ context.Context, carritoID, varianteID uuid.UUID, unidades int) (domain.ItemCarrito, error) {
	c := f.buscar(carritoID)
	for i := range c.Items {
		if c.Items[i].VarianteID == varianteID {
			c.Items[i].Unidades += unidades
			return c.Items[i], nil
		}
	}
	it := domain.ItemCarrito{ID: uuid.New(), CarritoID: carritoID, VarianteID: varianteID, Unidades: unidades}
	c.Items = append(c.Items, it)
	return it, nil
}

func (f *carritoFalso) CambiarUnidades(_ context.Context, carritoID, itemID uuid.UUID, unidades int) error {
	c := f.buscar(carritoID)
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			c.Items[i].Unidades = unidades
			return nil
		}
	}
	return apierr.ErrNoEncontrado
}

func (f *carritoFalso) QuitarItem(_ context.Context, carritoID, itemID uuid.UUID) error {
	c := f.buscar(carritoID)
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return nil
		}
	}
	return apierr.ErrNoEncontrado
}

func (f *carritoFalso) Vaciar(_ context.Context, carritoID uuid.UUID) error {
	f.buscar(carritoID).Items = nil
	return nil
}

type consultaStockFalsa map[uuid.UUID]int

func (s consultaStockFalsa) Disponible(_ context.Context, id uuid.UUID) (int, error) {
	return s[id], nil
}

type escenarioCarrito struct {
	carrito  *GestionCarrito
	repo     *carritoFalso
	catalogo catalogoFalso
	stock    consultaStockFalsa
	perfume  out.VarianteVendible
	crema    out.VarianteVendible
	ana      uuid.UUID
}

func nuevoEscenarioCarrito() *escenarioCarrito {
	perfume := out.VarianteVendible{ID: uuid.New(), Codigo: "PERF-01", Nombre: "Perfume 100ml", PrecioMinoristaCentavos: 1500000, Activa: true}
	crema := out.VarianteVendible{ID: uuid.New(), Codigo: "CREM-01", Nombre: "Crema", PrecioMinoristaCentavos: 250050, Activa: true}
	e := &escenarioCarrito{
		repo:     &carritoFalso{carritos: map[uuid.UUID]*domain.Carrito{}},
		catalogo: catalogoFalso{variantes: map[uuid.UUID]out.VarianteVendible{perfume.ID: perfume, crema.ID: crema}},
		stock:    consultaStockFalsa{perfume.ID: 5, crema.ID: 10},
		perfume:  perfume,
		crema:    crema,
		ana:      uuid.New(),
	}
	e.carrito = NuevaGestionCarrito(e.repo, e.catalogo, e.stock)
	return e
}

func TestCarrito_SinCarrito_EsVacio(t *testing.T) {
	e := nuevoEscenarioCarrito()
	v, err := e.carrito.Ver(context.Background(), e.ana)
	if err != nil || len(v.Lineas) != 0 || v.TotalEstimadoCentavos != 0 {
		t.Fatalf("un usuario sin carrito ve uno vacío: %+v, %v", v, err)
	}
}

func TestCarrito_AgregarYSumarUnidades(t *testing.T) {
	e := nuevoEscenarioCarrito()
	ctx := context.Background()

	v, nueva, err := e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 2)
	if err != nil || !nueva || len(v.Lineas) != 1 || v.Lineas[0].Unidades != 2 {
		t.Fatalf("primera vez: %+v nueva=%v err=%v", v, nueva, err)
	}
	// Agregar la misma variante SUMA unidades, no crea otra línea.
	v, nueva, err = e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 1)
	if err != nil || nueva || len(v.Lineas) != 1 || v.Lineas[0].Unidades != 3 {
		t.Fatalf("segunda vez: %+v nueva=%v err=%v", v, nueva, err)
	}
	if v.Lineas[0].SubtotalCentavos != 3*1500000 || v.TotalEstimadoCentavos != 3*1500000 {
		t.Errorf("importes mal calculados: %+v", v)
	}
}

func TestCarrito_Disponibilidad_409NoEs400(t *testing.T) {
	e := nuevoEscenarioCarrito()
	ctx := context.Background()

	// Borde: pedir EXACTAMENTE el disponible (5) está permitido.
	if _, _, err := e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 5); err != nil {
		t.Fatalf("pedir justo el disponible debe andar: %v", err)
	}
	// Uno más que el disponible: 409 (ErrRegla), con el mapa variante → disponible.
	_, _, err := e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 1)
	var r apierr.ErrRegla
	if !errors.As(err, &r) {
		t.Fatalf("esperaba ErrRegla (409), vino %v", err)
	}
	if datos, ok := r.Datos.(map[string]int); !ok || datos[e.perfume.ID.String()] != 5 {
		t.Errorf("los Datos deben enumerar variante → disponible: %#v", r.Datos)
	}
	// Y no quedó agregada.
	if v, _ := e.carrito.Ver(ctx, e.ana); v.Lineas[0].Unidades != 5 {
		t.Errorf("el rechazo no debe modificar el carrito: %+v", v)
	}
}

func TestCarrito_LoQueYaTengoCuentaParaElDisponible(t *testing.T) {
	e := nuevoEscenarioCarrito()
	ctx := context.Background()
	_, _, _ = e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 3)

	// Tengo 3, hay 5: agregar 3 más (total 6) debe rechazarse.
	if _, _, err := e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 3); !esRegla(err) {
		t.Fatalf("la suma 3+3 supera el disponible 5: %v", err)
	}
}

func TestCarrito_AgregarInvalido(t *testing.T) {
	e := nuevoEscenarioCarrito()
	ctx := context.Background()

	for _, u := range []int{0, -3} {
		if _, _, err := e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, u); !esValidacion(err) {
			t.Errorf("unidades %d: esperaba ErrValidacion (400), vino %v", u, err)
		}
	}
	if _, _, err := e.carrito.AgregarItem(ctx, e.ana, uuid.New(), 1); !esNoEncontrado(err) {
		t.Errorf("variante inexistente: esperaba 404, vino %v", err)
	}
	if _, ok := e.repo.carritos[e.ana]; ok {
		t.Error("un pedido inválido no debe ni crear el carrito")
	}
}

// El carrito no congela precios: si cambia el catálogo, la próxima lectura lo
// refleja.
func TestCarrito_RecalculaConElCatalogoVigente(t *testing.T) {
	e := nuevoEscenarioCarrito()
	ctx := context.Background()
	_, _, _ = e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 2)

	subio := e.perfume
	subio.PrecioMinoristaCentavos = 2000000
	e.catalogo.variantes[e.perfume.ID] = subio

	v, _ := e.carrito.Ver(ctx, e.ana)
	if v.TotalEstimadoCentavos != 2*2000000 {
		t.Errorf("el total debe usar el precio vigente: %d", v.TotalEstimadoCentavos)
	}
}

func TestCarrito_VarianteDesactivada_QuedaMarcadaYNoSuma(t *testing.T) {
	e := nuevoEscenarioCarrito()
	ctx := context.Background()
	_, _, _ = e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 1)
	_, _, _ = e.carrito.AgregarItem(ctx, e.ana, e.crema.ID, 2)

	inactiva := e.perfume
	inactiva.Activa = false
	e.catalogo.variantes[e.perfume.ID] = inactiva

	v, err := e.carrito.Ver(ctx, e.ana)
	if err != nil || len(v.Lineas) != 2 {
		t.Fatalf("la línea sigue visible: %+v, %v", v, err)
	}
	if v.Lineas[0].Disponible || v.Lineas[0].PrecioUnitarioCentavos != 0 {
		t.Errorf("la línea desactivada debe quedar marcada y sin precio: %+v", v.Lineas[0])
	}
	if v.TotalEstimadoCentavos != 2*250050 {
		t.Errorf("solo suman las líneas vendibles: %d", v.TotalEstimadoCentavos)
	}
}

func TestCarrito_CambiarUnidades(t *testing.T) {
	e := nuevoEscenarioCarrito()
	ctx := context.Background()
	v, _, _ := e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 2)
	item := v.Lineas[0].ItemID

	v, err := e.carrito.CambiarUnidades(ctx, e.ana, item, 4)
	if err != nil || v.Lineas[0].Unidades != 4 {
		t.Fatalf("debía quedar en 4: %+v, %v", v, err)
	}
	// CAMBIAR fija el valor (no suma): 5 es el máximo disponible.
	if _, err := e.carrito.CambiarUnidades(ctx, e.ana, item, 5); err != nil {
		t.Errorf("5 es justo el disponible: %v", err)
	}
	if _, err := e.carrito.CambiarUnidades(ctx, e.ana, item, 6); !esRegla(err) {
		t.Errorf("6 supera el disponible: esperaba 409, vino %v", err)
	}
	if _, err := e.carrito.CambiarUnidades(ctx, e.ana, item, 0); !esValidacion(err) {
		t.Errorf("0 unidades: esperaba 400, vino %v", err)
	}
}

// Una línea de otro carrito, para esta llamada, no existe: 404.
func TestCarrito_LineaAjena_404(t *testing.T) {
	e := nuevoEscenarioCarrito()
	ctx := context.Background()
	beto := uuid.New()
	v, _, _ := e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 1)
	itemDeAna := v.Lineas[0].ItemID
	_, _, _ = e.carrito.AgregarItem(ctx, beto, e.crema.ID, 1) // Beto tiene su propio carrito

	if _, err := e.carrito.CambiarUnidades(ctx, beto, itemDeAna, 2); !esNoEncontrado(err) {
		t.Errorf("cambiar la línea de otro: esperaba 404, vino %v", err)
	}
	if err := e.carrito.QuitarItem(ctx, beto, itemDeAna); !esNoEncontrado(err) {
		t.Errorf("quitar la línea de otro: esperaba 404, vino %v", err)
	}
	if v, _ := e.carrito.Ver(ctx, e.ana); len(v.Lineas) != 1 || v.Lineas[0].Unidades != 1 {
		t.Errorf("el carrito de Ana no debe haberse tocado: %+v", v)
	}
	// Quien no tiene carrito tampoco puede tocar nada.
	if err := e.carrito.QuitarItem(ctx, uuid.New(), itemDeAna); !esNoEncontrado(err) {
		t.Errorf("sin carrito: esperaba 404, vino %v", err)
	}
}

func TestCarrito_QuitarYVaciar(t *testing.T) {
	e := nuevoEscenarioCarrito()
	ctx := context.Background()
	_, _, _ = e.carrito.AgregarItem(ctx, e.ana, e.perfume.ID, 1)
	v, _, _ := e.carrito.AgregarItem(ctx, e.ana, e.crema.ID, 1)

	if err := e.carrito.QuitarItem(ctx, e.ana, v.Lineas[1].ItemID); err != nil {
		t.Fatal(err)
	}
	if v, _ := e.carrito.Ver(ctx, e.ana); len(v.Lineas) != 1 {
		t.Errorf("debía quedar una línea: %+v", v)
	}
	if err := e.carrito.Vaciar(ctx, e.ana); err != nil {
		t.Fatal(err)
	}
	if v, _ := e.carrito.Ver(ctx, e.ana); len(v.Lineas) != 0 {
		t.Errorf("debía quedar vacío: %+v", v)
	}
	// Vaciar sin carrito no es un error.
	if err := e.carrito.Vaciar(ctx, uuid.New()); err != nil {
		t.Errorf("vaciar sin carrito debe ser un éxito: %v", err)
	}
}

func TestValidarDisponibilidad(t *testing.T) {
	id := uuid.New()
	casos := []struct {
		total, disponible int
		ok                bool
	}{
		{1, 5, true}, {5, 5, true}, {6, 5, false}, {1, 0, false}, {0, 0, true},
	}
	for _, c := range casos {
		if err := ValidarDisponibilidad(id, c.total, c.disponible); (err == nil) != c.ok {
			t.Errorf("total=%d disponible=%d: ok esperado %v, error %v", c.total, c.disponible, c.ok, err)
		}
	}
}
