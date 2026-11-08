package pedidos

import (
	"context"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// ConfiguracionVigente es la configuración de la tienda con sus tarifas de
// envío, tal como la necesita el checkout para anticipar el costo.
type ConfiguracionVigente struct {
	Configuracion domain.ConfiguracionTienda
	Tarifas       []domain.TarifaDistancia
}

// Configuracion atiende la lectura pública de la configuración de la tienda.
type Configuracion struct{ lector out.LectorConfiguracion }

func NuevaConfiguracion(lector out.LectorConfiguracion) *Configuracion {
	return &Configuracion{lector: lector}
}

// Vigente devuelve la configuración más reciente y sus tarifas ordenadas de la
// distancia más corta a la más larga. Si la tienda todavía no cargó ninguna,
// el error es ErrNoEncontrado (404) y no se consultan las tarifas.
func (c *Configuracion) Vigente(ctx context.Context) (ConfiguracionVigente, error) {
	cfg, err := c.lector.Actual(ctx)
	if err != nil {
		return ConfiguracionVigente{}, err
	}
	tarifas, err := c.lector.TarifasOrdenadas(ctx)
	if err != nil {
		return ConfiguracionVigente{}, err
	}
	return ConfiguracionVigente{Configuracion: cfg, Tarifas: tarifas}, nil
}
