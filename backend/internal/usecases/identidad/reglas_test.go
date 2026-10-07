package identidad

import (
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/google/uuid"
	"testing"
)

func TestReglasDuenoInicial(t *testing.T) {
	a := domain.Usuario{ID: uuid.New(), Estado: "ACTIVO"}
	inicial := domain.Usuario{ID: uuid.New(), Estado: "ACTIVO", EsDuenoInicial: true}
	otro := domain.Usuario{ID: uuid.New(), Estado: "ACTIVO"}
	if PuedeCambiarEstado(a, inicial, "INACTIVO") == nil {
		t.Error("permitió desactivar dueño inicial")
	}
	if PuedeCambiarEstado(a, a, "INACTIVO") == nil {
		t.Error("permitió desactivarse")
	}
	if PuedeAsignarRoles(a, inicial, []string{"cliente"}) == nil {
		t.Error("permitió quitar rol dueño")
	}
	if err := PuedeCambiarEstado(a, otro, "INACTIVO"); err != nil {
		t.Errorf("otro administrador: %v", err)
	}
}
