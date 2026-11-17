package mp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestFirmaRechazaSecretoVacioAunqueLaFirmaCoincida(t *testing.T) {
	for _, secreto := range []string{"", "   "} {
		mac := hmac.New(sha256.New, []byte(secreto))
		mac.Write([]byte("id:123;request-id:prueba;ts:1;"))
		firma := "ts=1,v1=" + hex.EncodeToString(mac.Sum(nil))
		if VerificarFirma(secreto, firma, "prueba", "123") {
			t.Fatal("una clave ausente no autentica notificaciones")
		}
	}
}
