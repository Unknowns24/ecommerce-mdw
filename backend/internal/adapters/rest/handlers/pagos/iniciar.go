package pagos

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/external/mp"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	reglas "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pagos"
)

const endpointIniciarPago = "POST /api/pedidos/{id}/pago"

// headerTokenAcceso lleva el token del enlace privado del invitado. Va en un
// header y no en la URL para que no quede en los logs de acceso.
const headerTokenAcceso = "X-Token-Acceso"

// iniciarPago abre una preferencia de Mercado Pago para un pedido pendiente
// de pago (H14). En orden:
//
//  1. ¿es del comprador?          sesión del dueño o header X-Token-Acceso; si no, 404.
//  2. ¿sigue esperando el pago?   medio, estado y reserva vigente; si no, 409.
//  3. ¿hay otro pago en curso?    se le pregunta a Mercado Pago; si hay, 409.
//  4. abrir la preferencia        con las líneas y el envío: el total del pedido.
func (h *handler) iniciarPago(w http.ResponseWriter, r *http.Request) {
	pedidoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierr.Responder(w, endpointIniciarPago, apierr.ErrNoEncontrado)
		return
	}

	var usuarioID *uuid.UUID
	if u, ok := middleware.UsuarioDeContexto(r.Context()); ok {
		usuarioID = &u.ID
	}

	ctx := r.Context()
	pedido, existe, err := pedidoPorID(ctx, h.db, pedidoID, false)
	if err != nil {
		apierr.Responder(w, endpointIniciarPago, err)
		return
	}
	if !existe {
		apierr.Responder(w, endpointIniciarPago, apierr.ErrNoEncontrado)
		return
	}
	if err := reglas.PuedeIniciarPago(pedido, usuarioID, r.Header.Get(headerTokenAcceso), time.Now()); err != nil {
		apierr.Responder(w, endpointIniciarPago, err)
		return
	}

	estados, err := h.mp.EstadosDePagosDelPedido(ctx, pedidoID.String())
	if err != nil {
		apierr.Responder(w, endpointIniciarPago, err)
		return
	}
	if reglas.HayPagoActivo(estados) {
		apierr.Responder(w, endpointIniciarPago, apierr.ErrRegla{
			Mensaje: "Ya hay un pago en curso para este pedido. Esperá su resultado antes de intentar de nuevo.",
		})
		return
	}

	lineas, err := lineasDePedido(ctx, h.db, pedidoID)
	if err != nil {
		apierr.Responder(w, endpointIniciarPago, err)
		return
	}

	respuesta, err := h.mp.CrearPreferencia(ctx, mp.Preferencia{
		PedidoID: pedidoID.String(),
		Items:    itemsDePreferencia(lineas, pedido.EnvioCentavos),
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

// itemsDePreferencia arma lo que Mercado Pago va a cobrar: las líneas del
// pedido y, si hay envío, una línea más. La suma es el total del pedido
// (total = subtotal + envío, lo exige un CHECK de la tabla), que es lo que el
// webhook compara con lo acreditado.
func itemsDePreferencia(lineas []lineaPedido, envioCentavos int64) []mp.ItemPreferencia {
	items := make([]mp.ItemPreferencia, 0, len(lineas)+1)
	for _, l := range lineas {
		items = append(items, mp.ItemPreferencia{
			Titulo:                 l.Nombre,
			Cantidad:               l.Unidades,
			PrecioUnitarioCentavos: l.PrecioUnitarioCentavos,
		})
	}
	if envioCentavos > 0 {
		items = append(items, mp.ItemPreferencia{Titulo: "Envío", Cantidad: 1, PrecioUnitarioCentavos: envioCentavos})
	}
	return items
}
