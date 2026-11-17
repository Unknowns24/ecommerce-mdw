package identidad

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

func TestCuentaIncluyeRolesYPermisosSinCredencial(t *testing.T) {
	u := domain.Usuario{Nombre: "Ana", Correo: "ana@example.com", CredencialHash: "hash-secreto"}

	bruto, err := json.Marshal(NuevaCuenta(u, []string{"dueno"}, []string{"pedidos.gestionar"}))
	if err != nil {
		t.Fatal(err)
	}
	var cuerpo map[string]any
	if err := json.Unmarshal(bruto, &cuerpo); err != nil {
		t.Fatal(err)
	}
	if cuerpo["correo"] != "ana@example.com" {
		t.Errorf("faltan los datos del usuario: %s", bruto)
	}
	if roles, _ := cuerpo["roles"].([]any); len(roles) != 1 || roles[0] != "dueno" {
		t.Errorf("roles = %v, esperado [dueno]", cuerpo["roles"])
	}
	if permisos, _ := cuerpo["permisos"].([]any); len(permisos) != 1 || permisos[0] != "pedidos.gestionar" {
		t.Errorf("permisos = %v, esperado [pedidos.gestionar]", cuerpo["permisos"])
	}
	if strings.Contains(string(bruto), "hash-secreto") {
		t.Errorf("la respuesta expone la credencial: %s", bruto)
	}
}

// Un cliente sin roles administrativos recibe listas vacías, no null: la
// interfaz recorre la lista sin preguntar si existe.
func TestCuentaSinPermisosDevuelveListasVacias(t *testing.T) {
	bruto, _ := json.Marshal(NuevaCuenta(domain.Usuario{}, nil, nil))
	if !strings.Contains(string(bruto), `"roles":[]`) || !strings.Contains(string(bruto), `"permisos":[]`) {
		t.Errorf("esperado listas vacías: %s", bruto)
	}
}
