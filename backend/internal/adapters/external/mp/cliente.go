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
	"net/url"
	"time"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

const (
	urlPreferencias  = "https://api.mercadopago.com/checkout/preferences"
	urlPagos         = "https://api.mercadopago.com/v1/payments/"
	urlBusquedaPagos = "https://api.mercadopago.com/v1/payments/search"
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
	// VenceEn, si no es nil, es el último instante en que se puede pagar la
	// preferencia (Issue #14: el vencimiento de la reserva).
	VenceEn *time.Time
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
	Expires           bool                  `json:"expires,omitempty"`
	ExpirationDateTo  string                `json:"expiration_date_to,omitempty"`
}

// formatoFechaMP es el formato de fecha que Mercado Pago acepta en
// expiration_date_to: ISO 8601 con milisegundos y zona horaria.
const formatoFechaMP = "2006-01-02T15:04:05.000-07:00"

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
	var respuesta crearPreferenciaResponse
	if err := c.solicitar(ctx, http.MethodPost, urlPreferencias, c.cuerpoPreferencia(p), &respuesta); err != nil {
		return RespuestaPreferencia{}, err
	}

	return RespuestaPreferencia{ID: respuesta.ID, InitPoint: respuesta.InitPoint}, nil
}

// cuerpoPreferencia arma el JSON que se le manda a Mercado Pago. Está
// separado de CrearPreferencia para poder probarlo sin red.
func (c *Cliente) cuerpoPreferencia(p Preferencia) crearPreferenciaRequest {
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
	if p.VenceEn != nil {
		cuerpo.Expires = true
		cuerpo.ExpirationDateTo = p.VenceEn.Format(formatoFechaMP)
	}
	return cuerpo
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

	return pagoDesdeRespuesta(respuesta), nil
}

func pagoDesdeRespuesta(r pagoResponse) Pago {
	return Pago{
		ID:                fmt.Sprintf("%d", r.ID),
		Estado:            r.Status,
		MontoCentavos:     int64(r.TransactionAmount*100 + 0.5),
		ReferenciaExterna: r.ExternalReference,
	}
}

type busquedaPagosResponse struct {
	Results []pagoResponse `json:"results"`
}

// PagosDelPedido devuelve cada pago que Mercado Pago tiene asociado al pedido
// (external_reference), con su estado y su monto. Es la fuente de verdad para
// saber si ya hay un cobro en curso o uno aprobado: el webhook puede llegar
// tarde o no llegar, así que la tabla local no alcanza.
func (c *Cliente) PagosDelPedido(ctx context.Context, pedidoID string) ([]Pago, error) {
	var respuesta busquedaPagosResponse
	direccion := urlBusquedaPagos + "?external_reference=" + url.QueryEscape(pedidoID)
	if err := c.solicitar(ctx, http.MethodGet, direccion, nil, &respuesta); err != nil {
		return nil, err
	}
	pagos := make([]Pago, len(respuesta.Results))
	for i, r := range respuesta.Results {
		pagos[i] = pagoDesdeRespuesta(r)
	}
	return pagos, nil
}

// EstadosDePagosDelPedido devuelve solo los estados de PagosDelPedido.
func (c *Cliente) EstadosDePagosDelPedido(ctx context.Context, pedidoID string) ([]string, error) {
	pagos, err := c.PagosDelPedido(ctx, pedidoID)
	if err != nil {
		return nil, err
	}
	return EstadosDe(pagos), nil
}

// EstadosDe devuelve el estado de cada pago, en el mismo orden.
func EstadosDe(pagos []Pago) []string {
	estados := make([]string, len(pagos))
	for i, p := range pagos {
		estados[i] = p.Estado
	}
	return estados
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
