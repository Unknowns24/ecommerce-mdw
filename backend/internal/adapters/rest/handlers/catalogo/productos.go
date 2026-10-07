package catalogo

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/catalogo"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	usecase "github.com/Unknowns24/ecommerce-mdw/internal/usecases/catalogo"
)

func montarProductos(r chi.Router, s *usecase.ServicioProductos) {
	r.Post("/productos", func(w http.ResponseWriter, req *http.Request) {
		var in dto.CrearProducto
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "crear producto", err)
			return
		}
		out, err := s.Crear(req.Context(), in)
		if err != nil {
			apierr.Responder(w, "crear producto", err)
			return
		}
		apierr.JSON(w, http.StatusCreated, out)
	})
	r.Get("/productos", func(w http.ResponseWriter, req *http.Request) {
		pagina, err := entero(req.URL.Query().Get("pagina"), 1, 1, 1000000)
		if err != nil {
			apierr.Responder(w, "listar productos", apierr.ErrValidacion{Campos: map[string]string{"pagina": "inválida"}})
			return
		}
		porPagina, err := entero(req.URL.Query().Get("porPagina"), 24, 1, 100)
		if err != nil {
			apierr.Responder(w, "listar productos", apierr.ErrValidacion{Campos: map[string]string{"porPagina": "inválido"}})
			return
		}
		out, err := s.Listar(req.Context(), pagina, porPagina)
		if err != nil {
			apierr.Responder(w, "listar productos", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Get("/productos/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "producto", err)
			return
		}
		out, err := s.PorID(req.Context(), id)
		if err != nil {
			apierr.Responder(w, "producto", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Patch("/productos/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "editar producto", err)
			return
		}
		var in dto.EditarProducto
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "editar producto", err)
			return
		}
		out, err := s.Editar(req.Context(), id, in)
		if err != nil {
			apierr.Responder(w, "editar producto", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Delete("/productos/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "desactivar producto", err)
			return
		}
		if err := s.Desactivar(req.Context(), id); err != nil {
			apierr.Responder(w, "desactivar producto", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func idDeRuta(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apierr.ErrValidacion{Campos: map[string]string{"id": "uuid inválido"}}
	}
	return id, nil
}

func leerJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	d := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(dst); err != nil {
		return apierr.ErrValidacion{Campos: map[string]string{"body": err.Error()}}
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return apierr.ErrValidacion{Campos: map[string]string{"body": "un único objeto JSON"}}
	}
	return nil
}
