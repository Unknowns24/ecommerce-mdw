package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/Unknowns24/ecommerce-mdw/internal/adapters/jwt"
	"github.com/google/uuid"
)

func TestSesionYPermisos(t *testing.T) {
	const secreto = "secreto de prueba"
	m := Nuevo(secreto)
	emisor := jwt.NuevoEmisor(secreto, time.Hour)
	id := uuid.New()
	token, err := emisor.Emitir(jwt.Claims{Sub: id, Correo: "a@b.com", Permisos: []string{"usuarios.gestionar"}})
	if err != nil {
		t.Fatal(err)
	}
	casos := []struct {
		nombre, token string
		permiso       string
		esperado      int
	}{
		{"sin token", "", "", 401},
		{"firma inválida", token + "x", "", 401},
		{"sin permiso", token, "stock.gestionar", 403},
		{"con permiso", token, "usuarios.gestionar", 200},
		{"alg none", "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhZGQ5MzY5Zi00M2Q0LTRlYzQtOWFjYy1mZDkwZWMxM2ZjMGQiLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MX0.", "", 401},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				u, ok := UsuarioDeContexto(r.Context())
				if !ok || u.ID != id {
					t.Error("usuario ausente")
					return
				}
				w.WriteHeader(200)
			})
			var h http.Handler = base
			if c.permiso != "" {
				h = m.RequierePermiso(c.permiso)(h)
			}
			h = m.RequiereSesion(h)
			r := httptest.NewRequest("GET", "/", nil)
			if c.token != "" {
				r.Header.Set("Authorization", "Bearer "+c.token)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != c.esperado {
				t.Fatalf("status %d, esperado %d", w.Code, c.esperado)
			}
		})
	}
	vencido, _ := jwt.NuevoEmisor(secreto, time.Nanosecond).Emitir(jwt.Claims{Sub: id})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+vencido)
	w := httptest.NewRecorder()
	m.RequiereSesion(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("pasó token vencido") })).ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("vencido: %d", w.Code)
	}
}
