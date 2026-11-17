package inventario

import (
	"testing"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

func TestCancelarRevierteSoloLaSalidaRegistrada(t *testing.T) {
	for _, c := range []struct {
		estado   domain.EstadoReserva
		esperado int
	}{
		{domain.ReservaActiva, 0}, {domain.ReservaConsumida, 3}, {domain.ReservaLiberada, 0},
	} {
		if got := unidadesARestituir(c.estado, 3); got != c.esperado {
			t.Fatalf("cancelar reserva %s debe restituir %d, obtuvo %d", c.estado, c.esperado, got)
		}
	}
}
