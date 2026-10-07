package pedidos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

// gestorPedidosFalso guarda pedidos en memoria y respeta la regla del
// repositorio real: CambiarEstado solo actúa si el estado actual es el de origen.
type gestorPedidosFalso struct{ pedidos map[uuid.UUID]*domain.Pedido }

func (g *gestorPedidosFalso) PorID(_ context.Context, id uuid.UUID) (domain.Pedido, error) {
	if p, ok := g.pedidos[id]; ok {
		return *p, nil
	}
	return domain.Pedido{}, apierr.ErrNoEncontrado
}

func (g *gestorPedidosFalso) ListarAdmin(_ context.Context, f out.FiltrosPedidos, _, _ int) ([]domain.Pedido, int64, error) {
	var res []domain.Pedido
	for _, p := range g.pedidos {
		if f.Estado == nil || p.EstadoPedido == *f.Estado {
			res = append(res, *p)
		}
	}
	return res, int64(len(res)), nil
}

func (g *gestorPedidosFalso) CambiarEstado(_ context.Context, _ *gorm.DB, id uuid.UUID, desde, nuevo domain.EstadoPedido) error {
	p := g.pedidos[id]
	if p.EstadoPedido != desde {
		return apierr.ErrRegla{Mensaje: "El pedido cambió de estado mientras lo procesabas."}
	}
	p.EstadoPedido = nuevo
	return nil
}

func (g *gestorPedidosFalso) MarcarPagoAprobado(_ context.Context, _ *gorm.DB, id uuid.UUID) error {
	g.pedidos[id].EstadoPago = domain.PagoAprobado
	return nil
}

func (g *gestorPedidosFalso) RegistrarCancelacion(_ context.Context, _ *gorm.DB, id, resp uuid.UUID, motivo string, ahora time.Time) error {
	p := g.pedidos[id]
	p.CanceladoPor, p.CanceladoEn, p.MotivoCancelacion = &resp, &ahora, &motivo
	return nil
}

type gestorStockFalso struct{ liberaciones int }

func (s *gestorStockFalso) Confirmar(context.Context, *gorm.DB, uuid.UUID) error { return nil }
func (s *gestorStockFalso) Liberar(context.Context, *gorm.DB, uuid.UUID) error {
	s.liberaciones++
	return nil
}

type escenarioAdmin struct {
	router http.Handler
	stock  *gestorStockFalso
	admin  uuid.UUID
}

func nuevoEscenarioAdmin(sesion bool, pedidos ...*domain.Pedido) *escenarioAdmin {
	g := &gestorPedidosFalso{pedidos: map[uuid.UUID]*domain.Pedido{}}
	for _, p := range pedidos {
		g.pedidos[p.ID] = p
	}
	stock := &gestorStockFalso{}
	gestion := uc.NuevaGestion(g, stock, txFalsa{}, time.Now)

	h := NuevoHandler(nil, nil, gestion, nil)
	admin := uuid.New()
	h.usuario = func(context.Context) (middleware.Usuario, bool) {
		return middleware.Usuario{ID: admin}, sesion
	}

	r := chi.NewRouter()
	r.Get("/api/admin/pedidos", h.ListarPedidosAdmin)
	r.Get("/api/admin/pedidos/{id}", h.ObtenerPedidoAdmin)
	r.Post("/api/admin/pedidos/{id}/despacho", h.DespacharPedido)
	r.Post("/api/admin/pedidos/{id}/entrega", h.EntregarPedido)
	r.Post("/api/admin/pedidos/{id}/cobro-efectivo", h.CobrarPedidoEnEfectivo)
	r.Post("/api/admin/pedidos/{id}/cancelacion", h.CancelarPedido)
	return &escenarioAdmin{router: r, stock: stock, admin: admin}
}

func (e *escenarioAdmin) llamar(metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo)))
	return rec
}

func pedidoAdmin(modo domain.ModoEntrega, estado domain.EstadoPedido, pago domain.EstadoPago, medio domain.MedioPago) *domain.Pedido {
	return &domain.Pedido{
		ID: uuid.New(), ModoEntrega: modo, EstadoPedido: estado, EstadoPago: pago, MedioPago: medio,
		CompradorNombre: "Ana", CompradorApellido: "Pérez", CompradorCorreo: "ana@example.com",
		CompradorTelefono: "3415550000", CompradorDNI: "30111222", TokenAcceso: strings.Repeat("A", 43),
	}
}

func ruta(p *domain.Pedido, sufijo string) string {
	return "/api/admin/pedidos/" + p.ID.String() + sufijo
}

