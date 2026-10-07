package pedidos

import (
	"sort"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

// Tarifa es un rango de distancia vial con su costo de envío. El rango es
// semiabierto: incluye DesdeMetros y excluye HastaMetros. Así cada frontera
// pertenece a un solo rango (los 1.000 m exactos son del rango que empieza en
// 1.000, no del que termina en 1.000).
type Tarifa struct {
	DesdeMetros, HastaMetros int
	CostoCentavos            int64
}

var errSinCobertura = apierr.ErrRegla{
	Mensaje: "No hacemos envíos a esa distancia. Podés elegir retiro en el local.",
}

// CotizarEnvio devuelve el costo de enviar a distanciaMetros. La cobertura
// depende solo de la distancia vial máxima: si la excede, no se inventa un
// costo ni se asume envío gratis, se devuelve ErrRegla ofreciendo retiro. Los
// metros se comparan tal cual llegan, sin redondear antes de tarifar.
//
// Convención de la frontera máxima: la distancia igual al máximo SÍ está
// cubierta y paga la tarifa del último rango, aunque ese rango la excluya por
// ser semiabierto. El retiro no pasa por acá: no tiene cargo.
func CotizarEnvio(distanciaMetros int, maximoMetros int, tarifas []Tarifa) (int64, error) {
	if distanciaMetros < 0 {
		return 0, apierr.ErrValidacion{Campos: map[string]string{"distanciaMetros": "La distancia no puede ser negativa"}}
	}
	if distanciaMetros > maximoMetros {
		return 0, errSinCobertura
	}

	ordenadas := make([]Tarifa, len(tarifas))
	copy(ordenadas, tarifas)
	sort.Slice(ordenadas, func(i, j int) bool { return ordenadas[i].DesdeMetros < ordenadas[j].DesdeMetros })

	var elegida *Tarifa
	for i := range ordenadas {
		t := &ordenadas[i]
		if distanciaMetros >= t.DesdeMetros && distanciaMetros < t.HastaMetros {
			if elegida != nil {
				return 0, errTarifasInconsistentes
			}
			elegida = t
		}
	}

	// La frontera máxima exacta no cae en ningún rango semiabierto.
	if elegida == nil && distanciaMetros == maximoMetros {
		for i := range ordenadas {
			if ordenadas[i].HastaMetros == maximoMetros {
				elegida = &ordenadas[i]
			}
		}
	}

	if elegida == nil {
		// Hueco en la configuración: no se cotiza con un costo inventado.
		return 0, errTarifasInconsistentes
	}
	return elegida.CostoCentavos, nil
}

var errTarifasInconsistentes = apierr.ErrRegla{
	Mensaje: "No pudimos cotizar el envío para esa distancia. Podés elegir retiro en el local.",
}
