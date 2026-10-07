package catalogo

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/catalogo"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	usecase "github.com/Unknowns24/ecommerce-mdw/internal/usecases/catalogo"
)

func montarClasificacion(r chi.Router, s *usecase.ServicioClasificacion) {
	r.Post("/marcas", func(w http.ResponseWriter, req *http.Request) {
		var in dto.CrearMarca
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "crear marca", err)
			return
		}
		out, err := s.CrearMarca(req.Context(), in.Nombre)
		if err != nil {
			apierr.Responder(w, "crear marca", err)
			return
		}
		apierr.JSON(w, http.StatusCreated, out)
	})
	r.Get("/marcas", func(w http.ResponseWriter, req *http.Request) {
		pagina, porPagina, err := parametrosPagina(req)
		if err != nil {
			apierr.Responder(w, "listar marcas", err)
			return
		}
		out, err := s.ListarMarcas(req.Context(), pagina, porPagina)
		if err != nil {
			apierr.Responder(w, "listar marcas", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Get("/marcas/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "marca", err)
			return
		}
		out, err := s.Marca(req.Context(), id)
		if err != nil {
			apierr.Responder(w, "marca", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Patch("/marcas/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "editar marca", err)
			return
		}
		var in dto.EditarMarca
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "editar marca", err)
			return
		}
		out, err := s.EditarMarca(req.Context(), id, in.Nombre)
		if err != nil {
			apierr.Responder(w, "editar marca", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Delete("/marcas/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "desactivar marca", err)
			return
		}
		if err := s.DesactivarMarca(req.Context(), id); err != nil {
			apierr.Responder(w, "desactivar marca", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	r.Post("/categorias", func(w http.ResponseWriter, req *http.Request) {
		var in dto.CrearCategoria
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "crear categoria", err)
			return
		}
		out, err := s.CrearCategoria(req.Context(), in.Nombre, in.PadreID)
		if err != nil {
			apierr.Responder(w, "crear categoria", err)
			return
		}
		apierr.JSON(w, http.StatusCreated, out)
	})
	r.Get("/categorias", func(w http.ResponseWriter, req *http.Request) {
		pagina, porPagina, err := parametrosPagina(req)
		if err != nil {
			apierr.Responder(w, "listar categorias", err)
			return
		}
		out, err := s.ListarCategorias(req.Context(), pagina, porPagina)
		if err != nil {
			apierr.Responder(w, "listar categorias", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Get("/categorias/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "categoria", err)
			return
		}
		out, err := s.Categoria(req.Context(), id)
		if err != nil {
			apierr.Responder(w, "categoria", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Patch("/categorias/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "editar categoria", err)
			return
		}
		var in dto.EditarCategoria
		if err := leerJSON(req, &in); err != nil {
			apierr.Responder(w, "editar categoria", err)
			return
		}
		var padre **uuid.UUID
		if len(in.PadreID) > 0 {
			padre = new(*uuid.UUID)
			if !bytes.Equal(in.PadreID, []byte("null")) {
				var id uuid.UUID
				if err := json.Unmarshal(in.PadreID, &id); err != nil {
					apierr.Responder(w, "editar categoria", apierr.ErrValidacion{Campos: map[string]string{"padreId": "uuid inválido"}})
					return
				}
				*padre = &id
			}
		}
		out, err := s.EditarCategoria(req.Context(), id, in.Nombre, padre)
		if err != nil {
			apierr.Responder(w, "editar categoria", err)
			return
		}
		apierr.JSON(w, http.StatusOK, out)
	})
	r.Delete("/categorias/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, err := idDeRuta(req)
		if err != nil {
			apierr.Responder(w, "desactivar categoria", err)
			return
		}
		if err := s.DesactivarCategoria(req.Context(), id); err != nil {
			apierr.Responder(w, "desactivar categoria", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func parametrosPagina(r *http.Request) (int, int, error) {
	pagina, err := entero(r.URL.Query().Get("pagina"), 1, 1, 1000000)
	if err != nil {
		return 0, 0, apierr.ErrValidacion{Campos: map[string]string{"pagina": "inválida"}}
	}
	porPagina, err := entero(r.URL.Query().Get("porPagina"), 24, 1, 100)
	if err != nil {
		return 0, 0, apierr.ErrValidacion{Campos: map[string]string{"porPagina": "inválido"}}
	}
	return pagina, porPagina, nil
}