// La demo del martes: despachar un pedido no pagado es 409 (no 500), con un
// mensaje que dice qué hacer.
func TestDespacho_PedidoNoPagado_409(t *testing.T) {
	p := pedidoAdmin(domain.EntregaEnvio, domain.PedidoPendienteDePago, domain.PagoPendiente, domain.PagoMercadoPago)
	e := nuevoEscenarioAdmin(true, p)

	rec := e.llamar("POST", ruta(p, "/despacho"), "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("esperaba 409, vino %d: %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "todavía no fue pagado") {
		t.Errorf("el mensaje debe decir qué pasa: %s", rec.Body)
	}
}

func TestDespachoYEntrega_FlujoCompleto(t *testing.T) {
	p := pedidoAdmin(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago)
	e := nuevoEscenarioAdmin(true, p)

	if rec := e.llamar("POST", ruta(p, "/entrega"), ""); rec.Code != http.StatusConflict {
		t.Fatalf("entregar sin despachar debe dar 409, vino %d", rec.Code)
	}
	rec := e.llamar("POST", ruta(p, "/despacho"), "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"estadoPedido":"DESPACHADO"`) {
		t.Fatalf("despacho: %d %s", rec.Code, rec.Body)
	}
	rec = e.llamar("POST", ruta(p, "/entrega"), "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"estadoPedido":"COMPLETADO"`) {
		t.Fatalf("entrega: %d %s", rec.Code, rec.Body)
	}
}

func TestTransiciones_PedidoInexistenteOIdInvalido_404(t *testing.T) {
	e := nuevoEscenarioAdmin(true)
	for _, ruta := range []string{
		"/api/admin/pedidos/" + uuid.New().String() + "/despacho",
		"/api/admin/pedidos/no-es-un-uuid/entrega",
		"/api/admin/pedidos/" + uuid.New().String() + "/cobro-efectivo",
	} {
		if rec := e.llamar("POST", ruta, ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s: esperaba 404, vino %d", ruta, rec.Code)
		}
	}
}

