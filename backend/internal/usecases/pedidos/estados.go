package pedidos

import (
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

// Máquina de estados del pedido. Cada función responde una sola pregunta:
// "¿esta operación está permitida sobre este pedido?". Devuelve nil o un
// ErrRegla (→ 409) con un mensaje que dice qué hacer, no solo qué pasó.
// Las operaciones son sub-recursos (POST .../despacho), no un PATCH con el
// estado: el cliente pide que ocurra algo y el servidor decide si corresponde.

func regla(mensaje string) error { return apierr.ErrRegla{Mensaje: mensaje} }

// PuedeDespachar: solo un envío confirmado con el pago aprobado.
func PuedeDespachar(p domain.Pedido) error {
	switch {
	case p.EstadoPedido == domain.PedidoCancelado:
		return regla("No se puede despachar un pedido cancelado.")
	case p.ModoEntrega != domain.EntregaEnvio:
		return regla("Un pedido con retiro en el local no se despacha: completá la entrega cuando el comprador lo retire.")
	case p.EstadoPedido == domain.PedidoDespachado || p.EstadoPedido == domain.PedidoCompletado:
		return regla("Este pedido ya fue despachado.")
	case p.EstadoPedido != domain.PedidoConfirmado || p.EstadoPago != domain.PagoAprobado:
		return regla("No se puede despachar un pedido que todavía no fue pagado. Esperá la aprobación del pago.")
	}
	return nil
}

// PuedeEntregar: un envío se entrega desde DESPACHADO; un retiro, desde
// CONFIRMADO. Una compra web pendiente de pago nunca se completa, y un retiro
// en efectivo se completa con el cobro (PuedeCompletarEnEfectivo).
func PuedeEntregar(p domain.Pedido) error {
	switch {
	case p.EstadoPedido == domain.PedidoCancelado:
		return regla("No se puede entregar un pedido cancelado.")
	case p.EstadoPedido == domain.PedidoCompletado:
		return regla("Este pedido ya fue entregado.")
	case p.EstadoPedido == domain.PedidoPendienteDePago:
		return regla("No se puede completar un pedido que todavía no fue pagado en Mercado Pago.")
	}

	if p.ModoEntrega == domain.EntregaEnvio {
		if p.EstadoPedido != domain.PedidoDespachado {
			return regla("Primero hay que despachar el pedido: un envío se entrega recién después del despacho.")
		}
		return nil
	}

	// Retiro.
	if p.MedioPago == domain.PagoEfectivo {
		return regla("Un pedido en efectivo se completa registrando el cobro: usá el cobro en efectivo.")
	}
	if p.EstadoPedido != domain.PedidoConfirmado || p.EstadoPago != domain.PagoAprobado {
		return regla("No se puede entregar un retiro que todavía no fue pagado.")
	}
	return nil
}

// PuedeCancelar: solo desde PENDIENTE_DE_PAGO o CONFIRMADO. Nunca un pedido
// despachado o completado, y cancelar de nuevo uno cancelado se rechaza: eso
// es lo que impide devolver el stock dos veces.
func PuedeCancelar(p domain.Pedido) error {
	switch p.EstadoPedido {
	case domain.PedidoPendienteDePago, domain.PedidoConfirmado:
		return nil
	case domain.PedidoCancelado:
		return regla("Este pedido ya está cancelado.")
	case domain.PedidoDespachado:
		return regla("No se puede cancelar un pedido que ya fue despachado.")
	case domain.PedidoCompletado:
		return regla("No se puede cancelar un pedido que ya fue entregado.")
	}
	return regla("El pedido está en un estado desconocido y no se puede cancelar.")
}

// PuedeCompletarEnEfectivo: retiro en efectivo, confirmado, con el cobro
// todavía pendiente. La acción registra cobro y entrega en una sola vez.
func PuedeCompletarEnEfectivo(p domain.Pedido) error {
	switch {
	case p.EstadoPedido == domain.PedidoCancelado:
		return regla("No se puede cobrar un pedido cancelado.")
	case p.EstadoPedido == domain.PedidoCompletado:
		return regla("Este pedido ya fue completado.")
	case p.MedioPago != domain.PagoEfectivo || p.ModoEntrega != domain.EntregaRetiro:
		return regla("Solo se cobra en efectivo un pedido con retiro en el local pagado en efectivo.")
	case p.EstadoPedido != domain.PedidoConfirmado:
		return regla("El pedido no está confirmado: no se puede registrar el cobro.")
	case p.EstadoPago != domain.PagoPendiente:
		return regla("El cobro de este pedido ya fue registrado.")
	}
	return nil
}
