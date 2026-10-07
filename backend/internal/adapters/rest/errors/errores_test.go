package apierr

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestResponderTraduceCadaErrorAlStatusCorrecto(t *testing.T) {
	casos := []struct {
		nombre        string
		err           error
		statusPerado  int
		errorEsperado string
	}{
		{"no autenticado", ErrNoAutenticado, 401, "No autenticado"},
		{"no autorizado", ErrNoAutorizado, 403, "No podés realizar esta operación"},
		{"no encontrado", ErrNoEncontrado, 404, "No encontrado"},
		{"validacion", ErrValidacion{Campos: map[string]string{"correo": "es requerido"}}, 400, "Datos inválidos"},
		{"regla de negocio", ErrRegla{Mensaje: "No hay stock suficiente", Datos: map[string]int{"v1": 2}}, 409, "No hay stock suficiente"},
		{"error desconocido", errors.New("boom"), 500, "Error interno"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := httptest.NewRecorder()
			Responder(w, "POST /test", c.err)

			if w.Code != c.statusPerado {
				t.Fatalf("status = %d, want %d", w.Code, c.statusPerado)
			}

			var cuerpo Cuerpo
			if err := json.NewDecoder(w.Body).Decode(&cuerpo); err != nil {
				t.Fatalf("no se pudo decodificar el cuerpo: %v", err)
			}
			if cuerpo.Error != c.errorEsperado {
				t.Fatalf("error = %q, want %q", cuerpo.Error, c.errorEsperado)
			}
		})
	}
}

func TestResponderErrorDesconocidoNoRevelaDetalles(t *testing.T) {
	w := httptest.NewRecorder()
	Responder(w, "POST /test", errors.New("detalle sensible de implementación"))

	var cuerpo Cuerpo
	if err := json.NewDecoder(w.Body).Decode(&cuerpo); err != nil {
		t.Fatalf("no se pudo decodificar el cuerpo: %v", err)
	}
	if cuerpo.Detalles != nil {
		t.Fatalf("Detalles = %v, want nil para un error 500", cuerpo.Detalles)
	}
}
