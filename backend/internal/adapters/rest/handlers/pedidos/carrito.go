package pedidos

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/pedidos"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

// Carrito del cliente registrado (SPEC-H07). El del invitado vive en el
// navegador (localStorage) y nunca llega al backend: para comprar, el invitado
// manda sus items en el body del checkout. Por eso estas rutas exigen una
// sesión: no hay "carrito de nadie" guardado en el servidor.
//
// Ninguna acepta precios: el carrito se valoriza con el catálogo vigente en
// cada lectura (no congela precios ni reserva unidades).

// usuarioDelCarrito toma la identidad del contexto de la sesión (nunca del
// body ni de la query). Sin sesión: 401.
func (h *Handler) usuarioDelCarrito(w http.ResponseWriter, r *http.Request, endpoint string) (uuid.UUID, bool) {
	u, ok := h.usuario(r.Context())
	if !ok {
		apierr.Responder(w, endpoint, apierr.ErrNoAutenticado)
		return uuid.Nil, false
	}
	return u.ID, true
}

// VerCarrito atiende GET /api/carrito.
func (h *Handler) VerCarrito(w http.ResponseWriter, r *http.Request) {
	const endpoint = "GET /api/carrito"
	usuarioID, ok := h.usuarioDelCarrito(w, r, endpoint)
	if !ok {
		return
	}
	v, err := h.carrito.Ver(r.Context(), usuarioID)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.ACarrito(v))
}

// AgregarAlCarrito atiende POST /api/carrito/items: agrega una variante o, si
// ya está, suma unidades. 201 si la línea es nueva, 200 si solo sumó. Pedir más
// que el disponible es 409 (no 400): el request está bien, es el estado del
// sistema el que no lo permite.
func (h *Handler) AgregarAlCarrito(w http.ResponseWriter, r *http.Request) {
	const endpoint = "POST /api/carrito/items"
	usuarioID, ok := h.usuarioDelCarrito(w, r, endpoint)
	if !ok {
		return
	}

	var req dto.AgregarItemRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.Responder(w, endpoint, apierr.ErrValidacion{Campos: map[string]string{"body": "El cuerpo debe ser un JSON válido"}})
		return
	}
	varianteID, unidades, err := req.AEntrada()
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}

	v, nueva, err := h.carrito.AgregarItem(r.Context(), usuarioID, varianteID, unidades)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	estado := http.StatusOK
	if nueva {
		estado = http.StatusCreated
	}
	apierr.JSON(w, estado, dto.ACarrito(v))
}

// CambiarUnidadesDelCarrito atiende PATCH /api/carrito/items/{id}: fija las
// unidades de una línea. Una línea que no es de tu carrito no existe: 404.
func (h *Handler) CambiarUnidadesDelCarrito(w http.ResponseWriter, r *http.Request) {
	const endpoint = "PATCH /api/carrito/items/{id}"
	usuarioID, ok := h.usuarioDelCarrito(w, r, endpoint)
	if !ok {
		return
	}
	itemID, err := idDeRuta(r)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}

	var req dto.CambiarUnidadesRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.Responder(w, endpoint, apierr.ErrValidacion{Campos: map[string]string{"body": "El cuerpo debe ser un JSON válido"}})
		return
	}
	unidades, err := req.AEntrada()
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}

	v, err := h.carrito.CambiarUnidades(r.Context(), usuarioID, itemID, unidades)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.ACarrito(v))
}

// QuitarDelCarrito atiende DELETE /api/carrito/items/{id}.
func (h *Handler) QuitarDelCarrito(w http.ResponseWriter, r *http.Request) {
	const endpoint = "DELETE /api/carrito/items/{id}"
	usuarioID, ok := h.usuarioDelCarrito(w, r, endpoint)
	if !ok {
		return
	}
	itemID, err := idDeRuta(r)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	if err := h.carrito.QuitarItem(r.Context(), usuarioID, itemID); err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// VaciarCarrito atiende DELETE /api/carrito.
func (h *Handler) VaciarCarrito(w http.ResponseWriter, r *http.Request) {
	const endpoint = "DELETE /api/carrito"
	usuarioID, ok := h.usuarioDelCarrito(w, r, endpoint)
	if !ok {
		return
	}
	if err := h.carrito.Vaciar(r.Context(), usuarioID); err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
