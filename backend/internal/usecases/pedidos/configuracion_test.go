package pedidos

import (
	"context"
	"errors"
	"testing"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type lectorConfigFalso struct {
	cfg             domain.ConfiguracionTienda
	errActual       error
	tarifas         []domain.TarifaDistancia
	errTarifas      error
	consultoTarifas bool
}

func (l *lectorConfigFalso) Actual(context.Context) (domain.ConfiguracionTienda, error) {
	return l.cfg, l.errActual
}

func (l *lectorConfigFalso) TarifasOrdenadas(context.Context) ([]domain.TarifaDistancia, error) {
	l.consultoTarifas = true
	return l.tarifas, l.errTarifas
}

func TestConfiguracionVigente_DevuelveConfiguracionYTarifas(t *testing.T) {
	l := &lectorConfigFalso{
		cfg:     domain.ConfiguracionTienda{EfectivoHabilitado: true, DistanciaMaximaMetros: 5000},
		tarifas: []domain.TarifaDistancia{{DesdeMetros: 0, HastaMetros: 1000, CostoCentavos: 150000}},
	}
	got, err := NuevaConfiguracion(l).Vigente(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.Configuracion.EfectivoHabilitado || got.Configuracion.DistanciaMaximaMetros != 5000 || len(got.Tarifas) != 1 {
		t.Errorf("resultado inesperado: %+v", got)
	}
}

func TestConfiguracionVigente_SinConfiguracionEsNoEncontradoYNoMiraLasTarifas(t *testing.T) {
	l := &lectorConfigFalso{errActual: apierr.ErrNoEncontrado}
	_, err := NuevaConfiguracion(l).Vigente(context.Background())
	if !errors.Is(err, apierr.ErrNoEncontrado) {
		t.Fatalf("esperaba ErrNoEncontrado, vino %v", err)
	}
	if l.consultoTarifas {
		t.Error("no debería consultar las tarifas si no hay configuración")
	}
}

func TestConfiguracionVigente_PropagaElErrorDeLasTarifas(t *testing.T) {
	boom := errors.New("base caída")
	l := &lectorConfigFalso{errTarifas: boom}
	if _, err := NuevaConfiguracion(l).Vigente(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("esperaba el error de las tarifas, vino %v", err)
	}
}
