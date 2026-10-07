package pagos

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
)

const endpointEstadoPago = "GET /api/pedidos/{id}/pago"

// estadoPago devuelve el último intento de pago de un pedido propio. La
// pertenencia se valida en el WHERE de la consulta al pedido, no con un if
// después de traer la fila.
func (h *handler) estadoPago(w http.ResponseWriter, r *http.Request) {
	pedidoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierr.Responder(w, endpointEstadoPago, apierr.ErrNoEncontrado)
		return
	}

	usuario, ok := middleware.UsuarioDeContexto(r.Context())
	if !ok {
		apierr.Responder(w, endpointEstadoPago, apierr.ErrNoAutenticado)
		return
	}

	ctx := r.Context()
	var propio bool
	if err := h.db.WithContext(ctx).
		Raw("SELECT EXISTS(SELECT 1 FROM pedido WHERE id = ? AND usuario_id = ?)", pedidoID, usuario.ID).
		Scan(&propio).Error; err != nil {
		apierr.Responder(w, endpointEstadoPago, err)
		return
	}
	if !propio {
		apierr.Responder(w, endpointEstadoPago, apierr.ErrNoEncontrado)
		return
	}

	intento, err := h.intentos.UltimoPorPedido(ctx, pedidoID)
	if err != nil {
		apierr.Responder(w, endpointEstadoPago, apierr.ErrNoEncontrado)
		return
	}

	apierr.JSON(w, http.StatusOK, map[string]any{
		"estado":        intento.Estado,
		"proveedor":     intento.Proveedor,
		"montoCentavos": intento.MontoCentavos,
		"creadoEn":      intento.CreadoEn,
	})
}