func TestCobroEfectivo_CompletaYCobra(t *testing.T) {
	p := pedidoAdmin(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoPendiente, domain.PagoEfectivo)
	e := nuevoEscenarioAdmin(true, p)

	rec := e.llamar("POST", ruta(p, "/cobro-efectivo"), "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"estadoPedido":"COMPLETADO"`) ||
		!strings.Contains(rec.Body.String(), `"estadoPago":"APROBADO"`) {
		t.Fatalf("cobro: %d %s", rec.Code, rec.Body)
	}
	if rec := e.llamar("POST", ruta(p, "/cobro-efectivo"), ""); rec.Code != http.StatusConflict {
		t.Errorf("cobrar dos veces debe dar 409, vino %d", rec.Code)
	}
}

// ---- cancelación ----

func TestCancelacion_Pagado_AvisaReintegroManual_YElResponsableSaleDeLaSesion(t *testing.T) {
	p := pedidoAdmin(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago)
	e := nuevoEscenarioAdmin(true, p)
	otro := uuid.New()

	// El body intenta colar un "canceladoPor": se ignora; manda la sesión.
	rec := e.llamar("POST", ruta(p, "/cancelacion"), `{"motivo": "El cliente lo pidió", "canceladoPor": "`+otro.String()+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, vino %d: %s", rec.Code, rec.Body)
	}
	var r struct {
		Error  *string `json:"error"`
		Aviso  string  `json:"aviso"`
		Pedido struct {
			Estado       string `json:"estadoPedido"`
			CanceladoPor string `json:"canceladoPor"`
			Motivo       string `json:"motivoCancelacion"`
		} `json:"pedido"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	if r.Error != nil || !strings.Contains(rec.Body.String(), `"error":null`) {
		t.Errorf(`la respuesta debe llevar "error": null: %s`, rec.Body)
	}
	if r.Aviso != "El reintegro debe realizarse manualmente en Mercado Pago. Este sistema no ejecuta devoluciones." {
		t.Errorf("aviso inesperado: %q", r.Aviso)
	}
	if r.Pedido.Estado != "CANCELADO" || r.Pedido.Motivo != "El cliente lo pidió" {
		t.Errorf("pedido mal cancelado: %+v", r.Pedido)
	}
	if r.Pedido.CanceladoPor != e.admin.String() {
		t.Errorf("el responsable debe ser el de la sesión (%s), no el del body: %s", e.admin, r.Pedido.CanceladoPor)
	}
}

func TestCancelacion_NoPagado_SinAviso(t *testing.T) {
	p := pedidoAdmin(domain.EntregaRetiro, domain.PedidoPendienteDePago, domain.PagoPendiente, domain.PagoMercadoPago)
	e := nuevoEscenarioAdmin(true, p)

	rec := e.llamar("POST", ruta(p, "/cancelacion"), `{"motivo": "No pagó"}`)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "aviso") {
		t.Fatalf("un pedido sin pagar no lleva aviso de reintegro: %d %s", rec.Code, rec.Body)
	}
}

func TestCancelacion_DosVeces_SegundaEs409YElStockSeLiberaUnaVez(t *testing.T) {
	p := pedidoAdmin(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoPendiente, domain.PagoEfectivo)
	e := nuevoEscenarioAdmin(true, p)

	if rec := e.llamar("POST", ruta(p, "/cancelacion"), `{"motivo": "uno"}`); rec.Code != http.StatusOK {
		t.Fatalf("primera: %d %s", rec.Code, rec.Body)
	}
	if rec := e.llamar("POST", ruta(p, "/cancelacion"), `{"motivo": "dos"}`); rec.Code != http.StatusConflict {
		t.Fatalf("segunda: esperaba 409, vino %d", rec.Code)
	}
	if e.stock.liberaciones != 1 {
		t.Errorf("el stock se libera una sola vez, no %d", e.stock.liberaciones)
	}
}

func TestCancelacion_Validaciones(t *testing.T) {
	p := pedidoAdmin(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoPendiente, domain.PagoEfectivo)
	e := nuevoEscenarioAdmin(true, p)

	casos := map[string]string{
		"sin body":       "",
		"JSON roto":      `{"motivo": `,
		"motivo vacío":   `{"motivo": "   "}`,
		"motivo gigante": `{"motivo": "` + strings.Repeat("x", 501) + `"}`,
	}
	for nombre, cuerpo := range casos {
		rec := e.llamar("POST", ruta(p, "/cancelacion"), cuerpo)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "motivo") {
			t.Errorf("%s: esperaba 400 con el campo motivo, vino %d %s", nombre, rec.Code, rec.Body)
		}
	}
}

func TestCancelacion_Completado_409_SinSesion_401(t *testing.T) {
	completado := pedidoAdmin(domain.EntregaEnvio, domain.PedidoCompletado, domain.PagoAprobado, domain.PagoMercadoPago)
	e := nuevoEscenarioAdmin(true, completado)
	if rec := e.llamar("POST", ruta(completado, "/cancelacion"), `{"motivo": "x"}`); rec.Code != http.StatusConflict {
		t.Errorf("un pedido COMPLETADO no se cancela: esperaba 409, vino %d", rec.Code)
	}

	sinSesion := nuevoEscenarioAdmin(false, completado)
	if rec := sinSesion.llamar("POST", ruta(completado, "/cancelacion"), `{"motivo": "x"}`); rec.Code != http.StatusUnauthorized {
		t.Errorf("sin sesión: esperaba 401, vino %d", rec.Code)
	}
}

// ---- consulta administrativa ----

func TestAdmin_DetalleIncluyeComprador_PeroNoElToken(t *testing.T) {
	p := pedidoAdmin(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago)
	e := nuevoEscenarioAdmin(true, p)

	rec := e.llamar("GET", ruta(p, ""), "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "ana@example.com") {
		t.Fatalf("el detalle administrativo debe mostrar al comprador: %d %s", rec.Code, rec.Body)
	}
	if strings.Contains(rec.Body.String(), "tokenAcceso") || strings.Contains(rec.Body.String(), p.TokenAcceso) {
		t.Error("la administración no necesita (ni recibe) el token del enlace del invitado")
	}
}

func TestAdmin_Listado_Filtros(t *testing.T) {
	a := pedidoAdmin(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago)
	b := pedidoAdmin(domain.EntregaEnvio, domain.PedidoCancelado, domain.PagoPendiente, domain.PagoMercadoPago)
	e := nuevoEscenarioAdmin(true, a, b)

	var r struct {
		Total int `json:"total"`
	}
	rec := e.llamar("GET", "/api/admin/pedidos?estado=CANCELADO&desde=2026-10-01&hasta=2026-10-31", "")
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	if rec.Code != http.StatusOK || r.Total != 1 {
		t.Fatalf("el filtro por estado debía dejar 1 pedido: %d %s", rec.Code, rec.Body)
	}

	for nombre, query := range map[string]string{
		"estado inventado":   "?estado=VOLANDO",
		"fecha mal formada":  "?desde=01/10/2026",
		"rango invertido":    "?desde=2026-10-31&hasta=2026-10-01",
		"página no numérica": "?pagina=abc",
	} {
		if rec := e.llamar("GET", "/api/admin/pedidos"+query, ""); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, vino %d", nombre, rec.Code)
		}
	}
}
