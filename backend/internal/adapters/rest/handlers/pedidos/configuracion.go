package pedidos

import (
	"net/http"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/pedidos"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

// ObtenerConfiguracionTienda atiende GET /api/tienda/configuracion (H13): si
// el efectivo está habilitado, hasta qué distancia se envía y cuánto cuesta
// cada tramo. Es público a propósito (no pide sesión): el checkout lo lee
// antes de que el comprador se identifique, y no expone nada personal. El
// costo que vale es el que calcula POST /api/pedidos, no este.
//
// Sin configuración cargada responde 404: el checkout lo lee como "el envío
// no está disponible".
func (h *Handler) ObtenerConfiguracionTienda(w http.ResponseWriter, r *http.Request) {
	const endpoint = "GET /api/tienda/configuracion"

	cfg, err := h.configuracion.Vigente(r.Context())
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.AConfiguracionTienda(cfg))
}
