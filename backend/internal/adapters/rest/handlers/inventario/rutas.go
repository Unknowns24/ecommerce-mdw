// Package inventario mounts the stock administration endpoints: providers,
// lot intake, availability, movement history and adjustments. Todos bajo
// mw.RequierePermiso("stock.gestionar").
package inventario

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
)

// Montar attaches this module's routes to the shared router. main.go calls
// it exactly once.
func Montar(r chi.Router, db *gorm.DB, cfg config.Config, mw *middleware.Middleware) {
	h := nuevoHandler(db)

	protegido := r.With(mw.RequiereSesion, mw.RequierePermiso("stock.gestionar"))
	protegido.Post("/api/admin/proveedores", h.altaProveedor)
	protegido.Get("/api/admin/proveedores", h.listarProveedores)
	protegido.Post("/api/admin/lotes", h.ingresoLote)
	protegido.Get("/api/admin/variantes/{id}/stock", h.stockVariante)
	protegido.Get("/api/admin/variantes/{id}/movimientos", h.movimientosVariante)
	protegido.Post("/api/admin/variantes/{id}/ajustes", h.ajusteStock)
}
