package mp

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// VerificarFirma valida la notificación de webhook según el esquema de
// Mercado Pago: el header x-signature trae "ts=<timestamp>,v1=<firma>"; la
// firma es un HMAC-SHA256 sobre el manifiesto "id:<dataID>;request-id:<x-
// request-id>;ts:<ts>;" con el secreto del webhook. Sin firma válida, la
// notificación no se procesa.
func VerificarFirma(secreto, xSignature, xRequestID, dataID string) bool {
	ts, v1 := partesDeFirma(xSignature)
	if ts == "" || v1 == "" {
		return false
	}

	manifiesto := "id:" + strings.ToLower(dataID) + ";request-id:" + xRequestID + ";ts:" + ts + ";"

	mac := hmac.New(sha256.New, []byte(secreto))
	mac.Write([]byte(manifiesto))
	esperada := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(esperada), []byte(v1)) == 1
}

func partesDeFirma(xSignature string) (ts, v1 string) {
	for _, parte := range strings.Split(xSignature, ",") {
		clave, valor, encontrado := strings.Cut(strings.TrimSpace(parte), "=")
		if !encontrado {
			continue
		}
		switch strings.TrimSpace(clave) {
		case "ts":
			ts = strings.TrimSpace(valor)
		case "v1":
			v1 = strings.TrimSpace(valor)
		}
	}
	return ts, v1
}
