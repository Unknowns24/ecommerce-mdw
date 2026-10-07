// Package middleware protects HTTP endpoints with session and permission
// checks. This file is a stub: it only exists so the other three modules can
// compile and mount their routes from minute zero.
//
// TODO(Yasmín): reemplazar con el middleware real de sesión y permisos
// (JWT con HS256, verificación explícita del algoritmo, claims con roles y
// permisos). Esta firma está congelada: Agustín, Genaro y Nicolás ya
// programan contra ella. No la cambies sin avisar por el grupo.
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	jwt "github.com/Unknowns24/ecommerce-mdw/internal/adapters/jwt"
	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

// Usuario is the identity carried in the request context once a session is
// verified. It never comes from the body, a custom header, or the query.
type Usuario struct {
	ID       uuid.UUID
	Correo   string
	Roles    []string
	Permisos []string
}

// Middleware holds whatever the session verification needs (today: the app
// secret used to verify JWTs).
type Middleware struct {
	emisor *jwt.Emisor
}

// Nuevo builds the middleware with the application secret used to verify
// sessions.
func Nuevo(secreto string) *Middleware {
	return &Middleware{emisor: jwt.NuevoEmisor(secreto, 24*time.Hour)}
}

// RequiereSesion stub: siempre responde 401. TODO(Yasmín): verificar el JWT
// del header Authorization o la cookie "sesion" y meter el Usuario en el
// contexto.
func (m *Middleware) RequiereSesion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := m.usuario(r)
		if !ok {
			apierr.Responder(w, r.Method+" "+r.URL.Path, apierr.ErrNoAutenticado)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), usuarioContextKey{}, u)))
	})
}

// RequierePermiso stub: siempre responde 401 (en la versión real corre
// después de RequiereSesion y responde 403 si falta el permiso).
func (m *Middleware) RequierePermiso(permiso string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UsuarioDeContexto(r.Context())
			if !ok {
				apierr.Responder(w, r.Method+" "+r.URL.Path, apierr.ErrNoAutenticado)
				return
			}
			for _, p := range u.Permisos {
				if p == permiso {
					next.ServeHTTP(w, r)
					return
				}
			}
			apierr.Responder(w, r.Method+" "+r.URL.Path, apierr.ErrNoAutorizado)
		})
	}
}

// SesionOpcional stub: deja pasar sin usuario en el contexto. TODO(Yasmín):
// si hay un token válido, ponerlo en el contexto; si no, dejar pasar igual.
func (m *Middleware) SesionOpcional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := m.usuario(r); ok {
			r = r.WithContext(context.WithValue(r.Context(), usuarioContextKey{}, u))
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) usuario(r *http.Request) (Usuario, bool) {
	token := ""
	if auth := r.Header.Get("Authorization"); auth != "" {
		partes := strings.SplitN(auth, " ", 2)
		if len(partes) != 2 || !strings.EqualFold(partes[0], "Bearer") || partes[1] == "" {
			return Usuario{}, false
		}
		token = partes[1]
	} else if cookie, err := r.Cookie("sesion"); err == nil {
		token = cookie.Value
	}
	c, err := m.emisor.Verificar(token)
	if err != nil {
		return Usuario{}, false
	}
	return Usuario{ID: c.Sub, Correo: c.Correo, Roles: c.Roles, Permisos: c.Permisos}, true
}

type usuarioContextKey struct{}

// UsuarioDeContexto lee el Usuario que RequiereSesion/SesionOpcional dejaron
// en el contexto de la request.
func UsuarioDeContexto(ctx context.Context) (Usuario, bool) {
	u, ok := ctx.Value(usuarioContextKey{}).(Usuario)
	return u, ok
}
