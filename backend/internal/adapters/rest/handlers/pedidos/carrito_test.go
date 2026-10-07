package pedidos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

// repoCarritoFalso: carritos en memoria por usuario; las operaciones sobre una
// línea llevan el id del carrito, como en el repositorio real.
type repoCarritoFalso struct{ carritos map[uuid.UUID]*domain.Carrito }

func (f *repoCarritoFalso) PorUsuario(_ context.Context, u uuid.UUID) (domain.Carrito, error) {
	c, ok := f.carritos[u]
	if !ok {
		return domain.Carrito{}, apierr.ErrNoEncontrado
	}
	copia := *c
	copia.Items = append([]domain.ItemCarrito(nil), c.Items...)
	return copia, nil
}

func (f *repoCarritoFalso) CrearSiNoExiste(ctx context.Context, u uuid.UUID) (domain.Carrito, error) {
	if _, ok := f.carritos[u]; !ok {
		f.carritos[u] = &domain.Carrito{ID: uuid.New(), UsuarioID: &u}
	}
	return f.PorUsuario(ctx, u)
}

func (f *repoCarritoFalso) buscar(id uuid.UUID) *domain.Carrito {
	for _, c := range f.carritos {
		if c.ID == id {
			return c
		}
	}
	return nil
}

func (f *repoCarritoFalso) AgregarItem(_ context.Context, carritoID, varianteID uuid.UUID, unidades int) (domain.ItemCarrito, error) {
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

func (f *repoCarritoFalso) CambiarUnidades(_ context.Context, carritoID, itemID uuid.UUID, unidades int) error {
	c := f.buscar(carritoID)
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			c.Items[i].Unidades = unidades
			return nil
		}
	}
	return apierr.ErrNoEncontrado
}

func (f *repoCarritoFalso) QuitarItem(_ context.Context, carritoID, itemID uuid.UUID) error {
	c := f.buscar(carritoID)
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return nil
		}
	}
	return apierr.ErrNoEncontrado
}

func (f *repoCarritoFalso) Vaciar(_ context.Context, carritoID uuid.UUID) error {
	f.buscar(carritoID).Items = nil
	return nil
}

type disponibleFalso map[uuid.UUID]int

func (d disponibleFalso) Disponible(_ context.Context, id uuid.UUID) (int, error) { return d[id], nil }

type escenarioCarritoHTTP struct {
	handler *Handler
	router  http.Handler
	perfume uuid.UUID
	ana     uuid.UUID
	beto    uuid.UUID
}

func nuevoEscenarioCarritoHTTP() *escenarioCarritoHTTP {
	perfume := out.VarianteVendible{ID: uuid.New(), Codigo: "PERF-01", Nombre: "Perfume", PrecioMinoristaCentavos: 1500000, Activa: true}
	cat := catalogoFalso{map[uuid.UUID]out.VarianteVendible{perfume.ID: perfume}}
	gestion := uc.NuevaGestionCarrito(&repoCarritoFalso{carritos: map[uuid.UUID]*domain.Carrito{}}, cat, disponibleFalso{perfume.ID: 5})

	h := NuevoHandler(nil, nil, nil, gestion)
	r := chi.NewRouter()
	r.Get("/api/carrito", h.VerCarrito)
	r.Delete("/api/carrito", h.VaciarCarrito)
	r.Post("/api/carrito/items", h.AgregarAlCarrito)
	r.Patch("/api/carrito/items/{id}", h.CambiarUnidadesDelCarrito)
	r.Delete("/api/carrito/items/{id}", h.QuitarDelCarrito)

	e := &escenarioCarritoHTTP{handler: h, router: r, perfume: perfume.ID, ana: uuid.New(), beto: uuid.New()}
	e.conSesion(&e.ana)
	return e
}

func (e *escenarioCarritoHTTP) conSesion(id *uuid.UUID) {
	e.handler.usuario = func(context.Context) (middleware.Usuario, bool) {
		if id == nil {
			return middleware.Usuario{}, false
		}
		return middleware.Usuario{ID: *id}, true
	}
}

func (e *escenarioCarritoHTTP) llamar(metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo)))
	return rec
}

