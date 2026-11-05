package pedidos

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Unknowns24/ecommerce-mdw/config"
	jwt "github.com/Unknowns24/ecommerce-mdw/internal/adapters/jwt"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
)

// Prueba que Montar expone cada ruta con el control de docs/api.md. Los casos
// cortan en el middleware, antes de tocar la base: por eso alcanza con db nil.
// Una ruta que no se montó respondería 404 y la prueba lo detecta.
func TestMontarExponeLasRutasConSuControl(t *testing.T) {
	const secreto = "secreto de prueba"
	r := chi.NewRouter()
	Montar(r, nil, config.Config{PaymentReservationTTL: 15 * time.Minute}, middleware.Nuevo(secreto))

	cliente, err := jwt.NuevoEmisor(secreto, time.Hour).Emitir(jwt.Claims{Sub: uuid.New(), Roles: []string{"cliente"}})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.NewString()

	casos := []struct {
		metodo, ruta, token string
		esperado            int
	}{
		{"GET", "/api/carrito", "", 401},
		{"DELETE", "/api/carrito", "", 401},
		{"POST", "/api/carrito/items", "", 401},
		{"PATCH", "/api/carrito/items/" + id, "", 401},
		{"DELETE", "/api/carrito/items/" + id, "", 401},
		{"GET", "/api/pedidos", "", 401},
		{"GET", "/api/pedidos/" + id, "", 401},
		{"GET", "/api/admin/pedidos", "", 401},
		{"GET", "/api/admin/pedidos", cliente, 403},
		{"GET", "/api/admin/pedidos/" + id, cliente, 403},
		{"POST", "/api/admin/pedidos/" + id + "/despacho", cliente, 403},
		{"POST", "/api/admin/pedidos/" + id + "/entrega", cliente, 403},
		{"POST", "/api/admin/pedidos/" + id + "/cobro-efectivo", cliente, 403},
		{"POST", "/api/admin/pedidos/" + id + "/cancelacion", cliente, 403},
		// Sesión opcional: sin sesión llega al handler, que rechaza el cuerpo vacío.
		{"POST", "/api/pedidos", "", 400},
	}
	for _, c := range casos {
		t.Run(c.metodo+" "+c.ruta, func(t *testing.T) {
			req := httptest.NewRequest(c.metodo, c.ruta, nil)
			if c.token != "" {
				req.Header.Set("Authorization", "Bearer "+c.token)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != c.esperado {
				t.Fatalf("status %d, esperado %d", w.Code, c.esperado)
			}
		})
	}
}
