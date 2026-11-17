// Package pedidos mounts the cart, checkout and order endpoints.
package pedidos

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/pedidos"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
	catalogo "github.com/Unknowns24/ecommerce-mdw/internal/usecases/catalogo"
	inventario "github.com/Unknowns24/ecommerce-mdw/internal/usecases/inventario"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

// Montar attaches this module's routes to the shared router. main.go calls
// it exactly once. Quién puede llamar a cada ruta está en docs/api.md.
func Montar(r chi.Router, db *gorm.DB, cfg config.Config, mw *middleware.Middleware) {
	pedidos := repo.NuevoRepositorioPedidos(db)
	configuracion := repo.NuevoRepositorioConfiguracion(db)
	stock := inventario.NuevoServicioStock(db)
	lector := lectorCatalogo{catalogo.NuevoLector(db)}

	h := NuevoHandler(
		uc.NuevoCheckout(lector, reservadorStock{stock}, pedidos, configuracion, repo.NuevasTransacciones(db), cfg.PaymentReservationTTL, time.Now),
		uc.NuevaConsulta(pedidos),
		uc.NuevaGestion(pedidos, stock, repo.NuevasTransacciones(db), time.Now),
		uc.NuevaGestionCarrito(repo.NuevoRepositorioCarritos(db), lector, stock),
		uc.NuevaConfiguracion(configuracion),
	)

	conSesion := r.With(mw.RequiereSesion)
	conSesion.Get("/api/carrito", h.VerCarrito)
	conSesion.Delete("/api/carrito", h.VaciarCarrito)
	conSesion.Post("/api/carrito/items", h.AgregarAlCarrito)
	conSesion.Patch("/api/carrito/items/{id}", h.CambiarUnidadesDelCarrito)
	conSesion.Delete("/api/carrito/items/{id}", h.QuitarDelCarrito)

	r.With(mw.SesionOpcional).Post("/api/pedidos", h.CrearPedido)
	conSesion.Get("/api/pedidos", h.ListarMisPedidos)
	conSesion.Get("/api/pedidos/{id}", h.ObtenerPedido)
	r.Get("/api/pedidos/publico/{token}", h.ObtenerPedidoPublico) // público a propósito: el token es la credencial del invitado

	r.Get("/api/tienda/configuracion", h.ObtenerConfiguracionTienda) // público a propósito: son datos comerciales que el checkout necesita antes de que haya sesión

	admin := r.With(mw.RequiereSesion, mw.RequierePermiso("pedidos.gestionar"))
	admin.Get("/api/admin/pedidos", h.ListarPedidosAdmin)
	admin.Get("/api/admin/pedidos/{id}", h.ObtenerPedidoAdmin)
	admin.Post("/api/admin/pedidos/{id}/despacho", h.DespacharPedido)
	admin.Post("/api/admin/pedidos/{id}/entrega", h.EntregarPedido)
	admin.Post("/api/admin/pedidos/{id}/cobro-efectivo", h.CobrarPedidoEnEfectivo)
	admin.Post("/api/admin/pedidos/{id}/cancelacion", h.CancelarPedido)
}

// lectorCatalogo adapta el Lector del catálogo al puerto del pedido: los dos
// tipos de variante tienen los mismos campos, pero son de paquetes distintos.
type lectorCatalogo struct{ l *catalogo.Lector }

func (a lectorCatalogo) VariantePorID(ctx context.Context, id uuid.UUID) (out.VarianteVendible, error) {
	v, err := a.l.VariantePorID(ctx, id)
	if err != nil {
		return out.VarianteVendible{}, err
	}
	return out.VarianteVendible(v), nil
}

// reservadorStock adapta ServicioStock.Reservar a las líneas del puerto.
type reservadorStock struct{ s *inventario.ServicioStock }

func (a reservadorStock) Reservar(ctx context.Context, tx *gorm.DB, pedidoID uuid.UUID, items []out.ItemReservaStock) error {
	lineas := make([]inventario.ItemReserva, len(items))
	for i, it := range items {
		lineas[i] = inventario.ItemReserva{VarianteID: it.VarianteID, Unidades: it.Unidades}
	}
	return a.s.Reservar(ctx, tx, pedidoID, lineas)
}
