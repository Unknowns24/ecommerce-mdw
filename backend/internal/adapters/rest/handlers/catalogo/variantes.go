package catalogo

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/catalogo"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	usecase "github.com/Unknowns24/ecommerce-mdw/internal/usecases/catalogo"
)

func montarVariantes(r chi.Router, s *usecase.ServicioVariantes) {
	r.Post("/variantes", func(w http.ResponseWriter, req *http.Request) {
		var in dto.CrearVariante
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "crear variante", err)
			return
		}
		out, err := s.Crear(req.Context(), in)
		if err != nil {
			apierr.Responder(w, "crear variante", err)
			return
		}
		apierr.JSON(w, http.StatusCreated, out)
	})
	r.Get("/variantes", func(w http.ResponseWriter, req *http.Request) {
		pagina, err := entero(req.URL.Query().Get("pagina"), 1, 1, 1000000)
		if err != nil {
			apierr.Responder(w, "listar variantes admin", apierr.ErrValidacion{Campos: map[string]string{"pagina": "inválida"}})
			return
		}
		porPagina, err := entero(req.URL.Query().Get("porPagina"), 24, 1, 100)
		if err != nil {
			apierr.Responder(w, "listar variantes admin", apierr.ErrValidacion{Campos: map[string]string{"porPagina": "inválido"}})
			return
		}
		out, err := s.Listar(req.Context(), pagina, porPagina)
		if err != nil {
			apierr.Responder(w, "listar variantes admin", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Get("/variantes/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "variante admin", err)
			return
		}
		out, err := s.PorID(req.Context(), id)
		if err != nil {
			apierr.Responder(w, "variante admin", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Patch("/variantes/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "editar variante", err)
			return
		}
		var in dto.EditarVariante
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "editar variante", err)
			return
		}
		out, err := s.Editar(req.Context(), id, in)
		if err != nil {
			apierr.Responder(w, "editar variante", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Delete("/variantes/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "desactivar variante", err)
			return
		}
		if err := s.Desactivar(req.Context(), id); err != nil {
			apierr.Responder(w, "desactivar variante", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
