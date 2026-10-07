// Command seed-pedidos carga la configuración inicial de la tienda: si el
// efectivo está habilitado, la distancia máxima de envío y las tarifas por
// rango de distancia. Sin esto, todo envío y todo pago en efectivo darían 409
// ("el envío no está disponible"), porque el checkout no inventa una
// configuración.
//
// Es idempotente: si ya hay una configuración, no hace nada. Se puede correr
// las veces que haga falta:
//
//	go run ./cmd/seed-pedidos
package main

import (
	"log"

	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuración inválida: %v", err)
	}
	db, err := postgresql.Abrir(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("no se pudo conectar a la base de datos: %v", err)
	}

	creada, err := sembrarConfiguracion(db)
	if err != nil {
		log.Fatalf("no se pudo sembrar la configuración de la tienda: %v", err)
	}
	if creada {
		log.Println("configuración de la tienda creada (efectivo habilitado, cobertura hasta 5 km, 3 tarifas)")
	} else {
		log.Println("la configuración de la tienda ya existía: no se cambió nada")
	}
}

// sembrarConfiguracion crea la configuración y sus tarifas en una transacción
// si todavía no hay ninguna. Devuelve true si la creó.
func sembrarConfiguracion(db *gorm.DB) (bool, error) {
	var cantidad int64
	if err := db.Model(&domain.ConfiguracionTienda{}).Count(&cantidad).Error; err != nil {
		return false, err
	}
	if cantidad > 0 {
		return false, nil
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		c := domain.ConfiguracionTienda{EfectivoHabilitado: true, DistanciaMaximaMetros: 5000}
		if err := tx.Create(&c).Error; err != nil {
			return err
		}
		// Rangos semiabiertos [desde, hasta): cada frontera pertenece a un solo
		// rango. Importes en centavos: $500, $900 y $1.500.
		tarifas := []domain.TarifaDistancia{
			{ConfiguracionID: c.ID, DesdeMetros: 0, HastaMetros: 1000, CostoCentavos: 50000},
			{ConfiguracionID: c.ID, DesdeMetros: 1000, HastaMetros: 3000, CostoCentavos: 90000},
			{ConfiguracionID: c.ID, DesdeMetros: 3000, HastaMetros: 5000, CostoCentavos: 150000},
		}
		return tx.Create(&tarifas).Error
	})
	return err == nil, err
}
