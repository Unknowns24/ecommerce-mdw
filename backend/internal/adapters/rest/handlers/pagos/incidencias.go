package pagos

import (
	"encoding/json"
	"net/http"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/pagos"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *handler) listarIncidencias(w http.ResponseWriter, r *http.Request) {
	const endpoint = "GET /api/admin/pedidos/{id}/incidencias"
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierr.Responder(w, endpoint, apierr.ErrNoEncontrado)
		return
	}
	items, err := h.incidencias.Listar(r.Context(), id)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.AIncidencias(items))
}

func (h *handler) resolverIncidencia(w http.ResponseWriter, r *http.Request) {
	const endpoint = "POST /api/admin/pedidos/{id}/incidencias/{incidenciaID}/resolucion"
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	incidencia, errIncidencia := uuid.Parse(chi.URLParam(r, "incidenciaID"))
	if err != nil || errIncidencia != nil {
		apierr.Responder(w, endpoint, apierr.ErrNoEncontrado)
		return
	}
	actor, ok := middleware.UsuarioDeContexto(r.Context())
	if !ok {
		apierr.Responder(w, endpoint, apierr.ErrNoAutenticado)
		return
	}
	var cuerpo struct {
		Nota string `json:"nota"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := json.NewDecoder(r.Body).Decode(&cuerpo); err != nil {
		apierr.Responder(w, endpoint, apierr.ErrValidacion{Campos: map[string]string{"nota": "Enviá una nota de resolución válida."}})
		return
	}
	if err := h.incidencias.Resolver(r.Context(), id, incidencia, actor.ID, cuerpo.Nota); err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, map[string]string{"estado": "RESUELTA"})
}
