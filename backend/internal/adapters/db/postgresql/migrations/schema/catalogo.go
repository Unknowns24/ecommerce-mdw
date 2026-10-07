package schema

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// MigracionesCatalogo crea el esquema del catálogo después de identidad.
func MigracionesCatalogo() []*gormigrate.Migration {
	return []*gormigrate.Migration{{
		ID: "20261004_02_catalogo",
		Migrate: func(db *gorm.DB) error {
			statements := []string{
				`CREATE TABLE marca (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), nombre varchar(150) NOT NULL UNIQUE, estado varchar(10) NOT NULL CHECK (estado IN ('ACTIVA','INACTIVA')), creado_en timestamptz NOT NULL DEFAULT now())`,
				// A child category is retained if its parent is deactivated; physical deletion is restricted.
				`CREATE TABLE categoria (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), nombre varchar(150) NOT NULL, padre_id uuid REFERENCES categoria(id) ON DELETE RESTRICT, estado varchar(10) NOT NULL CHECK (estado IN ('ACTIVA','INACTIVA')), creado_en timestamptz NOT NULL DEFAULT now())`,
				`CREATE INDEX idx_categoria_padre_id ON categoria(padre_id)`,
				// A brand with products remains for historical references.
				`CREATE TABLE producto (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), marca_id uuid NOT NULL REFERENCES marca(id) ON DELETE RESTRICT, nombre varchar(255) NOT NULL, descripcion text NOT NULL DEFAULT '', estado varchar(10) NOT NULL CHECK (estado IN ('ACTIVO','INACTIVO')), metodologia_rotacion varchar(4) NOT NULL CHECK (metodologia_rotacion IN ('FIFO','LIFO')), creado_en timestamptz NOT NULL DEFAULT now(), actualizado_en timestamptz NOT NULL DEFAULT now())`,
				`CREATE INDEX idx_producto_marca_id ON producto(marca_id)`,
				// Join rows have no meaning after either side is physically removed.
				`CREATE TABLE producto_categoria (producto_id uuid NOT NULL REFERENCES producto(id) ON DELETE CASCADE, categoria_id uuid NOT NULL REFERENCES categoria(id) ON DELETE CASCADE, PRIMARY KEY (producto_id,categoria_id))`,
				`CREATE INDEX idx_producto_categoria_categoria_id ON producto_categoria(categoria_id)`,
				// Variants with sales must keep their product; catalog removal is logical.
				`CREATE TABLE variante (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), producto_id uuid NOT NULL REFERENCES producto(id) ON DELETE RESTRICT, codigo varchar(100) NOT NULL UNIQUE, nombre varchar(255) NOT NULL, descripcion text NOT NULL DEFAULT '', precio_minorista_centavos bigint NOT NULL CHECK (precio_minorista_centavos >= 0), precio_mayorista_centavos bigint NOT NULL CHECK (precio_mayorista_centavos >= 0), estado varchar(10) NOT NULL CHECK (estado IN ('ACTIVA','INACTIVA')), creado_en timestamptz NOT NULL DEFAULT now(), actualizado_en timestamptz NOT NULL DEFAULT now())`,
				`CREATE INDEX idx_variante_producto_id ON variante(producto_id)`,
				`CREATE INDEX idx_variante_lower_nombre ON variante (lower(nombre))`,
				// Images have no meaning without their variant.
				`CREATE TABLE imagen_variante (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), variante_id uuid NOT NULL REFERENCES variante(id) ON DELETE CASCADE, referencia_imagen varchar(2048) NOT NULL, es_principal boolean NOT NULL DEFAULT false, orden integer NOT NULL DEFAULT 0)`,
				`CREATE INDEX idx_imagen_variante_variante_id ON imagen_variante(variante_id)`,
				`CREATE UNIQUE INDEX idx_imagen_variante_principal ON imagen_variante(variante_id) WHERE es_principal`,
			}
			for _, sql := range statements {
				if err := db.Exec(sql).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(db *gorm.DB) error {
			for _, table := range []string{"imagen_variante", "variante", "producto_categoria", "producto", "categoria", "marca"} {
				if err := db.Exec("DROP TABLE " + table).Error; err != nil {
					return err
				}
			}
			return nil
		},
	}}
}
