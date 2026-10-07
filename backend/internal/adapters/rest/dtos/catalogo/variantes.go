package catalogo

import "github.com/google/uuid"

type CrearVariante struct {
	ProductoID              uuid.UUID `json:"productoId"`
	Codigo                  string    `json:"codigo"`
	Nombre                  string    `json:"nombre"`
	Descripcion             string    `json:"descripcion"`
	PrecioMinoristaCentavos int64     `json:"precioMinoristaCentavos"`
	PrecioMayoristaCentavos int64     `json:"precioMayoristaCentavos"`
}

type EditarVariante struct {
	Codigo                  *string `json:"codigo"`
	Nombre                  *string `json:"nombre"`
	Descripcion             *string `json:"descripcion"`
	PrecioMinoristaCentavos *int64  `json:"precioMinoristaCentavos"`
	PrecioMayoristaCentavos *int64  `json:"precioMayoristaCentavos"`
}
