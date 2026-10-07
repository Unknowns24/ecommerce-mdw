// Command seed loads demo data for the inventory module: one provider and
// lots over three well-known variant ids — one with plenty of stock, one
// with a single unit, and one with zero — so the stock business rules are
// demonstrable on Tuesday. It's idempotent: running it many times never
// duplicates rows.
package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

// Estos ids tienen que coincidir con los que siembra backend/cmd/seed-catalogo
// (pendiente, lo siembra Genaro). Si todavía no existen, este comando avisa
// y sigue con lo que sí puede sembrar, en vez de romperse.
var (
	varianteConStock  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	varianteUnaUnidad = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	varianteSinStock  = uuid.MustParse("33333333-3333-3333-3333-333333333333")
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

	ctx := context.Background()

	proveedor, err := sembrarProveedor(ctx, db)
	if err != nil {
		log.Fatalf("sembrar proveedor: %v", err)
	}

	escenarios := []struct {
		varianteID uuid.UUID
		unidades   int
		motivo     string
	}{
		{varianteConStock, 50, "variante con stock de sobra"},
		{varianteUnaUnidad, 1, "variante con una sola unidad"},
		{varianteSinStock, 0, "variante sin stock"},
	}

	for _, e := range escenarios {
		if e.unidades == 0 {
			log.Printf("omitido a propósito: %s no recibe lote (debe quedar en cero)", e.motivo)
			continue
		}
		existe, err := varianteExisteYa(ctx, db, e.varianteID)
		if err != nil {
			log.Fatalf("verificar variante %s: %v", e.varianteID, err)
		}
		if !existe {
			log.Printf("AVISO: la variante %s todavía no existe (falta el seed de catálogo) — se omite por ahora", e.varianteID)
			continue
		}
		if err := sembrarLote(ctx, db, e.varianteID, proveedor.ID, e.unidades, e.motivo); err != nil {
			log.Fatalf("sembrar lote para %s: %v", e.varianteID, err)
		}
	}

	log.Println("seed de inventario aplicado correctamente")
}

func sembrarProveedor(ctx context.Context, db *gorm.DB) (domain.Proveedor, error) {
	var proveedor domain.Proveedor
	err := db.WithContext(ctx).Where("nombre = ?", "Distribuidora Central").FirstOrCreate(&proveedor, domain.Proveedor{
		ID:     uuid.New(),
		Nombre: "Distribuidora Central",
	}).Error
	return proveedor, err
}

func varianteExisteYa(ctx context.Context, db *gorm.DB, id uuid.UUID) (bool, error) {
	var existe bool
	err := db.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM variante WHERE id = ?)", id).Scan(&existe).Error
	return existe, err
}

func sembrarLote(ctx context.Context, db *gorm.DB, varianteID, proveedorID uuid.UUID, unidades int, motivo string) error {
	var cuenta int64
	if err := db.WithContext(ctx).Model(&domain.Lote{}).Where("variante_id = ?", varianteID).Count(&cuenta).Error; err != nil {
		return err
	}
	if cuenta > 0 {
		log.Printf("ya existe un lote para %s (%s), no se duplica", varianteID, motivo)
		return nil
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lote := domain.Lote{
			ID:                    uuid.New(),
			VarianteID:            varianteID,
			ProveedorID:           proveedorID,
			UnidadesIngresadas:    unidades,
			CostoUnitarioCentavos: 100000,
			FechaIngreso:          time.Now(),
		}
		if err := tx.Create(&lote).Error; err != nil {
			return err
		}
		movimiento := domain.MovimientoStock{
			ID:       uuid.New(),
			LoteID:   lote.ID,
			Tipo:     domain.MovimientoIngreso,
			Unidades: unidades,
			Motivo:   "Seed: " + motivo,
		}
		return tx.Create(&movimiento).Error
	})
}
