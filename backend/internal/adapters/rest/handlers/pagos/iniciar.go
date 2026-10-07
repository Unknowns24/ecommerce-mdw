package pagos

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/external/mp"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

const endpointIniciarPago = "POST /api/pedidos/{id}/pago"

// iniciarPago abre una preferencia de Mercado Pago para un pedido pendiente
// de pago. Sesión opcional: el comprador invitado paga igual que el
// registrado, esta ruta no decide identidad, sólo valida el estado del
// pedido en el servidor.
func (h *handler) iniciarPago(w http.ResponseWriter, r *http.Request) {
	pedidoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierr.Responder(w, endpointIniciarPago, apierr.ErrNoEncontrado)
		return
	}

	ctx := r.Context()
	pedido, existe, err := pedidoPorID(ctx, h.db, pedidoID)
	if err != nil {
		apierr.Responder(w, endpointIniciarPago, err)
		return
	}
	if !existe {
		apierr.Responder(w, endpointIniciarPago, apierr.ErrNoEncontrado)
		return
	}
	if pedido.MedioPago != string(domain.ProveedorMercadoPago) || pedido.EstadoPago != "PENDIENTE" {
		apierr.Responder(w, endpointIniciarPago, apierr.ErrRegla{
			Mensaje: "Este pedido no está pendiente de pago con Mercado Pago",
		})
		return
	}

	lineas, err := lineasDePedido(ctx, h.db, pedidoID)
	if err != nil {
		apierr.Responder(w, endpointIniciarPago, err)
		return
	}

	items := make([]mp.ItemPreferencia, len(lineas))
	for i, l := range lineas {
		items[i] = mp.ItemPreferencia{
			Titulo:                 l.Nombre,
			Cantidad:               l.Unidades,
			PrecioUnitarioCentavos: l.PrecioUnitarioCentavos,
		}
	}

	respuesta, err := h.mp.CrearPreferencia(ctx, mp.Preferencia{
		PedidoID: pedidoID.String(),
		Items:    items,
	})
	if err != nil {
		apierr.Responder(w, endpointIniciarPago, err)
		return
	}

	intento := &domain.IntentoPago{
		ID:                uuid.New(),
		PedidoID:          pedidoID,
		Proveedor:         domain.ProveedorMercadoPago,
		ReferenciaExterna: respuesta.ID,
		Estado:            domain.IntentoPendiente,
		MontoCentavos:     pedido.TotalCentavos,
	}
	if err := h.intentos.Crear(ctx, intento); err != nil {
		apierr.Responder(w, endpointIniciarPago, err)
		return
	}

	apierr.JSON(w, http.StatusCreated, map[string]string{"initPoint": respuesta.InitPoint})
}
