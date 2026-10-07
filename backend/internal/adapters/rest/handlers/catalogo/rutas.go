package catalogo

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/catalogo"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	usecase "github.com/Unknowns24/ecommerce-mdw/internal/usecases/catalogo"
)

func Montar(r chi.Router, db *gorm.DB, cfg config.Config, mw *middleware.Middleware) {
	s := usecase.NuevoPublico(db)
	// La tienda se navega sin cuenta: estas cuatro rutas son públicas a propósito.
	r.Get("/api/catalogo/variantes", listar(s))
	r.Get("/api/catalogo/variantes/{id}", detalle(s))
	r.Get("/api/catalogo/categorias", func(w http.ResponseWriter, req *http.Request) {
		rows, err := s.Categorias(req.Context())
		if err != nil {
			apierr.Responder(w, "categorias", err)
			return
		}
		apierr.JSON(w, http.StatusOK, rows)
	})
	r.Get("/api/catalogo/marcas", func(w http.ResponseWriter, req *http.Request) {
		rows, err := s.Marcas(req.Context())
		if err != nil {
			apierr.Responder(w, "marcas", err)
			return
		}
		apierr.JSON(w, http.StatusOK, rows)
	})
	r.Route("/api/admin", func(admin chi.Router) {
		admin.Use(mw.RequiereSesion, mw.RequierePermiso("catalogo.escribir"))
		montarProductos(admin, usecase.NuevoServicioProductos(db))
		montarVariantes(admin, usecase.NuevoServicioVariantes(db))
		montarClasificacion(admin, usecase.NuevoServicioClasificacion(db))
		montarImagenes(admin, usecase.NuevoServicioImagenes(db))
	})
}

func listar(s *usecase.Publico) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		f := repo.Filtros{Texto: strings.TrimSpace(q.Get("q")), Orden: q.Get("orden")}
		if f.Orden == "" {
			f.Orden = "nombre_asc"
		}
		if len(f.Texto) > 255 {
			apierr.Responder(w, "variantes", apierr.ErrValidacion{Campos: map[string]string{"q": "máximo 255 caracteres"}})
			return
		}
		for _, campo := range []string{"marca", "categoria"} {
			if q.Get(campo) == "" {
				continue
			}
			id, err := uuid.Parse(q.Get(campo))
			if err != nil {
				apierr.Responder(w, "variantes", apierr.ErrValidacion{Campos: map[string]string{campo: "uuid inválido"}})
				return
			}
			if campo == "marca" {
				f.MarcaID = &id
			} else {
				f.CategoriaID = &id
			}
		}
		pagina, err := entero(q.Get("pagina"), 1, 1, 1000000)
		if err != nil {
			apierr.Responder(w, "variantes", apierr.ErrValidacion{Campos: map[string]string{"pagina": "inválida"}})
			return
		}
		porPagina, err := entero(q.Get("porPagina"), 24, 1, 100)
		if err != nil {
			apierr.Responder(w, "variantes", apierr.ErrValidacion{Campos: map[string]string{"porPagina": "inválido"}})
			return
		}
		page, err := s.Listar(r.Context(), f, pagina, porPagina)
		if err != nil {
			apierr.Responder(w, "variantes", err)
			return
		}
		apierr.JSON(w, http.StatusOK, page)
	}
}

func detalle(s *usecase.Publico) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			apierr.Responder(w, "detalle variante", apierr.ErrValidacion{Campos: map[string]string{"id": "uuid inválido"}})
			return
		}
		item, err := s.Detalle(r.Context(), id)
		if err != nil {
			apierr.Responder(w, "detalle variante", err)
			return
		}
		apierr.JSON(w, http.StatusOK, item)
	}
}

func entero(value string, fallback, min, max int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < min || n > max {
		return 0, apierr.ErrValidacion{}
	}
	return n, nil
}
