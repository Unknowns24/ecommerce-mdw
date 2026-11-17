package catalogo

import "testing"

func TestLaVentaExigeVarianteProductoYMarcaActivos(t *testing.T) {
	for _, c := range []struct {
		variante, producto, marca string
		esperado                  bool
	}{
		{"ACTIVA", "ACTIVO", "ACTIVA", true},
		{"INACTIVA", "ACTIVO", "ACTIVA", false},
		{"ACTIVA", "INACTIVO", "ACTIVA", false},
		{"ACTIVA", "ACTIVO", "INACTIVA", false},
	} {
		if got := puedeVenderse(c.variante, c.producto, c.marca); got != c.esperado {
			t.Fatalf("venta con estados %+v: obtuvo %v", c, got)
		}
	}
}
