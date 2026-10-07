package pedidos

import (
	"context"

	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

// RepositorioConfiguracion lee la configuración vigente de la tienda.
type RepositorioConfiguracion struct{ db *gorm.DB }

func NuevoRepositorioConfiguracion(db *gorm.DB) *RepositorioConfiguracion {
	return &RepositorioConfiguracion{db: db}
}

// Actual devuelve la configuración más reciente, o ErrNoEncontrado si la
// tienda todavía no cargó ninguna.
func (r *RepositorioConfiguracion) Actual(ctx context.Context) (domain.ConfiguracionTienda, error) {
	var c domain.ConfiguracionTienda
	err := r.db.WithContext(ctx).Order("creado_en DESC").First(&c).Error
	return c, traducir(err)
}

// TarifasOrdenadas devuelve las tarifas de la configuración vigente, de la
// distancia más corta a la más larga.
func (r *RepositorioConfiguracion) TarifasOrdenadas(ctx context.Context) ([]domain.TarifaDistancia, error) {
	var tarifas []domain.TarifaDistancia
	err := r.db.WithContext(ctx).
		Where("configuracion_id = (SELECT id FROM configuracion_tienda ORDER BY creado_en DESC LIMIT 1)").
		Order("desde_metros").
		Find(&tarifas).Error
	return tarifas, err
}
