package pedidos

import (
	"errors"
	"testing"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

// Tarifas de prueba: [0,1000) $500 · [1000,3000) $900 · [3000,5000) $1.500
var tarifasPrueba = []Tarifa{
	{DesdeMetros: 0, HastaMetros: 1000, CostoCentavos: 50000},
	{DesdeMetros: 1000, HastaMetros: 3000, CostoCentavos: 90000},
	{DesdeMetros: 3000, HastaMetros: 5000, CostoCentavos: 150000},
}

const maximoPrueba = 5000

func TestCotizarEnvio_CostoPorRango(t *testing.T) {
	casos := []struct {
		nombre    string
		distancia int
		costo     int64
	}{
		{"dentro del primer rango", 500, 50000},
		{"dentro del segundo rango", 2000, 90000},
		{"dentro del último rango", 4000, 150000},
		{"el metro 0", 0, 50000},
		{"999 m: todavía primer rango", 999, 50000},
		{"1000 m exactos: es del segundo rango, no del primero", 1000, 90000},
		{"2999 m", 2999, 90000},
		{"3000 m exactos: es del tercer rango", 3000, 150000},
		{"4999 m", 4999, 150000},
		{"5000 m exactos: el máximo SÍ está cubierto", 5000, 150000},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			costo, err := CotizarEnvio(c.distancia, maximoPrueba, tarifasPrueba)
			if err != nil || costo != c.costo {
				t.Fatalf("distancia %d: costo=%d err=%v, esperaba %d", c.distancia, costo, err, c.costo)
			}
		})
	}
}

func TestCotizarEnvio_ExcedeLaCobertura(t *testing.T) {
	costo, err := CotizarEnvio(5001, maximoPrueba, tarifasPrueba) // un metro de más
	if costo != 0 {
		t.Errorf("no se inventa un costo: %d", costo)
	}
	var e apierr.ErrRegla
	if !errors.As(err, &e) {
		t.Fatalf("esperaba ErrRegla, vino %v", err)
	}
	if e.Mensaje != "No hacemos envíos a esa distancia. Podés elegir retiro en el local." {
		t.Errorf("el mensaje debe ofrecer retiro: %q", e.Mensaje)
	}
}

func TestCotizarEnvio_DistanciaNegativa(t *testing.T) {
	if _, err := CotizarEnvio(-1, maximoPrueba, tarifasPrueba); !esValidacion(err) {
		t.Fatalf("esperaba ErrValidacion, vino %v", err)
	}
}

func TestCotizarEnvio_ConfiguracionInconsistente(t *testing.T) {
	conHueco := []Tarifa{{0, 1000, 50000}, {2000, 5000, 90000}} // 1000..1999 sin tarifa
	if _, err := CotizarEnvio(1500, 5000, conHueco); !esRegla(err) {
		t.Errorf("un hueco no debe cotizar un costo inventado: %v", err)
	}

	solapadas := []Tarifa{{0, 2000, 50000}, {1500, 5000, 90000}}
	if _, err := CotizarEnvio(1700, 5000, solapadas); !esRegla(err) {
		t.Errorf("rangos solapados deben rechazarse: %v", err)
	}

	if _, err := CotizarEnvio(100, 5000, nil); !esRegla(err) {
		t.Errorf("sin tarifas no hay envío: %v", err)
	}
}

func TestCotizarEnvio_NoAlteraElSliceDeEntrada(t *testing.T) {
	desordenadas := []Tarifa{{3000, 5000, 150000}, {0, 1000, 50000}, {1000, 3000, 90000}}
	if _, err := CotizarEnvio(500, 5000, desordenadas); err != nil {
		t.Fatal(err)
	}
	if desordenadas[0].DesdeMetros != 3000 {
		t.Error("CotizarEnvio no debe reordenar el slice del que la llama")
	}
}
