// Package mp is the Mercado Pago adapter: it talks to their checkout API
// over plain net/http (no SDK, less dependencias y más fácil de defender) y
// verifica la firma de sus notificaciones de webhook.
package mp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

const (
	urlPreferencias = "https://api.mercadopago.com/checkout/preferences"
	urlPagos        = "https://api.mercadopago.com/v1/payments/"
)

// Cliente talks to the Mercado Pago sandbox/production API with a 10 second
// timeout: a provider that doesn't answer must not hang the checkout.
type Cliente struct {
	accessToken string
	baseURL     string
	http        *http.Client
}

func NuevoCliente(accessToken, baseURL string) *Cliente {
	return &Cliente{
		accessToken: accessToken,
		baseURL:     baseURL,
		http:        &http.Client{Timeout: 10 * time.Second},
	}
}

// ItemPreferencia is one line of the checkout preference.
type ItemPreferencia struct {
	Titulo                 string
	Cantidad               int
	PrecioUnitarioCentavos int64
}

// Preferencia is what the caller needs to decide; Cliente builds the
// back_urls and notification_url itself from baseURL, so ningún otro
// paquete arma esas URLs a mano.
type Preferencia struct {
	PedidoID string
	Items    []ItemPreferencia
}

// RespuestaPreferencia is what the checkout needs to redirect the comprador.
type RespuestaPreferencia struct {
	ID        string
	InitPoint string
}

// Pago is the payment state as Mercado Pago reports it, consultado por id —
// nunca confiado desde el cuerpo de una notificación.
type Pago struct {
	ID                string
	Estado            string
	MontoCentavos     int64
	ReferenciaExterna string
}

type itemPreferenciaJSON struct {
	Title     string  `json:"title"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type crearPreferenciaRequest struct {
	Items             []itemPreferenciaJSON `json:"items"`
	ExternalReference string                `json:"external_reference"`
	BackURLs          backURLs              `json:"back_urls"`
	NotificationURL   string                `json:"notification_url"`
}

type backURLs struct {
	Success string `json:"success"`
	Failure string `json:"failure"`
	Pending string `json:"pending"`
}

type crearPreferenciaResponse struct {
	ID        string `json:"id"`
	InitPoint string `json:"init_point"`
}

// CrearPreferencia abre una preferencia de pago. Si Mercado Pago no responde
// a tiempo o rechaza la request, el pedido no queda marcado como pagado: se
// devuelve apierr.ErrRegla para que el handler lo traduzca a 409.
func (c *Cliente) CrearPreferencia(ctx context.Context, p Preferencia) (RespuestaPreferencia, error) {
	items := make([]itemPreferenciaJSON, len(p.Items))
	for i, item := range p.Items {
		items[i] = itemPreferenciaJSON{
			Title:    item.Titulo,
			Quantity: item.Cantidad,
			// Mercado Pago exige el precio unitario en la unidad mayor de la
			// moneda (pesos con decimales), nunca en centavos: esta es la
			// única conversión del sistema, exigida por el contrato externo,
			// y no se usa en ningún cálculo interno (ADR-004).
			UnitPrice: float64(item.PrecioUnitarioCentavos) / 100.0,
		}
	}

	cuerpo := crearPreferenciaRequest{
		Items:             items,
		ExternalReference: p.PedidoID,
		BackURLs: backURLs{
			Success: c.baseURL + "/pago/retorno?estado=aprobado",
			Failure: c.baseURL + "/pago/retorno?estado=rechazado",
			Pending: c.baseURL + "/pago/retorno?estado=pendiente",
		},
		NotificationURL: c.baseURL + "/api/pagos/webhook",
	}

	var respuesta crearPreferenciaResponse
	if err := c.solicitar(ctx, http.MethodPost, urlPreferencias, cuerpo, &respuesta); err != nil {
		return RespuestaPreferencia{}, err
	}

	return RespuestaPreferencia{ID: respuesta.ID, InitPoint: respuesta.InitPoint}, nil
}

type pagoResponse struct {
	ID                int64   `json:"id"`
	Status            string  `json:"status"`
	TransactionAmount float64 `json:"transaction_amount"`
	ExternalReference string  `json:"external_reference"`
}

// ObtenerPago consulta el estado real de un pago por su id. El webhook nunca
// confía en el cuerpo de la notificación: sólo lee el payment_id y le
// pregunta a esta función.
func (c *Cliente) ObtenerPago(ctx context.Context, pagoID string) (Pago, error) {
	var respuesta pagoResponse
	if err := c.solicitar(ctx, http.MethodGet, urlPagos+pagoID, nil, &respuesta); err != nil {
		return Pago{}, err
	}

	return Pago{
		ID:                fmt.Sprintf("%d", respuesta.ID),
		Estado:            respuesta.Status,
		MontoCentavos:     int64(respuesta.TransactionAmount*100 + 0.5),
		ReferenciaExterna: respuesta.ExternalReference,
	}, nil
}

func (c *Cliente) solicitar(ctx context.Context, metodo, url string, cuerpo any, destino any) error {
	var lector io.Reader
	if cuerpo != nil {
		bruto, err := json.Marshal(cuerpo)
		if err != nil {
			return fmt.Errorf("mp: codificar request: %w", err)
		}
		lector = bytes.NewReader(bruto)
	}

	req, err := http.NewRequestWithContext(ctx, metodo, url, lector)
	if err != nil {
		return fmt.Errorf("mp: construir request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		// Timeout o proveedor caído: la rúbrica llama a esto "manejar la
		// falla del proveedor y comunicarla al usuario".
		return apierr.ErrRegla{Mensaje: "No pudimos iniciar el pago, probá de nuevo en unos minutos"}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apierr.ErrRegla{Mensaje: "No pudimos iniciar el pago, probá de nuevo en unos minutos"}
	}

	if destino == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(destino); err != nil {
		return fmt.Errorf("mp: decodificar respuesta: %w", err)
	}
	return nil
}
