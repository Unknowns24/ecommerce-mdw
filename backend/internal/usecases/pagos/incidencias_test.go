package pagos

import (
	"strings"
	"testing"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

func TestNotaDeResolucionObligatoriaYAcotada(t *testing.T) {
	for _, nota := range []string{"", " \n\t", strings.Repeat("á", 1001)} {
		if _, err := ValidarNotaResolucion(nota); err == nil {
			t.Fatalf("debe rechazar la nota de %d caracteres", len([]rune(nota)))
		}
	}
	if nota, err := ValidarNotaResolucion("  Reintegro verificado  "); err != nil || nota != "Reintegro verificado" {
		t.Fatalf("nota normalizada = %q, error = %v", nota, err)
	}
	if _, err := ValidarNotaResolucion(strings.Repeat("á", 1000)); err != nil {
		t.Fatal(err)
	}
}

func TestSoloUnaAprobacionAnomalaGeneraIncidencia(t *testing.T) {
	p := domain.Pedido{EstadoPedido: domain.PedidoPendienteDePago, TotalCentavos: 150000}
	if motivo := MotivoIncidencia(DecidirConfirmacion(p, 100000)); motivo != "MONTO_DISTINTO" {
		t.Fatal(motivo)
	}
	if motivo := MotivoIncidencia(DecidirConfirmacion(p, 150000)); motivo != "" {
		t.Fatal(motivo)
	}
	p.EstadoPedido = domain.PedidoCancelado
	if motivo := MotivoIncidencia(DecidirConfirmacion(p, 150000)); motivo != "PEDIDO_NO_PENDIENTE" {
		t.Fatal(motivo)
	}
}
