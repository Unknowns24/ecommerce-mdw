package pagos

import (
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
)

// Montar attaches this module's routes to the shared router. main.go calls
// it exactly once.
func Montar(r chi.Router, db *gorm.DB, cfg config.Config, mw *middleware.Middleware) {
	h := nuevoHandler(db, cfg)

	r.With(mw.SesionOpcional).Post("/api/pedidos/{id}/pago", h.iniciarPago)
	r.Post("/api/pagos/webhook", h.webhook) // público a propósito: lo llama Mercado Pago, no un usuario
	r.With(mw.RequiereSesion).Get("/api/pedidos/{id}/pago", h.estadoPago)
}
