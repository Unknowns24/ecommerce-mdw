// Package pedidos define la forma exacta de los JSON que entran y salen de los
// endpoints de carrito y pedidos.
package pedidos

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

const (
	maxItems    = 50
	maxUnidades = 1000
)

var reDNI = regexp.MustCompile(`^[0-9]{7,8}$`)

// CrearPedidoRequest es el body de POST /api/pedidos.
//
// Fijate lo que NO tiene: precios, totales, usuarioId ni estados. Si esos
// campos existieran acá habría un agujero: con el body cualquiera crearía un
// pedido pagado de $0 desde Postman. Si el cliente los manda igual, el
// decodificador JSON los ignora: el servidor nunca los lee.
type CrearPedidoRequest struct {
	Items           []ItemRequest    `json:"items"`
	Comprador       CompradorRequest `json:"comprador"`
	ModoEntrega     string           `json:"modoEntrega"`
	Domicilio       string           `json:"domicilio"`
	DistanciaMetros *int             `json:"distanciaMetros"`
	MedioPago       string           `json:"medioPago"`
}

type ItemRequest struct {
	VarianteID string `json:"varianteId"`
	Unidades   int    `json:"unidades"`
}

type CompradorRequest struct {
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Correo   string `json:"correo"`
	Telefono string `json:"telefono"`
	DNI      string `json:"dni"`
}

// AEntrada valida la FORMA del body (lo que se decide mirándolo solo) y lo
// convierte en la entrada del caso de uso. Si algo está mal devuelve
// ErrValidacion con un mensaje por campo (→ 400). Las reglas que necesitan
// mirar el estado del sistema (stock, cobertura) son del caso de uso (→ 409).
func (r CrearPedidoRequest) AEntrada() (uc.DatosPedido, error) {
	campos := map[string]string{}

	items := make([]uc.ItemPedido, 0, len(r.Items))
	switch {
	case len(r.Items) == 0:
		campos["items"] = "Agregá al menos un producto"
	case len(r.Items) > maxItems:
		campos["items"] = fmt.Sprintf("No se pueden pedir más de %d productos distintos", maxItems)
	default:
		for i, it := range r.Items {
			id, err := uuid.Parse(it.VarianteID)
			if err != nil {
				campos[fmt.Sprintf("items[%d].varianteId", i)] = "Identificador de producto inválido"
				continue
			}
			if it.Unidades < 1 || it.Unidades > maxUnidades {
				campos[fmt.Sprintf("items[%d].unidades", i)] = fmt.Sprintf("Las unidades deben ser un entero entre 1 y %d", maxUnidades)
				continue
			}
			items = append(items, uc.ItemPedido{VarianteID: id, Unidades: it.Unidades})
		}
	}

	c := r.Comprador
	nombre, apellido := strings.TrimSpace(c.Nombre), strings.TrimSpace(c.Apellido)
	correo, telefono, dni := strings.TrimSpace(c.Correo), strings.TrimSpace(c.Telefono), strings.TrimSpace(c.DNI)
	if nombre == "" || len(nombre) > 100 {
		campos["comprador.nombre"] = "Ingresá tu nombre (hasta 100 caracteres)"
	}
	if apellido == "" || len(apellido) > 100 {
		campos["comprador.apellido"] = "Ingresá tu apellido (hasta 100 caracteres)"
	}
	if dir, err := mail.ParseAddress(correo); err != nil || dir.Address != correo || len(correo) > 254 {
		campos["comprador.correo"] = "Ingresá un correo válido"
	}
	if telefono == "" || len(telefono) > 30 {
		campos["comprador.telefono"] = "Ingresá un teléfono (hasta 30 caracteres)"
	}
	if !reDNI.MatchString(dni) {
		campos["comprador.dni"] = "El DNI debe tener 7 u 8 dígitos, sin puntos"
	}

	modo := domain.ModoEntrega(r.ModoEntrega)
	var domicilio *string
	var distancia *int
	switch modo {
	case domain.EntregaRetiro:
		// El retiro no lleva domicilio ni distancia: se ignoran si llegan.
	case domain.EntregaEnvio:
		d := strings.TrimSpace(r.Domicilio)
		if d == "" || len(d) > 300 {
			campos["domicilio"] = "Ingresá el domicilio de entrega (hasta 300 caracteres)"
		} else {
			domicilio = &d
		}
		if r.DistanciaMetros == nil || *r.DistanciaMetros < 0 {
			campos["distanciaMetros"] = "Indicá la distancia en metros (cero o más)"
		} else {
			distancia = r.DistanciaMetros
		}
	default:
		campos["modoEntrega"] = "Elegí RETIRO o ENVIO"
	}

	medio := domain.MedioPago(r.MedioPago)
	if medio != domain.PagoMercadoPago && medio != domain.PagoEfectivo {
		campos["medioPago"] = "Elegí MERCADO_PAGO o EFECTIVO"
	}

	if len(campos) > 0 {
		return uc.DatosPedido{}, apierr.ErrValidacion{Campos: campos}
	}
	return uc.DatosPedido{
		Items:           items,
		Comprador:       uc.Comprador{Nombre: nombre, Apellido: apellido, Correo: correo, Telefono: telefono, DNI: dni},
		ModoEntrega:     modo,
		Domicilio:       domicilio,
		DistanciaMetros: distancia,
		MedioPago:       medio,
	}, nil
}

