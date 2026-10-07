package schema

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func MigracionesIdentidad() []*gormigrate.Migration {
	return []*gormigrate.Migration{{ID: "20261004_01_identidad", Migrate: migrarIdentidad, Rollback: revertirIdentidad}}
}

func migrarIdentidad(tx *gorm.DB) error {
	for _, sql := range []string{
		`CREATE TABLE usuario (id uuid PRIMARY KEY, nombre varchar(255) NOT NULL, apellido varchar(255) NOT NULL, correo varchar(255) NOT NULL, credencial_hash varchar(255) NOT NULL, telefono varchar(50), estado varchar(20) NOT NULL CHECK (estado IN ('ACTIVO','INACTIVO')), origen varchar(20) NOT NULL CHECK (origen IN ('PUBLICO','ADMINISTRATIVO')), es_dueno_inicial boolean NOT NULL DEFAULT false, creado_en timestamptz NOT NULL DEFAULT now(), actualizado_en timestamptz NOT NULL DEFAULT now())`,
		`CREATE UNIQUE INDEX idx_usuario_correo ON usuario (lower(correo))`,
		`CREATE UNIQUE INDEX idx_usuario_dueno_inicial ON usuario (es_dueno_inicial) WHERE es_dueno_inicial = true`,
		`CREATE TABLE rol (id uuid PRIMARY KEY, nombre varchar(100) NOT NULL UNIQUE, creado_en timestamptz NOT NULL DEFAULT now())`,
		`CREATE TABLE permiso (id uuid PRIMARY KEY, accion varchar(100) NOT NULL UNIQUE, descripcion text NOT NULL)`,
		// CASCADE: una relación de unión no tiene sentido sin alguno de sus extremos.
		`CREATE TABLE usuario_rol (usuario_id uuid NOT NULL REFERENCES usuario(id) ON DELETE CASCADE, rol_id uuid NOT NULL REFERENCES rol(id) ON DELETE CASCADE, PRIMARY KEY (usuario_id, rol_id))`,
		`CREATE INDEX idx_usuario_rol_usuario_id ON usuario_rol (usuario_id)`,
		`CREATE INDEX idx_usuario_rol_rol_id ON usuario_rol (rol_id)`,
		// CASCADE: una asignación de permiso no sobrevive al rol o permiso eliminado.
		`CREATE TABLE rol_permiso (rol_id uuid NOT NULL REFERENCES rol(id) ON DELETE CASCADE, permiso_id uuid NOT NULL REFERENCES permiso(id) ON DELETE CASCADE, PRIMARY KEY (rol_id, permiso_id))`,
		`CREATE INDEX idx_rol_permiso_rol_id ON rol_permiso (rol_id)`,
		`INSERT INTO rol (id,nombre) VALUES ('72c4806c-1bc6-482c-8ddd-4284d5275040','dueno'), ('754656a8-e56f-4d91-8c81-3b8cd58040e3','cliente')`,
		`INSERT INTO permiso (id,accion,descripcion) VALUES ('05bbba80-5427-45b9-985a-aeb59178541c','usuarios.gestionar','Gestionar usuarios y roles'), ('baf42621-42a9-4d04-8961-e7bf2b6dcc5d','catalogo.escribir','Modificar catálogo'), ('ae8f3d3d-8f58-4889-a8f3-feb24fdd11f5','stock.gestionar','Gestionar stock'), ('7c66b4f4-3ca3-4e3f-8b3c-bc430c974576','pedidos.gestionar','Gestionar pedidos')`,
		`INSERT INTO rol_permiso (rol_id,permiso_id) SELECT '72c4806c-1bc6-482c-8ddd-4284d5275040', id FROM permiso`,
	} {
		if err := tx.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func revertirIdentidad(tx *gorm.DB) error {
	for _, tabla := range []string{"rol_permiso", "usuario_rol", "permiso", "rol", "usuario"} {
		if err := tx.Exec("DROP TABLE IF EXISTS " + tabla).Error; err != nil {
			return err
		}
	}
	return nil
}
