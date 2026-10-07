package database

import (
	"github.com/go-gormigrate/gormigrate/v2"

	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/migrations/schema"
)

// Todas concatenates every module's migrations in the fixed order that
// respects the foreign keys between them: identidad → catalogo → inventario
// → pedidos. Each module owns exactly one file under schema/ and nobody
// edits another module's file.
func Todas() []*gormigrate.Migration {
	var todas []*gormigrate.Migration
	todas = append(todas, schema.MigracionesIdentidad()...)  // Yasmín  20261004_01_*
	todas = append(todas, schema.MigracionesCatalogo()...)   // Genaro  20261004_02_*
	todas = append(todas, schema.MigracionesInventario()...) // vos     20261004_03_*
	todas = append(todas, schema.MigracionesPedidos()...)    // Nicolás 20261004_04_*
	return todas
}
