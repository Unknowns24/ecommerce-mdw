package domain

import (
	"time"

	"github.com/google/uuid"
)

type Marca struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Nombre   string    `gorm:"size:150;not null;uniqueIndex" json:"nombre"`
	Estado   string    `gorm:"size:10;not null" json:"estado"`
	CreadoEn time.Time `gorm:"not null" json:"creadoEn"`
}

func (Marca) TableName() string { return "marca" }

type Categoria struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Nombre   string     `gorm:"size:150;not null" json:"nombre"`
	PadreID  *uuid.UUID `gorm:"type:uuid;index" json:"padreId,omitempty"`
	Estado   string     `gorm:"size:10;not null" json:"estado"`
	CreadoEn time.Time  `gorm:"not null" json:"creadoEn"`
}

func (Categoria) TableName() string { return "categoria" }

type Producto struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	MarcaID             uuid.UUID `gorm:"type:uuid;not null;index" json:"marcaId"`
	Nombre              string    `gorm:"size:255;not null" json:"nombre"`
	Descripcion         string    `gorm:"type:text" json:"descripcion"`
	Estado              string    `gorm:"size:10;not null" json:"estado"`
	MetodologiaRotacion string    `gorm:"size:4;not null" json:"metodologiaRotacion"`
	CreadoEn            time.Time `gorm:"not null" json:"creadoEn"`
	ActualizadoEn       time.Time `gorm:"not null" json:"actualizadoEn"`
}

func (Producto) TableName() string { return "producto" }

type ProductoCategoria struct {
	ProductoID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	CategoriaID uuid.UUID `gorm:"type:uuid;primaryKey;index"`
}

func (ProductoCategoria) TableName() string { return "producto_categoria" }

type Variante struct {
	ID                      uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductoID              uuid.UUID `gorm:"type:uuid;not null;index" json:"productoId"`
	Codigo                  string    `gorm:"size:100;not null;uniqueIndex" json:"codigo"`
	Nombre                  string    `gorm:"size:255;not null" json:"nombre"`
	Descripcion             string    `gorm:"type:text" json:"descripcion"`
	PrecioMinoristaCentavos int64     `gorm:"type:bigint;not null" json:"precioMinoristaCentavos"`
	PrecioMayoristaCentavos int64     `gorm:"type:bigint;not null" json:"precioMayoristaCentavos"`
	Estado                  string    `gorm:"size:10;not null" json:"estado"`
	CreadoEn                time.Time `gorm:"not null" json:"creadoEn"`
	ActualizadoEn           time.Time `gorm:"not null" json:"actualizadoEn"`
}

func (Variante) TableName() string { return "variante" }

type ImagenVariante struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	VarianteID       uuid.UUID `gorm:"type:uuid;not null;index" json:"varianteId"`
	ReferenciaImagen string    `gorm:"size:2048;not null" json:"referenciaImagen"`
	EsPrincipal      bool      `gorm:"not null" json:"esPrincipal"`
	Orden            int       `gorm:"not null" json:"orden"`
}

func (ImagenVariante) TableName() string { return "imagen_variante" }
