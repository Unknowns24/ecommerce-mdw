package pedidos

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/pedidos"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

// Los endpoints de este archivo son de administración y se montan bajo
// mw.RequierePermiso("pedidos.gestionar") (ver Montar).
//
// Fijate en la forma de las rutas: el verbo se convierte en sustantivo
// (POST .../despacho), no es PATCH /api/admin/pedidos/{id} con
// {"estado":"DESPACHADO"}. Con un PATCH, el CLIENTE decide la transición y el
// servidor obedece: nada impediría saltarse un paso. Con el sub-recurso, el
// cliente pide que OCURRA la operación y el SERVIDOR decide si corresponde
// (estados.go). Además cada operación tiene su propia regla y sus propios
// errores. Ninguno de estos endpoints acepta productos, cantidades, precios ni
// domicilio: la copia histórica del pedido es inmutable.

// ListarPedidosAdmin atiende GET /api/admin/pedidos?estado=&desde=&hasta=&pagina=&tamano=
func (h *Handler) ListarPedidosAdmin(w http.ResponseWriter, r *http.Request) {
	const endpoint = "GET /api/admin/pedidos"

	q := r.URL.Query()
	filtros, err := dto.FiltrosDesdeQuery(q.Get("estado"), q.Get("desde"), q.Get("hasta"))
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	pagina, err := enteroDeQuery(r, "pagina")
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	tamano, err := enteroDeQuery(r, "tamano")
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}

	res, err := h.gestion.Listar(r.Context(), filtros, pagina, tamano)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.APaginaAdmin(res))
}

// ObtenerPedidoAdmin atiende GET /api/admin/pedidos/{id}: detalle completo con
// comprador y pago.
func (h *Handler) ObtenerPedidoAdmin(w http.ResponseWriter, r *http.Request) {
	const endpoint = "GET /api/admin/pedidos/{id}"
	id, err := idDeRuta(r)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	p, err := h.gestion.Detalle(r.Context(), id)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.APedidoAdmin(p))
}

// DespacharPedido atiende POST /api/admin/pedidos/{id}/despacho.
func (h *Handler) DespacharPedido(w http.ResponseWriter, r *http.Request) {
	h.transicion(w, r, "POST /api/admin/pedidos/{id}/despacho", h.gestion.Despachar)
}

// EntregarPedido atiende POST /api/admin/pedidos/{id}/entrega.
func (h *Handler) EntregarPedido(w http.ResponseWriter, r *http.Request) {
	h.transicion(w, r, "POST /api/admin/pedidos/{id}/entrega", h.gestion.Entregar)
}

// CobrarPedidoEnEfectivo atiende POST /api/admin/pedidos/{id}/cobro-efectivo:
// registra cobro y entrega en una sola acción.
func (h *Handler) CobrarPedidoEnEfectivo(w http.ResponseWriter, r *http.Request) {
	h.transicion(w, r, "POST /api/admin/pedidos/{id}/cobro-efectivo", h.gestion.CobrarEnEfectivo)
}

// CancelarPedido atiende POST /api/admin/pedidos/{id}/cancelacion. El body
// trae solo el motivo; el responsable sale de la sesión, nunca del body. Si el
// pedido estaba pagado, la respuesta incluye el aviso de que el reintegro se
// hace fuera del sistema.
func (h *Handler) CancelarPedido(w http.ResponseWriter, r *http.Request) {
	const endpoint = "POST /api/admin/pedidos/{id}/cancelacion"

	usuario, ok := h.usuario(r.Context())
	if !ok {
		apierr.Responder(w, endpoint, apierr.ErrNoAutenticado)
		return
	}
	id, err := idDeRuta(r)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}

	var req dto.CancelarRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.Responder(w, endpoint, apierr.ErrValidacion{Campos: map[string]string{"motivo": "Indicá el motivo de la cancelación"}})
		return
	}
	motivo, err := req.AMotivo()
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}

	res, err := h.gestion.Cancelar(r.Context(), id, usuario.ID, motivo)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.ACancelacion(res))
}

// transicion es el esqueleto de las operaciones sin body: lee el id de la ruta,
// pide la operación y responde con el pedido actualizado.
func (h *Handler) transicion(
	w http.ResponseWriter, r *http.Request, endpoint string,
	operacion func(ctx context.Context, id uuid.UUID) (domain.Pedido, error),
) {
	id, err := idDeRuta(r)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	p, err := operacion(r.Context(), id)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.APedidoAdmin(p))
}

// idDeRuta lee {id}. Un id que no es un uuid no existe: 404, igual que en la
// consulta propia.
func idDeRuta(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apierr.ErrNoEncontrado
	}
	return id, nil
}
