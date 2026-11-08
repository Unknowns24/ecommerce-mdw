package pedidos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

// configTiendaFalsa imita al repositorio: sin configuración, Actual devuelve
// ErrNoEncontrado como lo hace la consulta real.
type configTiendaFalsa struct {
	cfg     *domain.ConfiguracionTienda
	tarifas []domain.TarifaDistancia
}

func (c configTiendaFalsa) Actual(context.Context) (domain.ConfiguracionTienda, error) {
	if c.cfg == nil {
		return domain.ConfiguracionTienda{}, apierr.ErrNoEncontrado
	}
	return *c.cfg, nil
}

func (c configTiendaFalsa) TarifasOrdenadas(context.Context) ([]domain.TarifaDistancia, error) {
	return c.tarifas, nil
}

func pedirConfiguracion(f configTiendaFalsa) *httptest.ResponseRecorder {
	h := NuevoHandler(nil, nil, nil, nil, uc.NuevaConfiguracion(f))
	rec := httptest.NewRecorder()
	h.ObtenerConfiguracionTienda(rec, httptest.NewRequest(http.MethodGet, "/api/tienda/configuracion", nil))
	return rec
}

func TestObtenerConfiguracionTienda_200ConLosCamposEnCamelCase(t *testing.T) {
	rec := pedirConfiguracion(configTiendaFalsa{
		cfg: &domain.ConfiguracionTienda{EfectivoHabilitado: true, DistanciaMaximaMetros: 5000},
		tarifas: []domain.TarifaDistancia{
			{DesdeMetros: 0, HastaMetros: 1000, CostoCentavos: 150000},
			{DesdeMetros: 1001, HastaMetros: 5000, CostoCentavos: 400000},
		},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, vino %d: %s", rec.Code, rec.Body)
	}
	var r map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if r["efectivoHabilitado"] != true || r["distanciaMaximaMetros"] != float64(5000) {
		t.Errorf("configuración inesperada: %v", r)
	}
	tarifas, ok := r["tarifas"].([]any)
	if !ok || len(tarifas) != 2 {
		t.Fatalf("esperaba 2 tarifas, vino %v", r["tarifas"])
	}
	segunda := tarifas[1].(map[string]any)
	if segunda["desdeMetros"] != float64(1001) || segunda["hastaMetros"] != float64(5000) || segunda["costoCentavos"] != float64(400000) {
		t.Errorf("tarifa inesperada: %v", segunda)
	}
	// Solo lo que se decidió publicar: nada de id ni fechas de la configuración.
	if len(r) != 3 {
		t.Errorf("esperaba 3 campos, vino %v", r)
	}
}

func TestObtenerConfiguracionTienda_SinTarifasDevuelveListaVaciaNoNull(t *testing.T) {
	rec := pedirConfiguracion(configTiendaFalsa{cfg: &domain.ConfiguracionTienda{DistanciaMaximaMetros: 3000}})

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, vino %d: %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), `"tarifas":[]`) {
		t.Errorf("esperaba tarifas como [], vino %s", rec.Body)
	}
}

func TestObtenerConfiguracionTienda_SinConfiguracionResponde404(t *testing.T) {
	rec := pedirConfiguracion(configTiendaFalsa{})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperaba 404, vino %d: %s", rec.Code, rec.Body)
	}
}
