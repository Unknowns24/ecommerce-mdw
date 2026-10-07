// Package apierr is the single error contract of the HTTP API.
//
// Any layer — usecase, repository, adapter — may return the sentinel errors
// or the typed errors declared here. Only the REST handler, through
// Responder, translates an error into an HTTP status and a response body.
// No other package decides status codes.
package apierr

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// Cuerpo is the single response body shape for every error the API returns.
type Cuerpo struct {
	Error    string `json:"error"`
	Detalles any    `json:"detalles,omitempty"`
}

// Sentinel domain errors. Any layer can return them; only the handler,
// through Responder, translates them into an HTTP status.
var (
	ErrNoAutenticado = errors.New("no autenticado") // → 401
	ErrNoAutorizado  = errors.New("no autorizado")  // → 403
	ErrNoEncontrado  = errors.New("no encontrado")  // → 404
)

// ErrValidacion reports that the request body itself is malformed: the shape
// is wrong independent of any system state. → 400
type ErrValidacion struct {
	Campos map[string]string
}

func (e ErrValidacion) Error() string {
	return "datos inválidos"
}

// ErrRegla reports a business rule violation that required checking state.
// Datos carries whatever structured payload the acceptance criteria asks to
// enumerate (e.g. variante → unidades disponibles). → 409
type ErrRegla struct {
	Mensaje string
	Datos   any
}

func (e ErrRegla) Error() string {
	return e.Mensaje
}

// JSON writes cuerpo as the JSON response body with the given status code.
func JSON(w http.ResponseWriter, status int, cuerpo any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if cuerpo == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(cuerpo); err != nil {
		log.Printf("apierr: no se pudo escribir la respuesta: %v", err)
	}
}

// Responder is the only place in the system that translates an error into an
// HTTP status and a response body. endpoint identifies where the error was
// produced, for the server log of unexpected failures.
func Responder(w http.ResponseWriter, endpoint string, err error) {
	switch {
	case errors.Is(err, ErrNoAutenticado):
		JSON(w, http.StatusUnauthorized, Cuerpo{Error: "No autenticado"})
	case errors.Is(err, ErrNoAutorizado):
		JSON(w, http.StatusForbidden, Cuerpo{Error: "No podés realizar esta operación"})
	case errors.Is(err, ErrNoEncontrado):
		JSON(w, http.StatusNotFound, Cuerpo{Error: "No encontrado"})
	default:
		var errVal ErrValidacion
		if errors.As(err, &errVal) {
			JSON(w, http.StatusBadRequest, Cuerpo{Error: "Datos inválidos", Detalles: map[string]any{"campos": errVal.Campos}})
			return
		}
		var errRegla ErrRegla
		if errors.As(err, &errRegla) {
			JSON(w, http.StatusConflict, Cuerpo{Error: errRegla.Mensaje, Detalles: errRegla.Datos})
			return
		}
		log.Printf("%s: %v", endpoint, err)
		JSON(w, http.StatusInternalServerError, Cuerpo{Error: "Error interno"})
	}
}
