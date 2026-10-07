// Package inventario holds the request/response DTOs of the stock
// administration endpoints. The body never carries roles, permissions, costs
// a comprador shouldn't see, or anything the server itself must decide.
package inventario

import "time"

type AltaProveedorDTO struct {
	Nombre   string  `json:"nombre" validate:"required,min=1,max=255"`
	Correo   *string `json:"correo" validate:"omitempty,email"`
	Telefono *string `json:"telefono" validate:"omitempty,max=50"`
	Web      *string `json:"web" validate:"omitempty,max=255"`
}

type IngresoLoteDTO struct {
	VarianteID            string `json:"varianteId" validate:"required,uuid"`
	ProveedorID           string `json:"proveedorId" validate:"required,uuid"`
	Unidades              int    `json:"unidades" validate:"required,gt=0"`
	CostoUnitarioCentavos int64  `json:"costoUnitarioCentavos" validate:"gte=0"`
	Fecha                 string `json:"fecha" validate:"required"`
}

type AjusteStockDTO struct {
	LoteID   string `json:"loteId" validate:"required,uuid"`
	Unidades int    `json:"unidades" validate:"required"`
	Motivo   string `json:"motivo" validate:"required,min=1,max=255"`
}

type ProveedorRespuesta struct {
	ID       string    `json:"id"`
	Nombre   string    `json:"nombre"`
	Correo   *string   `json:"correo,omitempty"`
	Telefono *string   `json:"telefono,omitempty"`
	Web      *string   `json:"web,omitempty"`
	CreadoEn time.Time `json:"creadoEn"`
}

type LoteRespuesta struct {
	ID                    string    `json:"id"`
	VarianteID            string    `json:"varianteId"`
	ProveedorID           string    `json:"proveedorId"`
	UnidadesIngresadas    int       `json:"unidadesIngresadas"`
	CostoUnitarioCentavos int64     `json:"costoUnitarioCentavos"`
	FechaIngreso          time.Time `json:"fechaIngreso"`
}

type StockVarianteRespuesta struct {
	VarianteID string          `json:"varianteId"`
	Disponible int             `json:"disponible"`
	Lotes      []LoteRespuesta `json:"lotes"`
}

type MovimientoRespuesta struct {
	ID       string    `json:"id"`
	LoteID   string    `json:"loteId"`
	Tipo     string    `json:"tipo"`
	Unidades int       `json:"unidades"`
	Motivo   string    `json:"motivo"`
	CreadoEn time.Time `json:"creadoEn"`
}
