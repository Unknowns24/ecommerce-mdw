package catalogo

import "github.com/google/uuid"

type CrearProducto struct {
	MarcaID             uuid.UUID   `json:"marcaId"`
	Nombre              string      `json:"nombre"`
	Descripcion         string      `json:"descripcion"`
	MetodologiaRotacion string      `json:"metodologiaRotacion"`
	Categorias          []uuid.UUID `json:"categorias"`
}

type EditarProducto struct {
	MarcaID             *uuid.UUID   `json:"marcaId"`
	Nombre              *string      `json:"nombre"`
	Descripcion         *string      `json:"descripcion"`
	MetodologiaRotacion *string      `json:"metodologiaRotacion"`
	Categorias          *[]uuid.UUID `json:"categorias"`
}
