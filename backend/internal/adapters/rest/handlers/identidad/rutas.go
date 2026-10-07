package identidad

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/identidad"
	jwt "github.com/Unknowns24/ecommerce-mdw/internal/adapters/jwt"
	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/identidad"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/identidad"
)

type handler struct {
	servicio *uc.Servicio
	repo     *repo.Repositorio
	cfg      config.Config
	validar  *validator.Validate
}

func Montar(r chi.Router, db *gorm.DB, cfg config.Config, mw *middleware.Middleware) {
	lo := repo.Nuevo(db)
	h := &handler{uc.Nuevo(lo, jwt.NuevoEmisor(cfg.AppSecretKey, 24*time.Hour), cfg.BootstrapMasterKey), lo, cfg, validator.New()}
	r.Post("/api/auth/bootstrap", h.bootstrap)
	r.Post("/api/auth/registro", h.registro)
	r.Post("/api/auth/login", h.login)
	r.With(mw.RequiereSesion).Post("/api/auth/logout", h.logout)
	r.With(mw.RequiereSesion).Get("/api/cuenta", h.cuenta)
	r.With(mw.RequiereSesion).Patch("/api/cuenta", h.actualizarCuenta)
	admin := r.With(mw.RequiereSesion, mw.RequierePermiso("usuarios.gestionar"), requiereDueno)
	admin.Get("/api/admin/usuarios", h.listarUsuarios)
	admin.Post("/api/admin/usuarios", h.crearUsuario)
	admin.Patch("/api/admin/usuarios/{id}/estado", h.cambiarEstado)
	admin.Get("/api/admin/roles", h.listarRoles)
	admin.Post("/api/admin/roles", h.crearRol)
	admin.Put("/api/admin/usuarios/{id}/roles", h.asignarRoles)
}
func (h *handler) leer(r *http.Request, dest any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := dec.Decode(dest); err != nil {
		return apierr.ErrValidacion{Campos: map[string]string{"body": "Enviá un JSON válido"}}
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return apierr.ErrValidacion{Campos: map[string]string{"body": "Enviá un solo objeto JSON"}}
	}
	if err := h.validar.Struct(dest); err != nil {
		campos := map[string]string{}
		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			for _, e := range verr {
				campo := strings.ToLower(e.Field())
				mensaje := "Completá este campo"
				if e.Tag() == "email" {
					mensaje = "Ingresá un correo válido"
				}
				if e.Tag() == "min" {
					mensaje = "La contraseña tiene que tener al menos 8 caracteres"
				}
				if e.Tag() == "oneof" {
					mensaje = "Elegí ACTIVO o INACTIVO"
				}
				campos[campo] = mensaje
			}
		}
		return apierr.ErrValidacion{Campos: campos}
	}
	return nil
}
func requiereDueno(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.UsuarioDeContexto(r.Context())
		if !ok {
			apierr.Responder(w, r.Method+" "+r.URL.Path, apierr.ErrNoAutenticado)
			return
		}
		for _, rol := range u.Roles {
			if rol == "dueno" {
				next.ServeHTTP(w, r)
				return
			}
		}
		apierr.Responder(w, r.Method+" "+r.URL.Path, apierr.ErrNoAutorizado)
	})
}
func responder(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, uc.ErrBootstrapClave) {
		apierr.JSON(w, 403, apierr.Cuerpo{Error: "Clave de inicialización inválida"})
		return
	}
	if errors.Is(err, uc.CredencialesIncorrectas()) {
		apierr.JSON(w, 401, apierr.Cuerpo{Error: "Correo o contraseña incorrectos"})
		return
	}
	if repo.EsNoEncontrado(err) {
		apierr.Responder(w, r.Method+" "+r.URL.Path, apierr.ErrNoEncontrado)
		return
	}
	apierr.Responder(w, r.Method+" "+r.URL.Path, err)
}
func (h *handler) bootstrap(w http.ResponseWriter, r *http.Request) {
	var b dto.Bootstrap
	if err := h.leer(r, &b); err != nil {
		responder(w, r, err)
		return
	}
	u, err := h.servicio.Bootstrap(r.Context(), b.Nombre, b.Apellido, b.Correo, b.Contrasena, b.ClaveBootstrap)
	if err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 201, u)
}
func (h *handler) registro(w http.ResponseWriter, r *http.Request) {
	var b dto.Registro
	if err := h.leer(r, &b); err != nil {
		responder(w, r, err)
		return
	}
	u, err := h.servicio.Registrar(r.Context(), b.Nombre, b.Apellido, b.Correo, b.Contrasena, false)
	if err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 201, u)
}
func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var b dto.Login
	if err := h.leer(r, &b); err != nil {
		responder(w, r, err)
		return
	}
	token, err := h.servicio.Login(r.Context(), b.Correo, b.Contrasena)
	if err != nil {
		responder(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "sesion", Value: token, Path: "/", HttpOnly: true, Secure: h.cfg.Environment == "production", SameSite: http.SameSiteLaxMode, MaxAge: 86400})
	apierr.JSON(w, 200, map[string]string{"token": token})
}
func (h *handler) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "sesion", Value: "", Path: "/", HttpOnly: true, Secure: h.cfg.Environment == "production", SameSite: http.SameSiteLaxMode, MaxAge: -1})
	apierr.JSON(w, 200, map[string]string{"mensaje": "Sesión cerrada"})
}
func usuarioID(r *http.Request) uuid.UUID {
	u, _ := middleware.UsuarioDeContexto(r.Context())
	return u.ID
}
func (h *handler) cuenta(w http.ResponseWriter, r *http.Request) {
	u, err := h.servicio.Cuenta(r.Context(), usuarioID(r))
	if err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 200, u)
}
func (h *handler) actualizarCuenta(w http.ResponseWriter, r *http.Request) {
	var b dto.Cuenta
	if err := h.leer(r, &b); err != nil {
		responder(w, r, err)
		return
	}
	u, err := h.servicio.ActualizarCuenta(r.Context(), usuarioID(r), b.Nombre, b.Apellido, b.Telefono)
	if err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 200, u)
}
func paginacion(r *http.Request) (int, int) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
func (h *handler) listarUsuarios(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginacion(r)
	estado := r.URL.Query().Get("estado")
	if estado != "" && estado != "ACTIVO" && estado != "INACTIVO" {
		responder(w, r, apierr.ErrValidacion{Campos: map[string]string{"estado": "Elegí ACTIVO o INACTIVO"}})
		return
	}
	x, err := h.servicio.ListarUsuarios(r.Context(), limit, offset, estado)
	if err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 200, map[string]any{"items": x})
}
func (h *handler) crearUsuario(w http.ResponseWriter, r *http.Request) {
	var b dto.Registro
	if err := h.leer(r, &b); err != nil {
		responder(w, r, err)
		return
	}
	u, err := h.servicio.Registrar(r.Context(), b.Nombre, b.Apellido, b.Correo, b.Contrasena, true)
	if err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 201, u)
}
func idRuta(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apierr.ErrValidacion{Campos: map[string]string{"id": "Ingresá un identificador válido"}}
	}
	return id, nil
}
func (h *handler) cambiarEstado(w http.ResponseWriter, r *http.Request) {
	id, err := idRuta(r)
	if err != nil {
		responder(w, r, err)
		return
	}
	var b dto.Estado
	if err := h.leer(r, &b); err != nil {
		responder(w, r, err)
		return
	}
	if err := h.servicio.CambiarEstado(r.Context(), usuarioID(r), id, b.Estado); err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 200, map[string]string{"estado": b.Estado})
}
func (h *handler) listarRoles(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginacion(r)
	x, err := h.servicio.ListarRoles(r.Context(), limit, offset)
	if err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 200, map[string]any{"items": x})
}
func (h *handler) crearRol(w http.ResponseWriter, r *http.Request) {
	var b dto.Rol
	if err := h.leer(r, &b); err != nil {
		responder(w, r, err)
		return
	}
	x, err := h.servicio.CrearRol(r.Context(), b.Nombre, b.Permisos)
	if err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 201, x)
}
func (h *handler) asignarRoles(w http.ResponseWriter, r *http.Request) {
	id, err := idRuta(r)
	if err != nil {
		responder(w, r, err)
		return
	}
	var b dto.Roles
	if err := h.leer(r, &b); err != nil {
		responder(w, r, err)
		return
	}
	if err := h.servicio.ReemplazarRoles(r.Context(), usuarioID(r), id, b.Roles); err != nil {
		responder(w, r, err)
		return
	}
	apierr.JSON(w, 200, map[string]any{"roles": b.Roles})
}
