// Package pedidos mounts the cart, checkout and order endpoints.
// TODO(Nicolás): reemplazar con las rutas reales (carrito, checkout,
// consulta de pedidos, transiciones de despacho/entrega/cancelación).
package pedidos

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
)

// Montar attaches this module's routes to the shared router. main.go calls
// it exactly once.
func Montar(r chi.Router, db *gorm.DB, cfg config.Config, mw *middleware.Middleware) {}
