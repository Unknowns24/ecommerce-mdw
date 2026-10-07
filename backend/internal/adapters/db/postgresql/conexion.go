// Package postgresql contains the PostgreSQL composition helpers: opening
// the connection pool and running the versioned migrations.
package postgresql

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Abrir opens the GORM connection against dsn, configures the connection
// pool, and verifies connectivity with a Ping before returning.
func Abrir(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("abrir conexión a postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("obtener *sql.DB subyacente: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("verificar conexión a postgres: %w", err)
	}

	return db, nil
}
