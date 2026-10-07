package catalogo

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/catalogo"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	usecase "github.com/Unknowns24/ecommerce-mdw/internal/usecases/catalogo"
)

func montarImagenes(r chi.Router, s *usecase.ServicioImagenes) {
	r.Post("/variantes/{id}/imagenes", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "agregar imagen", err)
			return
		}
		var in dto.CrearImagen
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "agregar imagen", err)
			return
		}
		out, err := s.Crear(req.Context(), id, in)
		if err != nil {
			apierr.Responder(w, "agregar imagen", err)
			return
		}
		apierr.JSON(w, http.StatusCreated, out)
	})
	r.Delete("/imagenes/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "eliminar imagen", err)
			return
		}
		if err := s.Eliminar(req.Context(), id); err != nil {
			apierr.Responder(w, "eliminar imagen", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	r.Put("/variantes/{id}/imagen-principal", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "imagen principal", err)
			return
		}
		var in dto.ImagenPrincipal
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "imagen principal", err)
			return
		}
		if err := s.DefinirPrincipal(req.Context(), id, in.ImagenID); err != nil {
			apierr.Responder(w, "imagen principal", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
