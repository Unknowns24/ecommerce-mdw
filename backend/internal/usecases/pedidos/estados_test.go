package pedidos

import (
	"testing"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

func pedido(modo domain.ModoEntrega, estado domain.EstadoPedido, pago domain.EstadoPago, medio domain.MedioPago) domain.Pedido {
	return domain.Pedido{ModoEntrega: modo, EstadoPedido: estado, EstadoPago: pago, MedioPago: medio}
}

type caso struct {
	nombre string
	p      domain.Pedido
	ok     bool
}

func verificar(t *testing.T, regla func(domain.Pedido) error, casos []caso) {
	t.Helper()
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := regla(c.p)
			switch {
			case c.ok && err != nil:
				t.Fatalf("debía permitirse y falló: %v", err)
			case !c.ok && err == nil:
				t.Fatal("debía rechazarse y se permitió")
			case !c.ok && !esRegla(err):
				t.Fatalf("debía ser ErrRegla (409), vino %T: %v", err, err)
			}
		})
	}
}

func TestPuedeDespachar(t *testing.T) {
	verificar(t, PuedeDespachar, []caso{
		{"envío confirmado y pagado", pedido(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago), true},
		{"envío todavía no pagado (el 409 de la demo)", pedido(domain.EntregaEnvio, domain.PedidoPendienteDePago, domain.PagoPendiente, domain.PagoMercadoPago), false},
		{"retiro no se despacha", pedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago), false},
		{"ya despachado", pedido(domain.EntregaEnvio, domain.PedidoDespachado, domain.PagoAprobado, domain.PagoMercadoPago), false},
		{"cancelado", pedido(domain.EntregaEnvio, domain.PedidoCancelado, domain.PagoAprobado, domain.PagoMercadoPago), false},
		{"confirmado pero pago rechazado", pedido(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoRechazado, domain.PagoMercadoPago), false},
	})
}

func TestPuedeDespachar_MensajeDiceQueHacer(t *testing.T) {
	err := PuedeDespachar(pedido(domain.EntregaEnvio, domain.PedidoPendienteDePago, domain.PagoPendiente, domain.PagoMercadoPago))
	if err == nil || err.Error() != "No se puede despachar un pedido que todavía no fue pagado. Esperá la aprobación del pago." {
		t.Fatalf("mensaje inesperado: %v", err)
	}
}

func TestPuedeEntregar(t *testing.T) {
	verificar(t, PuedeEntregar, []caso{
		{"envío despachado", pedido(domain.EntregaEnvio, domain.PedidoDespachado, domain.PagoAprobado, domain.PagoMercadoPago), true},
		{"envío sin despachar", pedido(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago), false},
		{"retiro MP confirmado y pagado", pedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago), true},
		{"compra web pendiente de pago no se completa", pedido(domain.EntregaRetiro, domain.PedidoPendienteDePago, domain.PagoPendiente, domain.PagoMercadoPago), false},
		{"retiro en efectivo va por el cobro", pedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoPendiente, domain.PagoEfectivo), false},
		{"ya completado", pedido(domain.EntregaEnvio, domain.PedidoCompletado, domain.PagoAprobado, domain.PagoMercadoPago), false},
		{"cancelado", pedido(domain.EntregaRetiro, domain.PedidoCancelado, domain.PagoPendiente, domain.PagoMercadoPago), false},
	})
}

func TestPuedeCancelar(t *testing.T) {
	verificar(t, PuedeCancelar, []caso{
		{"pendiente de pago", pedido(domain.EntregaEnvio, domain.PedidoPendienteDePago, domain.PagoPendiente, domain.PagoMercadoPago), true},
		{"confirmado", pedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago), true},
		{"completado nunca se cancela", pedido(domain.EntregaEnvio, domain.PedidoCompletado, domain.PagoAprobado, domain.PagoMercadoPago), false},
		{"despachado", pedido(domain.EntregaEnvio, domain.PedidoDespachado, domain.PagoAprobado, domain.PagoMercadoPago), false},
		{"cancelar dos veces", pedido(domain.EntregaEnvio, domain.PedidoCancelado, domain.PagoPendiente, domain.PagoMercadoPago), false},
	})
}

func TestPuedeCompletarEnEfectivo(t *testing.T) {
	verificar(t, PuedeCompletarEnEfectivo, []caso{
		{"retiro en efectivo confirmado sin cobrar", pedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoPendiente, domain.PagoEfectivo), true},
		{"ya cobrado", pedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoEfectivo), false},
		{"cancelado", pedido(domain.EntregaRetiro, domain.PedidoCancelado, domain.PagoPendiente, domain.PagoEfectivo), false},
		{"pedido de Mercado Pago", pedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago), false},
		{"completado", pedido(domain.EntregaRetiro, domain.PedidoCompletado, domain.PagoAprobado, domain.PagoEfectivo), false},
	})
}
