package pedidos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

// Versiones falsas mínimas de los puertos: el handler se prueba de punta a
// punta (JSON → caso de uso → JSON) sin base de datos.

type catalogoFalso struct {
	v map[uuid.UUID]out.VarianteVendible
}

func (c catalogoFalso) VariantePorID(_ context.Context, id uuid.UUID) (out.VarianteVendible, error) {
	if v, ok := c.v[id]; ok {
		return v, nil
	}
	return out.VarianteVendible{}, apierr.ErrNoEncontrado
}

type stockFalso struct{ disponible int }

func (s stockFalso) Reservar(_ context.Context, _ *gorm.DB, _ uuid.UUID, items []out.ItemReservaStock) error {
	faltan := map[string]int{}
	for _, it := range items {
		if it.Unidades > s.disponible {
			faltan[it.VarianteID.String()] = s.disponible
		}
	}
	if len(faltan) > 0 {
		return apierr.ErrRegla{Mensaje: "No hay stock suficiente", Datos: faltan}
	}
	return nil
}

type repoFalso struct{}

func (repoFalso) Crear(context.Context, *gorm.DB, *domain.Pedido, []domain.DetallePedido) error {
	return nil
}

type configFalsa struct{}

func (configFalsa) Actual(context.Context) (domain.ConfiguracionTienda, error) {
	return domain.ConfiguracionTienda{}, apierr.ErrNoEncontrado
}
func (configFalsa) TarifasOrdenadas(context.Context) ([]domain.TarifaDistancia, error) {
	return nil, nil
}

type txFalsa struct{}

func (txFalsa) EnTransaccion(_ context.Context, fn func(*gorm.DB) error) error { return fn(nil) }

const precioReal = 1500000 // $15.000,00

func armar(disponible int) (*Handler, uuid.UUID) {
	id := uuid.New()
	v := out.VarianteVendible{ID: id, Codigo: "PERF-01", Nombre: "Perfume", PrecioMinoristaCentavos: precioReal, Activa: true}
	co := uc.NuevoCheckout(catalogoFalso{map[uuid.UUID]out.VarianteVendible{id: v}}, stockFalso{disponible}, repoFalso{}, configFalsa{}, txFalsa{},
		15*time.Minute, time.Now)
	return NuevoHandler(co, nil, nil, nil), id
}

func post(h *Handler, cuerpo string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/pedidos", strings.NewReader(cuerpo))
	rec := httptest.NewRecorder()
	h.CrearPedido(rec, req)
	return rec
}

func cuerpoValido(varianteID uuid.UUID, unidades int, extra string) string {
	return fmt.Sprintf(`{
		"items": [{"varianteId": %q, "unidades": %d}],
		"comprador": {"nombre": "Ana", "apellido": "Pérez", "correo": "ana@example.com", "telefono": "3415550000", "dni": "30111222"},
		"modoEntrega": "RETIRO",
		"medioPago": "MERCADO_PAGO"%s
	}`, varianteID, unidades, extra)
}

func TestCrearPedido_FlujoFeliz_201ConTokenYEstadoPendiente(t *testing.T) {
	h, id := armar(10)
	rec := post(h, cuerpoValido(id, 2, ""))

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperaba 201, vino %d: %s", rec.Code, rec.Body)
	}
	var r map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	if r["tokenAcceso"] == "" || r["estadoPedido"] != "PENDIENTE_DE_PAGO" || r["estadoPago"] != "PENDIENTE" {
		t.Errorf("respuesta inesperada: %v", r)
	}
	if r["totalCentavos"] != float64(2*precioReal) {
		t.Errorf("total esperado %d, vino %v", 2*precioReal, r["totalCentavos"])
	}
}

