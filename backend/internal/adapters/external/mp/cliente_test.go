package mp

import (
	"encoding/json"
	"testing"
	"time"
)

// Issue #14: la preferencia vence junto con la reserva del pedido, en el
// formato de fecha que acepta Mercado Pago.
func TestCuerpoPreferenciaVenceConLaReserva(t *testing.T) {
	c := NuevoCliente("token", "https://mdw.unkcode.com")
	vence := time.Date(2026, 11, 5, 20, 15, 0, 0, time.FixedZone("ART", -3*60*60))

	bruto, err := json.Marshal(c.cuerpoPreferencia(Preferencia{PedidoID: "p1", VenceEn: &vence}))
	if err != nil {
		t.Fatal(err)
	}
	var cuerpo map[string]any
	if err := json.Unmarshal(bruto, &cuerpo); err != nil {
		t.Fatal(err)
	}
	if cuerpo["expires"] != true {
		t.Fatalf("expires = %v, esperado true", cuerpo["expires"])
	}
	if got := cuerpo["expiration_date_to"]; got != "2026-11-05T20:15:00.000-03:00" {
		t.Fatalf("expiration_date_to = %v, esperado 2026-11-05T20:15:00.000-03:00", got)
	}
}

func TestCuerpoPreferenciaSinPlazo(t *testing.T) {
	c := NuevoCliente("token", "https://mdw.unkcode.com")

	bruto, err := json.Marshal(c.cuerpoPreferencia(Preferencia{PedidoID: "p1"}))
	if err != nil {
		t.Fatal(err)
	}
	var cuerpo map[string]any
	if err := json.Unmarshal(bruto, &cuerpo); err != nil {
		t.Fatal(err)
	}
	if _, ok := cuerpo["expires"]; ok {
		t.Fatalf("sin VenceEn no se manda expires: %s", bruto)
	}
	if _, ok := cuerpo["expiration_date_to"]; ok {
		t.Fatalf("sin VenceEn no se manda expiration_date_to: %s", bruto)
	}
}
