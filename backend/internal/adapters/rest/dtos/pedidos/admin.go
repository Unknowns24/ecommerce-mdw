package pedidos

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

const maxMotivo = 500

// zonaComercial es la zona horaria del negocio (Argentina, UTC-3): "los pedidos
// del 6 de octubre" son los del día en Rosario, no los del día en UTC.
var zonaComercial = time.FixedZone("ART", -3*60*60)

// CompradorResponse son los datos de contacto del comprador. Solo los ve la
// administración; la vista pública del invitado no los incluye.
type CompradorResponse struct {
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Correo   string `json:"correo"`
	Telefono string `json:"telefono"`
	DNI      string `json:"dni"`
}

// PedidoAdminResponse es el pedido completo que ve quien administra: el
// detalle de siempre más comprador, pago y datos de cancelación. No incluye el
// token del enlace del invitado: la administración no lo necesita.
type PedidoAdminResponse struct {
	PedidoDetalleResponse
	Comprador         CompradorResponse `json:"comprador"`
	UsuarioID         *uuid.UUID        `json:"usuarioId,omitempty"`
	CanceladoPor      *uuid.UUID        `json:"canceladoPor,omitempty"`
	CanceladoEn       *time.Time        `json:"canceladoEn,omitempty"`
	MotivoCancelacion *string           `json:"motivoCancelacion,omitempty"`
}

func APedidoAdmin(p domain.Pedido) PedidoAdminResponse {
	return PedidoAdminResponse{
		PedidoDetalleResponse: APedidoDetalle(p),
		Comprador: CompradorResponse{
			Nombre: p.CompradorNombre, Apellido: p.CompradorApellido, Correo: p.CompradorCorreo,
			Telefono: p.CompradorTelefono, DNI: p.CompradorDNI,
		},
		UsuarioID: p.UsuarioID, CanceladoPor: p.CanceladoPor, CanceladoEn: p.CanceladoEn, MotivoCancelacion: p.MotivoCancelacion,
	}
}

// PedidoAdminResumenResponse es una fila del listado administrativo.
type PedidoAdminResumenResponse struct {
	PedidoResumenResponse
	CompradorNombre   string `json:"compradorNombre"`
	CompradorApellido string `json:"compradorApellido"`
}

type PaginaAdminResponse struct {
	Items  []PedidoAdminResumenResponse `json:"items"`
	Total  int64                        `json:"total"`
	Pagina int                          `json:"pagina"`
	Tamano int                          `json:"tamano"`
}

func APaginaAdmin(p uc.PaginaPedidos) PaginaAdminResponse {
	items := make([]PedidoAdminResumenResponse, len(p.Pedidos))
	for i, x := range p.Pedidos {
		items[i] = PedidoAdminResumenResponse{
			PedidoResumenResponse: PedidoResumenResponse{
				ID: x.ID, Numero: x.Numero, ModoEntrega: string(x.ModoEntrega), MedioPago: string(x.MedioPago),
				EstadoPedido: string(x.EstadoPedido), EstadoPago: string(x.EstadoPago),
				TotalCentavos: x.TotalCentavos, CreadoEn: x.CreadoEn,
			},
			CompradorNombre: x.CompradorNombre, CompradorApellido: x.CompradorApellido,
		}
	}
	return PaginaAdminResponse{Items: items, Total: p.Total, Pagina: p.Pagina, Tamano: p.Tamano}
}

// FiltrosDesdeQuery arma los filtros del listado administrativo a partir de
// ?estado=CONFIRMADO&desde=2026-10-01&hasta=2026-10-31 (fechas del día
// comercial; "hasta" incluye ese día completo). Un valor inválido es un 400 con
// el nombre del campo.
func FiltrosDesdeQuery(estado, desde, hasta string) (out.FiltrosPedidos, error) {
	var f out.FiltrosPedidos
	campos := map[string]string{}

	if estado != "" {
		e := domain.EstadoPedido(estado)
		switch e {
		case domain.PedidoPendienteDePago, domain.PedidoConfirmado, domain.PedidoDespachado, domain.PedidoCompletado, domain.PedidoCancelado:
			f.Estado = &e
		default:
			campos["estado"] = "Usá PENDIENTE_DE_PAGO, CONFIRMADO, DESPACHADO, COMPLETADO o CANCELADO"
		}
	}
	if desde != "" {
		if t, err := time.ParseInLocation("2006-01-02", desde, zonaComercial); err != nil {
			campos["desde"] = "Usá el formato AAAA-MM-DD"
		} else {
			f.Desde = &t
		}
	}
	if hasta != "" {
		if t, err := time.ParseInLocation("2006-01-02", hasta, zonaComercial); err != nil {
			campos["hasta"] = "Usá el formato AAAA-MM-DD"
		} else {
			fin := t.AddDate(0, 0, 1) // excluyente: el día "hasta" entra completo
			f.Hasta = &fin
		}
	}
	if f.Desde != nil && f.Hasta != nil && !f.Desde.Before(*f.Hasta) {
		campos["hasta"] = "No puede ser anterior a desde"
	}

	if len(campos) > 0 {
		return out.FiltrosPedidos{}, apierr.ErrValidacion{Campos: campos}
	}
	return f, nil
}

// CancelarRequest es el body de POST /api/admin/pedidos/{id}/cancelacion. Solo
// trae el motivo: quién cancela sale de la sesión, nunca del body.
type CancelarRequest struct {
	Motivo string `json:"motivo"`
}

// AMotivo valida y devuelve el motivo (obligatorio, hasta 500 caracteres).
func (r CancelarRequest) AMotivo() (string, error) {
	m := strings.TrimSpace(r.Motivo)
	if m == "" || utf8.RuneCountInString(m) > maxMotivo {
		return "", apierr.ErrValidacion{Campos: map[string]string{"motivo": "Indicá el motivo de la cancelación (hasta 500 caracteres)"}}
	}
	return m, nil
}

// CancelacionResponse es la respuesta de cancelar. Aviso solo aparece si el
// pedido estaba pagado: el sistema no devuelve plata, el reintegro se hace a
// mano en Mercado Pago.
type CancelacionResponse struct {
	Error  *string             `json:"error"`
	Aviso  string              `json:"aviso,omitempty"`
	Pedido PedidoAdminResponse `json:"pedido"`
}

func ACancelacion(r uc.ResultadoCancelacion) CancelacionResponse {
	return CancelacionResponse{Aviso: r.Aviso, Pedido: APedidoAdmin(r.Pedido)}
}
