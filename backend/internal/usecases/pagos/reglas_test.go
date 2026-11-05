package pagos

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

var ahora = time.Date(2026, 11, 5, 20, 0, 0, 0, time.UTC)

func pedidoPendiente(usuarioID *uuid.UUID) domain.Pedido {
	vence := ahora.Add(10 * time.Minute)
	return domain.Pedido{
		ID:            uuid.New(),
		UsuarioID:     usuarioID,
		TokenAcceso:   "token-del-enlace-privado",
		MedioPago:     domain.PagoMercadoPago,
		EstadoPedido:  domain.PedidoPendienteDePago,
		EstadoPago:    domain.PagoPendiente,
		VenceEn:       &vence,
		TotalCentavos: 1_650_000,
	}
}

func TestPuedeIniciarPago(t *testing.T) {
	duena := uuid.New()
	otro := uuid.New()

	casos := []struct {
		nombre   string
		pedido   func() domain.Pedido
		usuario  *uuid.UUID
		token    string
		ahora    time.Time
		esperado error // nil, ErrNoEncontrado o ErrRegla
	}{
		{"la dueña con su sesión", func() domain.Pedido { return pedidoPendiente(&duena) }, &duena, "", ahora, nil},
		{"el invitado con su token", func() domain.Pedido { return pedidoPendiente(nil) }, nil, "token-del-enlace-privado", ahora, nil},
		{"otra sesión: no existe para ella", func() domain.Pedido { return pedidoPendiente(&duena) }, &otro, "", ahora, apierr.ErrNoEncontrado},
		{"sin sesión ni token", func() domain.Pedido { return pedidoPendiente(nil) }, nil, "", ahora, apierr.ErrNoEncontrado},
		{"token equivocado", func() domain.Pedido { return pedidoPendiente(nil) }, nil, "otro-token", ahora, apierr.ErrNoEncontrado},
		{"una sesión no habilita el pedido de un invitado", func() domain.Pedido { return pedidoPendiente(nil) }, &otro, "", ahora, apierr.ErrNoEncontrado},
		{"reserva vencida justo en el límite", func() domain.Pedido { return pedidoPendiente(&duena) }, &duena, "", ahora.Add(10 * time.Minute), apierr.ErrRegla{}},
		{"pedido cancelado con el pago todavía pendiente", func() domain.Pedido {
			p := pedidoPendiente(&duena)
			p.EstadoPedido = domain.PedidoCancelado
			return p
		}, &duena, "", ahora, apierr.ErrRegla{}},
		{"pedido ya pagado", func() domain.Pedido {
			p := pedidoPendiente(&duena)
			p.EstadoPedido, p.EstadoPago = domain.PedidoConfirmado, domain.PagoAprobado
			return p
		}, &duena, "", ahora, apierr.ErrRegla{}},
		{"pedido en efectivo", func() domain.Pedido {
			p := pedidoPendiente(&duena)
			p.MedioPago, p.VenceEn = domain.PagoEfectivo, nil
			return p
		}, &duena, "", ahora, apierr.ErrRegla{}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := PuedeIniciarPago(c.pedido(), c.usuario, c.token, c.ahora)
			switch c.esperado.(type) {
			case nil:
				if err != nil {
					t.Fatalf("error = %v, esperado nil", err)
				}
			case apierr.ErrRegla:
				var regla apierr.ErrRegla
				if !errors.As(err, &regla) {
					t.Fatalf("error = %v, esperado un rechazo de regla (409)", err)
				}
			default:
				if !errors.Is(err, c.esperado) {
					t.Fatalf("error = %v, esperado %v", err, c.esperado)
				}
			}
		})
	}
}

func TestVencidoSinPlazoNuncaVence(t *testing.T) {
	p := pedidoPendiente(nil)
	p.VenceEn = nil
	if Vencido(p, ahora.Add(24*time.Hour)) {
		t.Fatal("un pedido sin vence_en (efectivo) no vence automáticamente")
	}
}

func TestHayPagoActivo(t *testing.T) {
	casos := []struct {
		estados  []string
		esperado bool
	}{
		{nil, false},
		{[]string{"rejected", "cancelled"}, false},
		{[]string{"refunded", "charged_back"}, false},
		{[]string{"rejected", "in_process"}, true},
		{[]string{"pending"}, true},
		{[]string{"authorized"}, true},
		{[]string{"approved"}, true},
	}
	for _, c := range casos {
		if got := HayPagoActivo(c.estados); got != c.esperado {
			t.Errorf("HayPagoActivo(%v) = %v, esperado %v", c.estados, got, c.esperado)
		}
	}
}

func TestDecidirConfirmacion(t *testing.T) {
	pendiente := pedidoPendiente(nil)
	cancelado := pedidoPendiente(nil)
	cancelado.EstadoPedido = domain.PedidoCancelado

	casos := []struct {
		nombre   string
		pedido   domain.Pedido
		monto    int64
		esperado Confirmacion
	}{
		{"monto exacto", pendiente, 1_650_000, Confirmar},
		{"cobrado sin el envío", pendiente, 1_500_000, NoConfirmarMontoDistinto},
		{"cobrado de más", pendiente, 1_650_001, NoConfirmarMontoDistinto},
		{"llega después del vencimiento", cancelado, 1_650_000, NoConfirmarPedidoNoPendiente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := DecidirConfirmacion(c.pedido, c.monto); got != c.esperado {
				t.Fatalf("DecidirConfirmacion = %v, esperado %v", got, c.esperado)
			}
		})
	}
}

// Propuesta del Issue #14: la preferencia vence junto con la reserva.
func TestVencimientoDePreferencia(t *testing.T) {
	p := pedidoPendiente(nil)
	vence, ok := VencimientoDePreferencia(p)
	if !ok || !vence.Equal(*p.VenceEn) {
		t.Fatalf("VencimientoDePreferencia = (%v, %v), esperado (%v, true)", vence, ok, *p.VenceEn)
	}

	p.MedioPago, p.VenceEn = domain.PagoEfectivo, nil
	if _, ok := VencimientoDePreferencia(p); ok {
		t.Fatal("un pedido sin vence_en (efectivo) no lleva plazo de preferencia")
	}
}

// Propuesta del Issue #14: qué hace el vencimiento según los pagos que
// informa Mercado Pago.
func TestDecidirAlVencer(t *testing.T) {
	casos := []struct {
		nombre   string
		estados  []string
		esperado AccionAlVencer
	}{
		{"sin pagos: se cancela", nil, Cancelar},
		{"solo rechazados: se cancela", []string{"rejected", "cancelled"}, Cancelar},
		{"reintegrado: se cancela", []string{"refunded"}, Cancelar},
		{"en proceso: se espera", []string{"in_process"}, Esperar},
		{"pendiente después de un rechazo: se espera", []string{"rejected", "pending"}, Esperar},
		{"aprobado sin webhook: se concilia", []string{"approved"}, ConciliarAprobado},
		{"aprobado manda sobre un pendiente", []string{"pending", "approved"}, ConciliarAprobado},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := DecidirAlVencer(c.estados); got != c.esperado {
				t.Fatalf("DecidirAlVencer(%v) = %v, esperado %v", c.estados, got, c.esperado)
			}
		})
	}
}
