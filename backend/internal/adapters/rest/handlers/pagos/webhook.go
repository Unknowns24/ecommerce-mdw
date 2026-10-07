package pagos

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/pagos"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/external/mp"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	usecases "github.com/Unknowns24/ecommerce-mdw/internal/usecases/inventario"
)

const endpointWebhook = "POST /api/pagos/webhook"

type cuerpoWebhook struct {
	Type string `json:"type"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

// webhook recibe las notificaciones de Mercado Pago. Público a propósito:
// sin firma válida responde 401 y no procesa nada; con firma válida siempre
// responde 200, incluso si algo interno falla, para que Mercado Pago no
// reintente en loop una notificación que ya recibimos.
func (h *handler) webhook(w http.ResponseWriter, r *http.Request) {
	var cuerpo cuerpoWebhook
	_ = json.NewDecoder(r.Body).Decode(&cuerpo) // el cuerpo nunca decide el resultado, sólo puede aportar el id

	dataID := r.URL.Query().Get("data.id")
	if dataID == "" {
		dataID = cuerpo.Data.ID
	}

	if !mp.VerificarFirma(h.webhookSecret, r.Header.Get("x-signature"), r.Header.Get("x-request-id"), dataID) {
		apierr.Responder(w, endpointWebhook, apierr.ErrNoAutenticado)
		return
	}

	if cuerpo.Type != "" && cuerpo.Type != "payment" || dataID == "" {
		// Notificación de un tipo que todavía no manejamos (ej. merchant_order):
		// se confirma la recepción igual, no hay nada que procesar.
		apierr.JSON(w, http.StatusOK, nil)
		return
	}

	ctx := r.Context()
	pago, err := h.mp.ObtenerPago(ctx, dataID)
	if err != nil {
		// Nunca se confía en el cuerpo de la notificación para saber si está
		// pagado: si no se puede consultar al proveedor, no se confirma nada.
		// Se responde 200 igual para no provocar un reintento en loop por una
		// falla que un reintento inmediato no va a resolver.
		log.Printf("%s: no se pudo consultar el pago %s: %v", endpointWebhook, dataID, err)
		apierr.JSON(w, http.StatusOK, nil)
		return
	}

	pedidoID, err := uuid.Parse(pago.ReferenciaExterna)
	if err != nil {
		log.Printf("%s: referencia externa inválida en el pago %s: %q", endpointWebhook, dataID, pago.ReferenciaExterna)
		apierr.JSON(w, http.StatusOK, nil)
		return
	}

	if err := procesarNotificacion(ctx, h.db, h.stock, h.intentos, pedidoID, pago); err != nil {
		log.Printf("%s: no se pudo procesar el pago %s: %v", endpointWebhook, dataID, err)
	}

	apierr.JSON(w, http.StatusOK, nil)
}

// procesarNotificacion aplica el resultado ya verificado del pago. Es
// idempotente: una notificación repetida del mismo pago no duplica ni el
// movimiento de stock ni la confirmación del pedido.
//   - approved: si ya hay un intento APROBADO para esta referencia, no hace
//     nada; si no, confirma el intento, descuenta stock (ServicioStock.Confirmar
//     ya es idempotente por su cuenta) y confirma el pedido.
//   - cualquier otro estado: sólo deja registrado el intento con su estado.
func procesarNotificacion(ctx context.Context, db *gorm.DB, stock *usecases.ServicioStock, intentos *repo.RepositorioIntentos, pedidoID uuid.UUID, pago mp.Pago) error {
	estado := estadoIntentoDesdeMercadoPago(pago.Estado)

	existente, encontrado, err := intentos.PorReferenciaExterna(ctx, domain.ProveedorMercadoPago, pago.ID)
	if err != nil {
		return err
	}
	if encontrado && existente.Estado == domain.IntentoAprobado {
		return nil // ya procesado: la misma notificación dos veces deja el sistema igual
	}

	return db.Transaction(func(tx *gorm.DB) error {
		intentosTx := repo.NuevoRepositorioIntentos(tx)

		if encontrado {
			if err := intentosTx.CambiarEstado(ctx, existente.ID, estado); err != nil {
				return err
			}
		} else {
			nuevo := &domain.IntentoPago{
				ID:                uuid.New(),
				PedidoID:          pedidoID,
				Proveedor:         domain.ProveedorMercadoPago,
				ReferenciaExterna: pago.ID,
				Estado:            estado,
				MontoCentavos:     pago.MontoCentavos,
			}
			if err := intentosTx.Crear(ctx, nuevo); err != nil {
				return err
			}
		}

		if estado != domain.IntentoAprobado {
			return nil
		}

		if err := stock.Confirmar(ctx, tx, pedidoID); err != nil {
			return err
		}
		return confirmarPedido(ctx, tx, pedidoID)
	})
}

func estadoIntentoDesdeMercadoPago(estado string) domain.EstadoIntentoPago {
	switch estado {
	case "approved":
		return domain.IntentoAprobado
	case "rejected", "cancelled":
		return domain.IntentoRechazado
	default:
		return domain.IntentoPendiente
	}
}
