package catalogo

import (
	"encoding/json"

	"github.com/google/uuid"
)

type CrearMarca struct {
	Nombre string `json:"nombre"`
}
type EditarMarca struct {
	Nombre *string `json:"nombre"`
}
type CrearCategoria struct {
	Nombre  string     `json:"nombre"`
	PadreID *uuid.UUID `json:"padreId"`
}
type EditarCategoria struct {
	Nombre  *string         `json:"nombre"`
	PadreID json.RawMessage `json:"padreId"`
}
