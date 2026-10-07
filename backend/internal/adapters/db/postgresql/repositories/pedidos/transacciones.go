package pedidos

import (
	"context"

	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// Transacciones abre transacciones de GORM para los casos de uso.
type Transacciones struct{ db *gorm.DB }

func NuevasTransacciones(db *gorm.DB) *Transacciones { return &Transacciones{db: db} }

// EnTransaccion corre fn dentro de una transacción: si devuelve error (o
// hay un panic), GORM hace ROLLBACK de todo; si no, COMMIT.
func (t *Transacciones) EnTransaccion(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(fn)
}

// Comprobaciones en compilación: si alguna firma deja de coincidir con el
// puerto, el proyecto no compila.
var (
	_ out.Transaccionador     = (*Transacciones)(nil)
	_ out.CreadorDePedidos    = (*RepositorioPedidos)(nil)
	_ out.LectorConfiguracion = (*RepositorioConfiguracion)(nil)
	_ out.LectorPedidos       = (*RepositorioPedidos)(nil)
	_ out.GestorPedidos       = (*RepositorioPedidos)(nil)
	_ out.RepositorioCarrito  = (*RepositorioCarritos)(nil)
)