// La demo del martes: mandar el precio (y el total, el estado y el usuario) en
// el body no sirve de nada. El pedido se crea con el precio del SERVIDOR.
func TestCrearPedido_IgnoraPrecioTotalEstadoYUsuarioDelBody(t *testing.T) {
	h, id := armar(10)
	extra := fmt.Sprintf(`, "precioUnitarioCentavos": 100, "totalCentavos": 100, "subtotalCentavos": 100,
		"estadoPago": "APROBADO", "estadoPedido": "COMPLETADO", "usuarioId": %q`, uuid.New())
	body := strings.Replace(cuerpoValido(id, 1, extra), `"unidades": 1}`, `"unidades": 1, "precioUnitarioCentavos": 1}`, 1)

	rec := post(h, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("esperaba 201, vino %d: %s", rec.Code, rec.Body)
	}
	var r struct {
		Total      int64  `json:"totalCentavos"`
		EstadoPago string `json:"estadoPago"`
		Estado     string `json:"estadoPedido"`
		Detalles   []struct {
			Precio int64 `json:"precioUnitarioCentavos"`
		} `json:"detalles"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	if r.Total != precioReal || r.Detalles[0].Precio != precioReal {
		t.Errorf("el precio debe ser el del servidor (%d): total=%d línea=%d", precioReal, r.Total, r.Detalles[0].Precio)
	}
	if r.EstadoPago != "PENDIENTE" || r.Estado != "PENDIENTE_DE_PAGO" {
		t.Errorf("los estados los decide el servidor: %s / %s", r.EstadoPago, r.Estado)
	}
}

func TestCrearPedido_BodyInvalido_400ConCampos(t *testing.T) {
	h, _ := armar(10)
	rec := post(h, `{"items": [], "comprador": {"nombre": "", "correo": "no-es-un-correo", "dni": "12.345"},
		"modoEntrega": "ENVIO", "medioPago": "BITCOIN"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, vino %d: %s", rec.Code, rec.Body)
	}
	var r struct {
		Detalles struct {
			Campos map[string]string `json:"campos"`
		} `json:"detalles"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	for _, campo := range []string{"items", "comprador.nombre", "comprador.correo", "comprador.dni", "domicilio", "distanciaMetros", "medioPago"} {
		if r.Detalles.Campos[campo] == "" {
			t.Errorf("faltó el error del campo %q: %v", campo, r.Detalles.Campos)
		}
	}
}

func TestCrearPedido_JSONRoto_400_NoUn500(t *testing.T) {
	h, _ := armar(10)
	if rec := post(h, `{"items": [`); rec.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, vino %d", rec.Code)
	}
}

func TestCrearPedido_StockInsuficiente_409EnumerandoQueFalta(t *testing.T) {
	h, id := armar(3)
	rec := post(h, cuerpoValido(id, 9, ""))

	if rec.Code != http.StatusConflict {
		t.Fatalf("esperaba 409 (no 500), vino %d: %s", rec.Code, rec.Body)
	}
	var r struct {
		Detalles map[string]int `json:"detalles"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	if r.Detalles[id.String()] != 3 {
		t.Errorf("el 409 debe enumerar variante → disponible: %v", r.Detalles)
	}
}

func TestCrearPedido_VarianteInexistente_404(t *testing.T) {
	h, _ := armar(10)
	if rec := post(h, cuerpoValido(uuid.New(), 1, "")); rec.Code != http.StatusNotFound {
		t.Fatalf("esperaba 404, vino %d: %s", rec.Code, rec.Body)
	}
}

func TestCrearPedido_EnvioSinConfiguracion_409OfreciendoRetiro(t *testing.T) {
	h, id := armar(10)
	body := strings.Replace(cuerpoValido(id, 1, `, "domicilio": "San Martín 123", "distanciaMetros": 800`),
		`"modoEntrega": "RETIRO"`, `"modoEntrega": "ENVIO"`, 1)
	rec := post(h, body)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "retiro") {
		t.Fatalf("esperaba 409 que ofrezca retiro, vino %d: %s", rec.Code, rec.Body)
	}
}
