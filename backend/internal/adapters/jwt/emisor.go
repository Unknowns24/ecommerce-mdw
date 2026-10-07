package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrToken = errors.New("token inválido")

type Claims struct {
	Sub      uuid.UUID `json:"sub"`
	Correo   string    `json:"correo"`
	Roles    []string  `json:"roles"`
	Permisos []string  `json:"permisos"`
	Exp      int64     `json:"exp"`
	Iat      int64     `json:"iat"`
}

type Emisor struct {
	secreto  []byte
	duracion time.Duration
}

func NuevoEmisor(secreto string, duracion time.Duration) *Emisor {
	return &Emisor{[]byte(secreto), duracion}
}

func (e *Emisor) firma(mensaje string) []byte {
	mac := hmac.New(sha256.New, e.secreto)
	_, _ = mac.Write([]byte(mensaje))
	return mac.Sum(nil)
}

func (e *Emisor) Emitir(u Claims) (string, error) {
	if len(e.secreto) == 0 || e.duracion <= 0 || u.Sub == uuid.Nil {
		return "", ErrToken
	}
	u.Iat = time.Now().Unix()
	u.Exp = time.Now().Add(e.duracion).Unix()
	cuerpo, err := json.Marshal(u)
	if err != nil {
		return "", err
	}
	enc := base64.RawURLEncoding
	mensaje := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." + enc.EncodeToString(cuerpo)
	return mensaje + "." + enc.EncodeToString(e.firma(mensaje)), nil
}

func (e *Emisor) Verificar(token string) (Claims, error) {
	var u Claims
	if len(e.secreto) == 0 {
		return u, ErrToken
	}
	partes := strings.Split(token, ".")
	if len(partes) != 3 {
		return u, ErrToken
	}
	enc := base64.RawURLEncoding
	cabecera, err := enc.DecodeString(partes[0])
	if err != nil {
		return u, ErrToken
	}
	var h struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if json.Unmarshal(cabecera, &h) != nil || h.Alg != "HS256" || h.Typ != "JWT" {
		return u, ErrToken
	}
	firma, err := enc.DecodeString(partes[2])
	if err != nil || !hmac.Equal(firma, e.firma(partes[0]+"."+partes[1])) {
		return u, ErrToken
	}
	cuerpo, err := enc.DecodeString(partes[1])
	if err != nil || json.Unmarshal(cuerpo, &u) != nil {
		return Claims{}, ErrToken
	}
	ahora := time.Now().Unix()
	if u.Sub == uuid.Nil || u.Exp <= ahora || u.Iat > ahora || u.Iat <= 0 {
		return Claims{}, ErrToken
	}
	return u, nil
}
