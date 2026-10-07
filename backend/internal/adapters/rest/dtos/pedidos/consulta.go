package pedidos

import (
	"time"

	"github.com/google/uuid"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

// PedidoResumenResponse es una fila del listado "mis pedidos": lo justo para
// mostrar la lista, sin el detalle de productos.
type PedidoResumenResponse struct {
	ID            uuid.UUID `json:"id"`
	Numero        int64     `json:"numero"`
	ModoEntrega   string    `json:"modoEntrega"`
	MedioPago     string    `json:"medioPago"`
	EstadoPedido  string    `json:"estadoPedido"`
	EstadoPago    string    `json:"estadoPago"`
	TotalCentavos int64     `json:"totalCentavos"`
	CreadoEn      time.Time `json:"creadoEn"`
}

// PaginaPedidosResponse es el listado paginado.
type PaginaPedidosResponse struct {
	Items  []PedidoResumenResponse `json:"items"`
	Total  int64                   `json:"total"`
	Pagina int                     `json:"pagina"`
	Tamano int                     `json:"tamano"`
}

func APaginaPedidos(p uc.PaginaPedidos) PaginaPedidosResponse {
	items := make([]PedidoResumenResponse, len(p.Pedidos))
	for i, x := range p.Pedidos {
		items[i] = PedidoResumenResponse{
			ID: x.ID, Numero: x.Numero, ModoEntrega: string(x.ModoEntrega), MedioPago: string(x.MedioPago),
			EstadoPedido: string(x.EstadoPedido), EstadoPago: string(x.EstadoPago),
			TotalCentavos: x.TotalCentavos, CreadoEn: x.CreadoEn,
		}
	}
	return PaginaPedidosResponse{Items: items, Total: p.Total, Pagina: p.Pagina, Tamano: p.Tamano}
}

// PedidoPublicoResponse es lo que sale por el enlace privado del invitado
// (GET /api/pedidos/publico/{token}), que es público a propósito: cualquiera
// que tenga el enlace lo ve. Por eso es un tipo APARTE, con los campos
// enumerados a mano: productos, cantidades, precios, estado y datos de
// entrega. No tiene id, ni usuario, ni correo, ni DNI, ni teléfono del
// comprador. Si alguien le agrega un campo al pedido, ese campo NO se publica
// salvo que alguien lo agregue acá, a propósito.
type PedidoPublicoResponse struct {
	Numero           int64                    `json:"numero"`
	ModoEntrega      string                   `json:"modoEntrega"`
	Domicilio        *string                  `json:"domicilio,omitempty"`
	DistanciaMetros  *int                     `json:"distanciaMetros,omitempty"`
	EnvioCentavos    int64                    `json:"envioCentavos"`
	SubtotalCentavos int64                    `json:"subtotalCentavos"`
	TotalCentavos    int64                    `json:"totalCentavos"`
	EstadoPedido     string                   `json:"estadoPedido"`
	EstadoPago       string                   `json:"estadoPago"`
	MedioPago        string                   `json:"medioPago"`
	VenceEn          *time.Time               `json:"venceEn,omitempty"`
	CreadoEn         time.Time                `json:"creadoEn"`
	Detalles         []DetallePublicoResponse `json:"detalles"`
}

type DetallePublicoResponse struct {
	Codigo                 string `json:"codigo"`
	Nombre                 string `json:"nombre"`
	Unidades               int    `json:"unidades"`
	PrecioUnitarioCentavos int64  `json:"precioUnitarioCentavos"`
	SubtotalCentavos       int64  `json:"subtotalCentavos"`
}

func APedidoPublico(p domain.PedidoPublico) PedidoPublicoResponse {
	detalles := make([]DetallePublicoResponse, len(p.Detalles))
	for i, d := range p.Detalles {
		detalles[i] = DetallePublicoResponse{
			Codigo: d.Codigo, Nombre: d.Nombre, Unidades: d.Unidades,
			PrecioUnitarioCentavos: d.PrecioUnitarioCentavos, SubtotalCentavos: d.SubtotalCentavos,
		}
	}
	return PedidoPublicoResponse{
		Numero: p.Numero, ModoEntrega: string(p.ModoEntrega), Domicilio: p.Domicilio, DistanciaMetros: p.DistanciaMetros,
		EnvioCentavos: p.EnvioCentavos, SubtotalCentavos: p.SubtotalCentavos, TotalCentavos: p.TotalCentavos,
		EstadoPedido: string(p.EstadoPedido), EstadoPago: string(p.EstadoPago), MedioPago: string(p.MedioPago),
		VenceEn: p.VenceEn, CreadoEn: p.CreadoEn, Detalles: detalles,
	}
}
