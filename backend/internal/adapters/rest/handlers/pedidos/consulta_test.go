package pedidos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

// lectorFalso imita las consultas del repositorio, incluida la que importa:
// PorIDYUsuario solo devuelve el pedido si el id Y el dueño coinciden, como el
// WHERE de la consulta real.
type lectorFalso struct {
	pedidos      []domain.Pedido
	publicos     map[string]domain.PedidoPublico
	ultimoLimit  int
	ultimoOffset int
}

func (l *lectorFalso) PorIDYUsuario(_ context.Context, id, usuarioID uuid.UUID) (domain.Pedido, error) {
	for _, p := range l.pedidos {
		if p.ID == id && p.UsuarioID != nil && *p.UsuarioID == usuarioID {
			return p, nil
		}
	}
	return domain.Pedido{}, apierr.ErrNoEncontrado
}

func (l *lectorFalso) PorToken(_ context.Context, token string) (domain.PedidoPublico, error) {
	if p, ok := l.publicos[token]; ok {
		return p, nil
	}
	return domain.PedidoPublico{}, apierr.ErrNoEncontrado
}

func (l *lectorFalso) ListarDeUsuario(_ context.Context, usuarioID uuid.UUID, limit, offset int) ([]domain.Pedido, int64, error) {
	l.ultimoLimit, l.ultimoOffset = limit, offset
	var propios []domain.Pedido
	for _, p := range l.pedidos {
		if p.UsuarioID != nil && *p.UsuarioID == usuarioID {
			propios = append(propios, p)
		}
	}
	return propios, int64(len(propios)), nil
}

type escenarioConsulta struct {
	handler *Handler
	router  http.Handler
	lector  *lectorFalso
	ana     uuid.UUID // dueña del pedido
	beto    uuid.UUID // otro cliente
	pedido  domain.Pedido
	token   string
}

func nuevoEscenarioConsulta() *escenarioConsulta {
	ana, beto := uuid.New(), uuid.New()
	pedido := domain.Pedido{
		ID: uuid.New(), Numero: 7, UsuarioID: &ana,
		CompradorNombre: "Ana", CompradorCorreo: "ana@example.com", CompradorDNI: "30111222", CompradorTelefono: "3415550000",
		ModoEntrega: domain.EntregaRetiro, MedioPago: domain.PagoMercadoPago,
		EstadoPedido: domain.PedidoConfirmado, EstadoPago: domain.PagoAprobado,
		SubtotalCentavos: 3000100, TotalCentavos: 3000100, TokenAcceso: strings.Repeat("A", 43),
		CreadoEn: time.Now(),
		Detalles: []domain.DetallePedido{{VarianteID: uuid.New(), Codigo: "PERF-01", Nombre: "Perfume", Unidades: 2, PrecioUnitarioCentavos: 1500050, SubtotalCentavos: 3000100}},
	}
	token := strings.Repeat("B", 43)
	lector := &lectorFalso{
		pedidos: []domain.Pedido{pedido},
		publicos: map[string]domain.PedidoPublico{token: {
			Numero: 7, ModoEntrega: domain.EntregaRetiro, MedioPago: domain.PagoMercadoPago,
			EstadoPedido: domain.PedidoConfirmado, EstadoPago: domain.PagoAprobado,
			SubtotalCentavos: 3000100, TotalCentavos: 3000100, CreadoEn: time.Now(),
			Detalles: []domain.DetallePublico{{Codigo: "PERF-01", Nombre: "Perfume", Unidades: 2, PrecioUnitarioCentavos: 1500050, SubtotalCentavos: 3000100}},
		}},
	}
	h := NuevoHandler(nil, uc.NuevaConsulta(lector), nil, nil)
	r := chi.NewRouter()
	r.Get("/api/pedidos", h.ListarMisPedidos)
	r.Get("/api/pedidos/publico/{token}", h.ObtenerPedidoPublico)
	r.Get("/api/pedidos/{id}", h.ObtenerPedido)
	return &escenarioConsulta{handler: h, router: r, lector: lector, ana: ana, beto: beto, pedido: pedido, token: token}
}

// conSesion simula la sesión de un usuario (nil = sin sesión).
func (e *escenarioConsulta) conSesion(id *uuid.UUID) {
	e.handler.usuario = func(context.Context) (middleware.Usuario, bool) {
		if id == nil {
			return middleware.Usuario{}, false
		}
		return middleware.Usuario{ID: *id}, true
	}
}

func (e *escenarioConsulta) get(ruta string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	return rec
}

func claves(t *testing.T, cuerpo []byte) []string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(cuerpo, &m); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// ---- GET /api/pedidos/{id} ----

func TestObtenerPedido_EsDelUsuario_200SinToken(t *testing.T) {
	e := nuevoEscenarioConsulta()
	e.conSesion(&e.ana)

	rec := e.get("/api/pedidos/" + e.pedido.ID.String())
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, vino %d: %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "tokenAcceso") || strings.Contains(rec.Body.String(), e.pedido.TokenAcceso) {
		t.Error("la consulta propia no debe devolver el token del enlace de invitado")
	}
	if !strings.Contains(rec.Body.String(), `"precioUnitarioCentavos":1500050`) {
		t.Errorf("debía mostrar el precio histórico: %s", rec.Body)
	}
}

