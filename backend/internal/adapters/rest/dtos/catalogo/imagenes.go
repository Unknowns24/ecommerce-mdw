package catalogo

import "github.com/google/uuid"

type CrearImagen struct {
	ReferenciaImagen string `json:"referenciaImagen"`
	Orden            int    `json:"orden"`
}
type ImagenPrincipal struct {
	ImagenID uuid.UUID `json:"imagenId"`
}