func (e *escenarioCarritoHTTP) agregar(unidades string) *httptest.ResponseRecorder {
	return e.llamar("POST", "/api/carrito/items", `{"varianteId": "`+e.perfume.String()+`", "unidades": `+unidades+`}`)
}

type carritoJSON struct {
	Items []struct {
		ID       string `json:"id"`
		Unidades int    `json:"unidades"`
		Precio   int64  `json:"precioUnitarioCentavos"`
	} `json:"items"`
	Total int64 `json:"totalEstimadoCentavos"`
}

func leerCarrito(t *testing.T, rec *httptest.ResponseRecorder) carritoJSON {
	t.Helper()
	var c carritoJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &c); err != nil {
		t.Fatalf("JSON inválido: %v (%s)", err, rec.Body)
	}
	return c
}

func TestCarritoHTTP_SinSesion_401EnTodasLasRutas(t *testing.T) {
	e := nuevoEscenarioCarritoHTTP()
	e.conSesion(nil) // el invitado guarda su carrito en el navegador, no acá
	id := uuid.New().String()

	for _, c := range [][2]string{
		{"GET", "/api/carrito"}, {"DELETE", "/api/carrito"}, {"POST", "/api/carrito/items"},
		{"PATCH", "/api/carrito/items/" + id}, {"DELETE", "/api/carrito/items/" + id},
	} {
		if rec := e.llamar(c[0], c[1], `{}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin sesión: esperaba 401, vino %d", c[0], c[1], rec.Code)
		}
	}
}

func TestCarritoHTTP_Vacio_200ConListaVaciaNoNull(t *testing.T) {
	e := nuevoEscenarioCarritoHTTP()
	rec := e.llamar("GET", "/api/carrito", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Fatalf("un carrito vacío es 200 con items []: %d %s", rec.Code, rec.Body)
	}
}

func TestCarritoHTTP_Agregar_201LaPrimeraVez_200AlSumar_PrecioDelServidor(t *testing.T) {
	e := nuevoEscenarioCarritoHTTP()

	// El body intenta mandar un precio: se ignora.
	rec := e.llamar("POST", "/api/carrito/items",
		`{"varianteId": "`+e.perfume.String()+`", "unidades": 2, "precioUnitarioCentavos": 1}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("línea nueva: esperaba 201, vino %d: %s", rec.Code, rec.Body)
	}
	c := leerCarrito(t, rec)
	if c.Items[0].Precio != 1500000 || c.Total != 3000000 {
		t.Errorf("el precio es el del servidor: %+v", c)
	}

	rec = e.agregar("1")
	if rec.Code != http.StatusOK {
		t.Fatalf("sumar a una línea existente: esperaba 200, vino %d", rec.Code)
	}
	if c := leerCarrito(t, rec); len(c.Items) != 1 || c.Items[0].Unidades != 3 {
		t.Errorf("debía sumar unidades en la misma línea: %+v", c)
	}
}

func TestCarritoHTTP_MasQueElDisponible_409NoEs400(t *testing.T) {
	e := nuevoEscenarioCarritoHTTP()
	rec := e.agregar("6") // hay 5

	if rec.Code != http.StatusConflict {
		t.Fatalf("esperaba 409 (el request está bien, el estado no lo permite), vino %d: %s", rec.Code, rec.Body)
	}
	var r struct {
		Detalles map[string]int `json:"detalles"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	if r.Detalles[e.perfume.String()] != 5 {
		t.Errorf("el 409 debe enumerar variante → disponible: %s", rec.Body)
	}
}

func TestCarritoHTTP_BodyInvalido_400(t *testing.T) {
	e := nuevoEscenarioCarritoHTTP()
	casos := map[string]string{
		"cero unidades":    `{"varianteId": "` + e.perfume.String() + `", "unidades": 0}`,
		"negativas":        `{"varianteId": "` + e.perfume.String() + `", "unidades": -2}`,
		"demasiadas":       `{"varianteId": "` + e.perfume.String() + `", "unidades": 1001}`,
		"variante no uuid": `{"varianteId": "perfume", "unidades": 1}`,
		"JSON roto":        `{"varianteId": `,
	}
	for nombre, cuerpo := range casos {
		if rec := e.llamar("POST", "/api/carrito/items", cuerpo); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, vino %d: %s", nombre, rec.Code, rec.Body)
		}
	}
}

func TestCarritoHTTP_VarianteInexistente_404(t *testing.T) {
	e := nuevoEscenarioCarritoHTTP()
	rec := e.llamar("POST", "/api/carrito/items", `{"varianteId": "`+uuid.New().String()+`", "unidades": 1}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperaba 404, vino %d", rec.Code)
	}
}

func TestCarritoHTTP_CambiarUnidades(t *testing.T) {
	e := nuevoEscenarioCarritoHTTP()
	item := leerCarrito(t, e.agregar("2")).Items[0].ID

	if rec := e.llamar("PATCH", "/api/carrito/items/"+item, `{"unidades": 4}`); rec.Code != http.StatusOK ||
		leerCarrito(t, rec).Items[0].Unidades != 4 {
		t.Errorf("debía quedar en 4: %d %s", rec.Code, rec.Body)
	}
	if rec := e.llamar("PATCH", "/api/carrito/items/"+item, `{"unidades": 6}`); rec.Code != http.StatusConflict {
		t.Errorf("más que el disponible: esperaba 409, vino %d", rec.Code)
	}
	if rec := e.llamar("PATCH", "/api/carrito/items/"+item, `{"unidades": 0}`); rec.Code != http.StatusBadRequest {
		t.Errorf("cero unidades: esperaba 400, vino %d", rec.Code)
	}
	if rec := e.llamar("PATCH", "/api/carrito/items/"+uuid.New().String(), `{"unidades": 1}`); rec.Code != http.StatusNotFound {
		t.Errorf("línea inexistente: esperaba 404, vino %d", rec.Code)
	}
	if rec := e.llamar("PATCH", "/api/carrito/items/no-es-uuid", `{"unidades": 1}`); rec.Code != http.StatusNotFound {
		t.Errorf("id que no es uuid: esperaba 404, vino %d", rec.Code)
	}
}

// Beto no puede tocar la línea de Ana: para él no existe.
func TestCarritoHTTP_LineaDeOtroUsuario_404(t *testing.T) {
	e := nuevoEscenarioCarritoHTTP()
	itemDeAna := leerCarrito(t, e.agregar("2")).Items[0].ID

	e.conSesion(&e.beto)
	if rec := e.llamar("PATCH", "/api/carrito/items/"+itemDeAna, `{"unidades": 1}`); rec.Code != http.StatusNotFound {
		t.Errorf("cambiar la línea de otro: esperaba 404, vino %d", rec.Code)
	}
	if rec := e.llamar("DELETE", "/api/carrito/items/"+itemDeAna, ""); rec.Code != http.StatusNotFound {
		t.Errorf("quitar la línea de otro: esperaba 404, vino %d", rec.Code)
	}

	e.conSesion(&e.ana)
	if c := leerCarrito(t, e.llamar("GET", "/api/carrito", "")); len(c.Items) != 1 || c.Items[0].Unidades != 2 {
		t.Errorf("el carrito de Ana no debe haberse tocado: %+v", c)
	}
}

func TestCarritoHTTP_QuitarYVaciar_204(t *testing.T) {
	e := nuevoEscenarioCarritoHTTP()
	item := leerCarrito(t, e.agregar("1")).Items[0].ID

	if rec := e.llamar("DELETE", "/api/carrito/items/"+item, ""); rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Errorf("quitar: esperaba 204 sin cuerpo, vino %d %q", rec.Code, rec.Body)
	}
	e.agregar("1")
	if rec := e.llamar("DELETE", "/api/carrito", ""); rec.Code != http.StatusNoContent {
		t.Errorf("vaciar: esperaba 204, vino %d", rec.Code)
	}
	if c := leerCarrito(t, e.llamar("GET", "/api/carrito", "")); len(c.Items) != 0 {
		t.Errorf("debía quedar vacío: %+v", c)
	}
}
