// Package database contains persistence composition helpers.
package database

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// NewMigrator builds the versioned migrator used by backend commands. Domain
// migrations are supplied explicitly so schema changes stay reviewable and
// are never silently inferred from GORM models at server startup.
func NewMigrator(db *gorm.DB, migrations []*gormigrate.Migration) *gormigrate.Gormigrate {
	return gormigrate.New(db, gormigrate.DefaultOptions, migrations)
}