// PedidoResponse es el pedido recién creado, tal como lo ve quien lo compró:
// el detalle de siempre más TokenAcceso, el secreto del enlace privado, que se
// entrega una sola vez, acá. (Al consultar después, el token no vuelve a salir.)
type PedidoResponse struct {
	PedidoDetalleResponse
	TokenAcceso string `json:"tokenAcceso"`
}

// PedidoDetalleResponse es el pedido visto desde la cuenta de quien lo compró
// (o desde el checkout, que le suma el token). Es la vista "propia": más
// completa que la pública, y por eso nunca se usa en el enlace del invitado.
type PedidoDetalleResponse struct {
	ID               uuid.UUID         `json:"id"`
	Numero           int64             `json:"numero"`
	ModoEntrega      string            `json:"modoEntrega"`
	Domicilio        *string           `json:"domicilio,omitempty"`
	DistanciaMetros  *int              `json:"distanciaMetros,omitempty"`
	EnvioCentavos    int64             `json:"envioCentavos"`
	SubtotalCentavos int64             `json:"subtotalCentavos"`
	TotalCentavos    int64             `json:"totalCentavos"`
	EstadoPedido     string            `json:"estadoPedido"`
	EstadoPago       string            `json:"estadoPago"`
	MedioPago        string            `json:"medioPago"`
	VenceEn          *time.Time        `json:"venceEn,omitempty"`
	CreadoEn         time.Time         `json:"creadoEn"`
	Detalles         []DetalleResponse `json:"detalles"`
}

type DetalleResponse struct {
	VarianteID             uuid.UUID `json:"varianteId"`
	Codigo                 string    `json:"codigo"`
	Nombre                 string    `json:"nombre"`
	Unidades               int       `json:"unidades"`
	PrecioUnitarioCentavos int64     `json:"precioUnitarioCentavos"`
	SubtotalCentavos       int64     `json:"subtotalCentavos"`
}

// APedidoResponse convierte el pedido recién creado en su respuesta JSON,
// con el token del enlace privado.
func APedidoResponse(p domain.Pedido) PedidoResponse {
	return PedidoResponse{PedidoDetalleResponse: APedidoDetalle(p), TokenAcceso: p.TokenAcceso}
}

// APedidoDetalle convierte el pedido del dominio en la vista propia (sin token).
func APedidoDetalle(p domain.Pedido) PedidoDetalleResponse {
	detalles := make([]DetalleResponse, len(p.Detalles))
	for i, d := range p.Detalles {
		detalles[i] = DetalleResponse{
			VarianteID: d.VarianteID, Codigo: d.Codigo, Nombre: d.Nombre, Unidades: d.Unidades,
			PrecioUnitarioCentavos: d.PrecioUnitarioCentavos, SubtotalCentavos: d.SubtotalCentavos,
		}
	}
	return PedidoDetalleResponse{
		ID: p.ID, Numero: p.Numero,
		ModoEntrega: string(p.ModoEntrega), Domicilio: p.Domicilio, DistanciaMetros: p.DistanciaMetros,
		EnvioCentavos: p.EnvioCentavos, SubtotalCentavos: p.SubtotalCentavos, TotalCentavos: p.TotalCentavos,
		EstadoPedido: string(p.EstadoPedido), EstadoPago: string(p.EstadoPago), MedioPago: string(p.MedioPago),
		VenceEn: p.VenceEn, CreadoEn: p.CreadoEn, Detalles: detalles,
	}
}
