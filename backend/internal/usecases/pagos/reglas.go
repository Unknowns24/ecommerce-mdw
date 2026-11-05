// Package pagos holds the payment rules of SPEC-H14 as pure functions: no
// base, no red, la hora entra por parámetro. El adaptador REST lee el pedido
// y le pregunta a Mercado Pago; estas funciones deciden.
package pagos

import (
	"crypto/subtle"
	"time"

	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

// PuedeIniciarPago decide si quien llama puede abrir un pago para el pedido.
// Es del comprador si la sesión es la del dueño del pedido o si trae el token
// del enlace privado; si no, el pedido no existe para él (404, no 403).
func PuedeIniciarPago(p domain.Pedido, usuarioID *uuid.UUID, token string, ahora time.Time) error {
	if !esDelComprador(p, usuarioID, token) {
		return apierr.ErrNoEncontrado
	}
	if p.MedioPago != domain.PagoMercadoPago || p.EstadoPedido != domain.PedidoPendienteDePago || p.EstadoPago != domain.PagoPendiente {
		return apierr.ErrRegla{Mensaje: "Este pedido no está pendiente de pago con Mercado Pago"}
	}
	if Vencido(p, ahora) {
		return apierr.ErrRegla{Mensaje: "La reserva de este pedido venció. Volvé a armar el carrito para comprar."}
	}
	return nil
}

func esDelComprador(p domain.Pedido, usuarioID *uuid.UUID, token string) bool {
	if usuarioID != nil && p.UsuarioID != nil && *p.UsuarioID == *usuarioID {
		return true
	}
	return token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(p.TokenAcceso)) == 1
}

// Vencido dice si la reserva de 15 minutos de un pedido de Mercado Pago ya
// se agotó. El plazo es el original de la creación: reintentar no lo extiende.
func Vencido(p domain.Pedido, ahora time.Time) bool {
	return p.VenceEn != nil && !ahora.Before(*p.VenceEn)
}

// estadosActivos son los estados de Mercado Pago en los que el dinero ya se
// cobró o todavía se puede cobrar. Con uno de estos no se abre otro pago ni
// se da por vencido el pedido.
var estadosActivos = map[string]bool{
	"pending":      true,
	"in_process":   true,
	"authorized":   true,
	"in_mediation": true,
	"approved":     true,
}

// HayPagoActivo dice si alguno de los pagos que Mercado Pago tiene para el
// pedido puede terminar en un cobro (H14: no abrir otro pago que pueda
// producir un cobro duplicado).
func HayPagoActivo(estados []string) bool {
	for _, e := range estados {
		if estadosActivos[e] {
			return true
		}
	}
	return false
}

// Confirmacion es lo que hay que hacer con un pago aprobado.
type Confirmacion int

const (
	Confirmar Confirmacion = iota
	// NoConfirmarPedidoNoPendiente: el pago llegó cuando el pedido ya no
	// esperaba pago (vencido o cancelado). Se registra el pago y se deja la
	// incidencia para la administración (spec, sección 6, «Propuesta de
	// contingencia»): el stock ya pudo haberse liberado.
	NoConfirmarPedidoNoPendiente
	// NoConfirmarMontoDistinto: lo acreditado no es el total del pedido.
	NoConfirmarMontoDistinto
)

// DecidirConfirmacion decide si un pago ya aprobado por Mercado Pago confirma
// el pedido. Solo lo confirma si el pedido todavía espera ese pago y si se
// acreditó exactamente su total.
func DecidirConfirmacion(p domain.Pedido, montoAcreditadoCentavos int64) Confirmacion {
	if p.EstadoPedido != domain.PedidoPendienteDePago {
		return NoConfirmarPedidoNoPendiente
	}
	if montoAcreditadoCentavos != p.TotalCentavos {
		return NoConfirmarMontoDistinto
	}
	return Confirmar
}

// VencimientoDePreferencia dice hasta cuándo se puede pagar la preferencia de
// Mercado Pago del pedido: el mismo instante en que vence su reserva. Así no
// se puede pagar un pedido ya cancelado, con su stock liberado. Sin vence_en
// (efectivo) no hay plazo. Propuesta del Issue #14 (spec, sección 6,
// «Pendiente crítico de integración»): no se mergea sin la decisión del
// equipo.
func VencimientoDePreferencia(p domain.Pedido) (time.Time, bool) {
	if p.VenceEn == nil {
		return time.Time{}, false
	}
	return *p.VenceEn, true
}

// AccionAlVencer es lo que el vencimiento automático hace con un pedido cuya
// reserva se agotó, según lo que Mercado Pago informa de sus pagos.
type AccionAlVencer int

const (
	// Cancelar: ningún pago se cobró ni puede cobrarse; se cancela y se
	// libera el stock.
	Cancelar AccionAlVencer = iota
	// Esperar: hay un pago que todavía puede acreditarse; se revisa en la
	// próxima vuelta.
	Esperar
	// ConciliarAprobado: Mercado Pago tiene un pago aprobado cuyo webhook
	// no llegó; se aplica DecidirConfirmacion, igual que en el webhook.
	ConciliarAprobado
)

// DecidirAlVencer decide qué hacer con un pedido vencido a partir de los
// estados de sus pagos en Mercado Pago. Un pago aprobado manda sobre todo lo
// demás: el dinero ya se cobró y el pedido no se puede cancelar en silencio.
// Propuesta del Issue #14.
func DecidirAlVencer(estados []string) AccionAlVencer {
	for _, e := range estados {
		if e == "approved" {
			return ConciliarAprobado
		}
	}
	if HayPagoActivo(estados) {
		return Esperar
	}
	return Cancelar
}
