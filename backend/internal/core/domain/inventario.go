// Package domain contains the core entities of the kernel, inventory and
// payments modules. Movements and reservations are append-only: a mistake is
// corrected with a compensating movement, never by editing or deleting a row.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// TipoMovimiento is the closed list of stock movement kinds.
type TipoMovimiento string

const (
	MovimientoIngreso TipoMovimiento = "INGRESO"
	MovimientoAjuste  TipoMovimiento = "AJUSTE"
	MovimientoSalida  TipoMovimiento = "SALIDA"
)

// EstadoReserva is the closed list of stock reservation states.
type EstadoReserva string

const (
	ReservaActiva    EstadoReserva = "ACTIVA"
	ReservaConsumida EstadoReserva = "CONSUMIDA"
	ReservaLiberada  EstadoReserva = "LIBERADA"
)

// Proveedor is a supplier of merchandise lots. Only the name is required.
type Proveedor struct {
	ID       uuid.UUID `gorm:"column:id;primaryKey"`
	Nombre   string    `gorm:"column:nombre;not null"`
	Correo   *string   `gorm:"column:correo"`
	Telefono *string   `gorm:"column:telefono"`
	Web      *string   `gorm:"column:web"`
	CreadoEn time.Time `gorm:"column:creado_en;not null;autoCreateTime"`
}

func (Proveedor) TableName() string { return "proveedor" }

// Lote is one intake of units of a variant from a provider, at a given cost.
// Stock availability is derived from lots minus active reservations — see
// usecases/inventario/disponibilidad.go.
type Lote struct {
	ID                    uuid.UUID `gorm:"column:id;primaryKey"`
	VarianteID            uuid.UUID `gorm:"column:variante_id;not null"`
	ProveedorID           uuid.UUID `gorm:"column:proveedor_id;not null"`
	UnidadesIngresadas    int       `gorm:"column:unidades_ingresadas;not null"`
	CostoUnitarioCentavos int64     `gorm:"column:costo_unitario_centavos;not null"`
	FechaIngreso          time.Time `gorm:"column:fecha_ingreso;not null"`
	CreadoEn              time.Time `gorm:"column:creado_en;not null;autoCreateTime"`
}

func (Lote) TableName() string { return "lote" }

// MovimientoStock is an immutable record of a change in a lot's units:
// intake, manual adjustment, or outbound sale. Movements are history: they
// are never edited or deleted.
type MovimientoStock struct {
	ID            uuid.UUID      `gorm:"column:id;primaryKey"`
	LoteID        uuid.UUID      `gorm:"column:lote_id;not null"`
	Tipo          TipoMovimiento `gorm:"column:tipo;not null"`
	Unidades      int            `gorm:"column:unidades;not null"`
	Motivo        string         `gorm:"column:motivo;not null"`
	ResponsableID *uuid.UUID     `gorm:"column:responsable_id"`
	PedidoID      *uuid.UUID     `gorm:"column:pedido_id"`
	CreadoEn      time.Time      `gorm:"column:creado_en;not null;autoCreateTime"`
}

func (MovimientoStock) TableName() string { return "movimiento_stock" }

// ReservaStock is a commitment of units from a specific lot toward a pedido.
// Reservations are history: they are never edited, only transitioned between
// ACTIVA, CONSUMIDA and LIBERADA by the inventory service.
type ReservaStock struct {
	ID       uuid.UUID     `gorm:"column:id;primaryKey"`
	PedidoID uuid.UUID     `gorm:"column:pedido_id;not null"`
	LoteID   uuid.UUID     `gorm:"column:lote_id;not null"`
	Unidades int           `gorm:"column:unidades;not null"`
	Estado   EstadoReserva `gorm:"column:estado;not null"`
	VenceEn  *time.Time    `gorm:"column:vence_en"`
	CreadoEn time.Time     `gorm:"column:creado_en;not null;autoCreateTime"`
}

func (ReservaStock) TableName() string { return "reserva_stock" }
