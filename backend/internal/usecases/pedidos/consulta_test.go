package pedidos

import "testing"

func TestNormalizarPaginacion(t *testing.T) {
	casos := []struct {
		nombre                             string
		pagina, tamano                     int
		wantPag, wantTam, wantLim, wantOff int
	}{
		{"primera página por defecto", 0, 0, 1, 20, 20, 0},
		{"página 3 de 10", 3, 10, 3, 10, 10, 20},
		{"página negativa es la 1", -5, 10, 1, 10, 10, 0},
		{"tamaño negativo usa el por defecto", 2, -1, 2, 20, 20, 20},
		{"tamaño justo en el máximo", 1, 100, 1, 100, 100, 0},
		{"un punto sobre el máximo se recorta", 1, 101, 1, 100, 100, 0},
		{"tamaño enorme se recorta", 2, 1000000, 2, 100, 100, 100},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			pag, tam, lim, off := NormalizarPaginacion(c.pagina, c.tamano)
			if pag != c.wantPag || tam != c.wantTam || lim != c.wantLim || off != c.wantOff {
				t.Errorf("(%d,%d) → pag=%d tam=%d limit=%d offset=%d; esperaba %d %d %d %d",
					c.pagina, c.tamano, pag, tam, lim, off, c.wantPag, c.wantTam, c.wantLim, c.wantOff)
			}
		})
	}
}
