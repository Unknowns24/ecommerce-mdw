package domain

import "time"

// PedidoPublico es lo único que se muestra por el enlace privado del
// invitado (GET /api/pedidos/publico/{token}). Es un tipo aparte, con menos
// campos que Pedido, a propósito: sin usuario_id, sin correo ni DNI del
// comprador y sin ids internos. Si el enlace reusara el modelo completo, el
// día que alguien le agregue un campo a Pedido, ese campo se publicaría en
// internet sin que nadie lo decidiera.
type PedidoPublico struct {
	Numero           int64
	ModoEntrega      ModoEntrega
	Domicilio        *string
	DistanciaMetros  *int
	EnvioCentavos    int64
	SubtotalCentavos int64
	TotalCentavos    int64
	EstadoPedido     EstadoPedido
	EstadoPago       EstadoPago
	MedioPago        MedioPago
	VenceEn          *time.Time
	CreadoEn         time.Time
	Detalles         []DetallePublico
}

// DetallePublico es una línea del pedido sin ids internos.
type DetallePublico struct {
	Codigo                 string
	Nombre                 string
	Unidades               int
	PrecioUnitarioCentavos int64
	SubtotalCentavos       int64
}
