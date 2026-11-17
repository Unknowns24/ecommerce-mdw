package pagos

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/external/mp"
)

type transporteFallido struct{}

func (transporteFallido) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("proveedor no disponible")
}

func TestWebhookPermiteReintentoSiElProveedorFalla(t *testing.T) {
	anterior := http.DefaultTransport
	http.DefaultTransport = transporteFallido{}
	t.Cleanup(func() { http.DefaultTransport = anterior })
	mac := hmac.New(sha256.New, []byte("secreto-de-prueba"))
	mac.Write([]byte("id:123;request-id:req;ts:1700000000;"))
	r := httptest.NewRequest(http.MethodPost, "/api/pagos/webhook?data.id=123", nil)
	r.Header.Set("x-signature", "ts=1700000000,v1="+hex.EncodeToString(mac.Sum(nil)))
	r.Header.Set("x-request-id", "req")
	w := httptest.NewRecorder()
	h := handler{mp: mp.NuevoCliente("token-de-prueba", "http://localhost"), webhookSecret: "secreto-de-prueba"}
	h.webhook(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, se necesita 500 para permitir reintentos", w.Code)
	}
}