// Pregunta de la defensa: "mostrame qué pasa si le pido este pedido con la
// cuenta de otro". Es un 404, y es EXACTAMENTE el mismo 404 que un id
// inexistente: no se puede distinguir "no existe" de "no es tuyo".
func TestObtenerPedido_DeOtroUsuario_404IgualAUnInexistente(t *testing.T) {
	e := nuevoEscenarioConsulta()
	e.conSesion(&e.beto) // Beto pide el pedido de Ana

	ajeno := e.get("/api/pedidos/" + e.pedido.ID.String())
	inexistente := e.get("/api/pedidos/" + uuid.New().String())

	if ajeno.Code != http.StatusNotFound {
		t.Fatalf("un pedido ajeno debe dar 404 (no 403), vino %d", ajeno.Code)
	}
	if ajeno.Body.String() != inexistente.Body.String() || ajeno.Code != inexistente.Code {
		t.Errorf("el 404 ajeno debe ser idéntico al de un id inexistente:\n%s\n%s", ajeno.Body, inexistente.Body)
	}
}

func TestObtenerPedido_IdQueNoEsUUID_404(t *testing.T) {
	e := nuevoEscenarioConsulta()
	e.conSesion(&e.ana)
	if rec := e.get("/api/pedidos/42"); rec.Code != http.StatusNotFound {
		t.Fatalf("esperaba 404, vino %d", rec.Code)
	}
}

func TestPedidosPropios_SinSesion_401(t *testing.T) {
	e := nuevoEscenarioConsulta()
	e.conSesion(nil)
	for _, ruta := range []string{"/api/pedidos", "/api/pedidos/" + e.pedido.ID.String()} {
		if rec := e.get(ruta); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s sin sesión: esperaba 401, vino %d", ruta, rec.Code)
		}
	}
}

// ---- GET /api/pedidos ----

func TestListarMisPedidos_SoloLosPropios(t *testing.T) {
	e := nuevoEscenarioConsulta()

	e.conSesion(&e.ana)
	var propio struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}
	_ = json.Unmarshal(e.get("/api/pedidos").Body.Bytes(), &propio)
	if propio.Total != 1 || len(propio.Items) != 1 {
		t.Fatalf("Ana debía ver su único pedido: %+v", propio)
	}

	e.conSesion(&e.beto)
	var ajeno struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.Unmarshal(e.get("/api/pedidos").Body.Bytes(), &ajeno)
	if len(ajeno.Items) != 0 {
		t.Errorf("Beto no debe ver pedidos de Ana: %+v", ajeno)
	}
}

func TestListarMisPedidos_Paginacion(t *testing.T) {
	e := nuevoEscenarioConsulta()
	e.conSesion(&e.ana)

	e.get("/api/pedidos?pagina=3&tamano=10")
	if e.lector.ultimoLimit != 10 || e.lector.ultimoOffset != 20 {
		t.Errorf("página 3 de 10 → limit 10, offset 20; vino %d/%d", e.lector.ultimoLimit, e.lector.ultimoOffset)
	}

	e.get("/api/pedidos?tamano=1000000") // un cliente no puede traerse la tabla entera
	if e.lector.ultimoLimit != uc.TamanoPaginaMaximo {
		t.Errorf("el tamaño máximo es %d, vino %d", uc.TamanoPaginaMaximo, e.lector.ultimoLimit)
	}

	if rec := e.get("/api/pedidos?pagina=abc"); rec.Code != http.StatusBadRequest {
		t.Errorf("una página que no es número debe dar 400, vino %d", rec.Code)
	}
}

// ---- GET /api/pedidos/publico/{token} ----

func TestObtenerPedidoPublico_SinSesion_VistaReducida(t *testing.T) {
	e := nuevoEscenarioConsulta()
	e.conSesion(nil) // el invitado no tiene sesión: el secreto es el token

	rec := e.get("/api/pedidos/publico/" + e.token)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, vino %d: %s", rec.Code, rec.Body)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("el enlace privado no debe cachearse")
	}

	// Lista cerrada: si alguien agrega un campo a la respuesta pública, este
	// test falla y obliga a decidirlo a propósito.
	esperadas := []string{"creadoEn", "detalles", "envioCentavos", "estadoPago", "estadoPedido", "medioPago",
		"modoEntrega", "numero", "subtotalCentavos", "totalCentavos"}
	if got := claves(t, rec.Body.Bytes()); strings.Join(got, ",") != strings.Join(esperadas, ",") {
		t.Errorf("campos públicos inesperados:\n got %v\nwant %v", got, esperadas)
	}
	for _, prohibido := range []string{"usuario", "correo", "dni", "telefono", "comprador", "tokenAcceso", "varianteId", `"id"`} {
		if strings.Contains(rec.Body.String(), prohibido) {
			t.Errorf("la vista pública no debe contener %q: %s", prohibido, rec.Body)
		}
	}

	var r struct {
		Detalles []map[string]any `json:"detalles"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	if got := claves(t, mustJSON(t, r.Detalles[0])); strings.Join(got, ",") != "codigo,nombre,precioUnitarioCentavos,subtotalCentavos,unidades" {
		t.Errorf("campos de línea inesperados: %v", got)
	}
}

func TestObtenerPedidoPublico_TokenInvalido_404(t *testing.T) {
	e := nuevoEscenarioConsulta()
	e.conSesion(nil)
	casos := map[string]string{
		"largo incorrecto":           "corto",
		"del largo pero inexistente": strings.Repeat("Z", 43),
	}
	for nombre, token := range casos {
		if rec := e.get("/api/pedidos/publico/" + token); rec.Code != http.StatusNotFound {
			t.Errorf("%s: esperaba 404, vino %d", nombre, rec.Code)
		}
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
