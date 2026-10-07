package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProveedorPago is the closed list of payment providers.
type ProveedorPago string

const (
	ProveedorMercadoPago ProveedorPago = "MERCADO_PAGO"
	ProveedorEfectivo    ProveedorPago = "EFECTIVO"
)

// EstadoIntentoPago is the closed list of payment attempt states.
type EstadoIntentoPago string

const (
	IntentoPendiente EstadoIntentoPago = "PENDIENTE"
	IntentoAprobado  EstadoIntentoPago = "APROBADO"
	IntentoRechazado EstadoIntentoPago = "RECHAZADO"
)

// IntentoPago is one payment attempt against a pedido. The unique index on
// (proveedor, referencia_externa) is what makes it impossible to process the
// same Mercado Pago notification twice.
type IntentoPago struct {
	ID                uuid.UUID         `gorm:"column:id;primaryKey"`
	PedidoID          uuid.UUID         `gorm:"column:pedido_id;not null"`
	Proveedor         ProveedorPago     `gorm:"column:proveedor;not null"`
	ReferenciaExterna string            `gorm:"column:referencia_externa;not null"`
	Estado            EstadoIntentoPago `gorm:"column:estado;not null"`
	MontoCentavos     int64             `gorm:"column:monto_centavos;not null"`
	CreadoEn          time.Time         `gorm:"column:creado_en;not null;autoCreateTime"`
	ActualizadoEn     time.Time         `gorm:"column:actualizado_en;not null;autoUpdateTime"`
}

func (IntentoPago) TableName() string { return "intento_pago" }
