package domain

import (
	"time"

	"github.com/google/uuid"
)

// Listas cerradas del pedido. Son strings con tipo propio para que el
// compilador no deje mezclar un modo de entrega con un medio de pago; en la
// base van con CHECK (ver la migración 20261004_04_pedidos).
type (
	ModoEntrega  string
	EstadoPedido string
	EstadoPago   string
	MedioPago    string
)

const (
	EntregaRetiro ModoEntrega = "RETIRO"
	EntregaEnvio  ModoEntrega = "ENVIO"

	PedidoPendienteDePago EstadoPedido = "PENDIENTE_DE_PAGO"
	PedidoConfirmado      EstadoPedido = "CONFIRMADO"
	PedidoDespachado      EstadoPedido = "DESPACHADO"
	PedidoCompletado      EstadoPedido = "COMPLETADO"
	PedidoCancelado       EstadoPedido = "CANCELADO"

	PagoPendiente EstadoPago = "PENDIENTE"
	PagoAprobado  EstadoPago = "APROBADO"
	PagoRechazado EstadoPago = "RECHAZADO"

	PagoMercadoPago MedioPago = "MERCADO_PAGO"
	PagoEfectivo    MedioPago = "EFECTIVO"
)

// Carrito es la selección editable de un cliente registrado. El invitado
// guarda el suyo en el navegador, por eso UsuarioID puede ser nulo en el
// modelo aunque en la práctica el backend solo crea carritos con usuario.
type Carrito struct {
	ID            uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UsuarioID     *uuid.UUID    `gorm:"type:uuid;uniqueIndex"`
	Items         []ItemCarrito `gorm:"foreignKey:CarritoID"`
	CreadoEn      time.Time     `gorm:"autoCreateTime"`
	ActualizadoEn time.Time     `gorm:"autoUpdateTime"`
}

func (Carrito) TableName() string { return "carrito" }

// ItemCarrito es una línea del carrito: una variante y cuántas unidades.
// No guarda precio: el carrito nunca congela precios.
type ItemCarrito struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CarritoID  uuid.UUID `gorm:"type:uuid;not null;index"`
	VarianteID uuid.UUID `gorm:"type:uuid;not null"`
	Unidades   int       `gorm:"not null"`
	CreadoEn   time.Time `gorm:"autoCreateTime"`
}

func (ItemCarrito) TableName() string { return "item_carrito" }

// Pedido es la compra ya creada. Todo lo que dice (comprador, entrega,
// importes) es una copia de lo que pasó ese día y no se edita después.
type Pedido struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Numero int64     `gorm:"default:nextval('pedido_numero_seq');uniqueIndex"`

	// UsuarioID es nulo cuando compró un invitado.
	UsuarioID *uuid.UUID `gorm:"type:uuid;index"`

	CompradorNombre   string `gorm:"not null"`
	CompradorApellido string `gorm:"not null"`
	CompradorCorreo   string `gorm:"not null"`
	CompradorTelefono string `gorm:"not null"`
	CompradorDNI      string `gorm:"not null"`

	ModoEntrega     ModoEntrega `gorm:"type:varchar(10);not null"`
	Domicilio       *string
	DistanciaMetros *int

	EnvioCentavos    int64 `gorm:"not null"`
	SubtotalCentavos int64 `gorm:"not null"`
	TotalCentavos    int64 `gorm:"not null"`

	EstadoPedido EstadoPedido `gorm:"type:varchar(20);not null;index"`
	EstadoPago   EstadoPago   `gorm:"type:varchar(10);not null"`
	MedioPago    MedioPago    `gorm:"type:varchar(15);not null"`

	// TokenAcceso es el secreto del enlace privado del invitado: 32 bytes
	// aleatorios en base64 URL. Nunca se escribe en un log.
	TokenAcceso string `gorm:"not null;uniqueIndex"`

	// VenceEn solo existe para Mercado Pago (reserva de 15 minutos).
	VenceEn *time.Time

	// Quién canceló, cuándo y por qué. Son nulos mientras el pedido no se
	// cancela; el responsable sale de la sesión, nunca del body.
	CanceladoPor      *uuid.UUID `gorm:"type:uuid"`
	CanceladoEn       *time.Time
	MotivoCancelacion *string

	Detalles      []DetallePedido `gorm:"foreignKey:PedidoID"`
	CreadoEn      time.Time       `gorm:"autoCreateTime;index"`
	ActualizadoEn time.Time       `gorm:"autoUpdateTime"`
}

func (Pedido) TableName() string { return "pedido" }

// DetallePedido es la relación N-N con datos propios entre Pedido y
// Variante. Código, nombre y precio unitario son una copia inmutable de lo
// que el comprador compró ese día, no una desnormalización del catálogo: si
// mañana suben el precio, el pedido viejo tiene que seguir diciendo lo que
// decía.
type DetallePedido struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PedidoID               uuid.UUID `gorm:"type:uuid;not null;index"`
	VarianteID             uuid.UUID `gorm:"type:uuid;not null"`
	Codigo                 string    `gorm:"not null"`
	Nombre                 string    `gorm:"not null"`
	Unidades               int       `gorm:"not null"`
	PrecioUnitarioCentavos int64     `gorm:"not null"`
	SubtotalCentavos       int64     `gorm:"not null"`
}

func (DetallePedido) TableName() string { return "detalle_pedido" }

// ConfiguracionTienda guarda las decisiones comerciales que el pedido
// consulta al cotizar. Cambiarla no altera pedidos ya creados.
type ConfiguracionTienda struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EfectivoHabilitado    bool      `gorm:"not null"`
	DistanciaMaximaMetros int       `gorm:"not null"`
	CreadoEn              time.Time `gorm:"autoCreateTime"`
}

func (ConfiguracionTienda) TableName() string { return "configuracion_tienda" }

// TarifaDistancia es un rango de distancia con su costo de envío. Los rangos
// se evalúan en metros y no se solapan (ver CotizarEnvio).
type TarifaDistancia struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ConfiguracionID uuid.UUID `gorm:"type:uuid;not null"`
	DesdeMetros     int       `gorm:"not null"`
	HastaMetros     int       `gorm:"not null"`
	CostoCentavos   int64     `gorm:"not null"`
}

func (TarifaDistancia) TableName() string { return "tarifa_distancia" }
