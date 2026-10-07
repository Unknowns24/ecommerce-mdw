// Command seed-catalogo loads repeatable demo data before the inventory seed.
package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/config"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

var (
	conStock  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	unaUnidad = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	sinStock  = uuid.MustParse("33333333-3333-3333-3333-333333333333")
)

func id(clave string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("bc-importados/"+clave))
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := postgresql.Abrir(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	if err := sembrar(db); err != nil {
		log.Fatal(err)
	}
	log.Println("seed de catálogo aplicado")
}

func sembrar(db *gorm.DB) error {
	now := time.Now().UTC()
	marcas := []string{"Victoria's Secret", "Bath & Body Works", "JBL"}
	for _, nombre := range marcas {
		marca := domain.Marca{ID: id("marca/" + nombre), Nombre: nombre, Estado: "ACTIVA", CreadoEn: now}
		if err := db.Where("id = ?", marca.ID).Attrs(marca).FirstOrCreate(&domain.Marca{}).Error; err != nil {
			return err
		}
	}
	categorias := []struct{ nombre, padre string }{{"Perfumes", ""}, {"Cremas", ""}, {"Body splash", "Perfumes"}, {"Electrónicos", ""}}
	for _, c := range categorias {
		var padreID *uuid.UUID
		if c.padre != "" {
			p := id("categoria/" + c.padre)
			padreID = &p
		}
		categoria := domain.Categoria{ID: id("categoria/" + c.nombre), Nombre: c.nombre, PadreID: padreID, Estado: "ACTIVA", CreadoEn: now}
		if err := db.Where("id = ?", categoria.ID).Attrs(categoria).FirstOrCreate(&domain.Categoria{}).Error; err != nil {
			return err
		}
	}
	productos := []struct{ clave, marca, nombre, categoria, estado string }{
		{"love-spell", "Victoria's Secret", "Love Spell", "Cremas", "ACTIVO"},
		{"pure-seduction", "Victoria's Secret", "Pure Seduction", "Body splash", "ACTIVO"},
		{"bare-vanilla", "Victoria's Secret", "Bare Vanilla", "Cremas", "ACTIVO"},
		{"bombshell", "Victoria's Secret", "Bombshell", "Perfumes", "ACTIVO"},
		{"japanese-cherry-blossom", "Bath & Body Works", "Japanese Cherry Blossom", "Cremas", "ACTIVO"},
		{"gingham", "Bath & Body Works", "Gingham", "Body splash", "ACTIVO"},
		{"tune-510bt", "JBL", "Tune 510BT", "Electrónicos", "ACTIVO"},
		{"go-4", "JBL", "Go 4", "Electrónicos", "INACTIVO"},
	}
	for _, p := range productos {
		producto := domain.Producto{ID: id("producto/" + p.clave), MarcaID: id("marca/" + p.marca), Nombre: p.nombre, Estado: p.estado, MetodologiaRotacion: "FIFO", CreadoEn: now, ActualizadoEn: now}
		if err := db.Where("id = ?", producto.ID).Attrs(producto).FirstOrCreate(&domain.Producto{}).Error; err != nil {
			return err
		}
		pc := domain.ProductoCategoria{ProductoID: producto.ID, CategoriaID: id("categoria/" + p.categoria)}
		if err := db.Where("producto_id = ? AND categoria_id = ?", pc.ProductoID, pc.CategoriaID).Attrs(pc).FirstOrCreate(&domain.ProductoCategoria{}).Error; err != nil {
			return err
		}
	}
	variantes := []struct {
		clave, producto, codigo, nombre, estado string
		identificador                           uuid.UUID
		precio                                  int64
	}{
		{"love-spell-236", "love-spell", "VS-LOVE-236", "Love Spell crema 236 ml", "ACTIVA", conStock, 1590000},
		{"pure-seduction-250", "pure-seduction", "VS-PURE-250", "Pure Seduction splash 250 ml", "ACTIVA", unaUnidad, 1690000},
		{"bare-vanilla-236", "bare-vanilla", "VS-BARE-236", "Bare Vanilla crema 236 ml", "ACTIVA", sinStock, 1590000},
		{"love-spell-100", "love-spell", "VS-LOVE-100", "Love Spell crema 100 ml", "ACTIVA", id("variante/love-spell-100"), 890000},
		{"pure-seduction-100", "pure-seduction", "VS-PURE-100", "Pure Seduction splash 100 ml", "INACTIVA", id("variante/pure-seduction-100"), 990000},
		{"bare-vanilla-100", "bare-vanilla", "VS-BARE-100", "Bare Vanilla crema 100 ml", "ACTIVA", id("variante/bare-vanilla-100"), 890000},
		{"bombshell-50", "bombshell", "VS-BOMB-50", "Bombshell eau de parfum 50 ml", "ACTIVA", id("variante/bombshell-50"), 6900000},
		{"japanese-cherry-236", "japanese-cherry-blossom", "BBW-JCB-236", "Japanese Cherry Blossom crema 236 ml", "ACTIVA", id("variante/japanese-cherry-236"), 1790000},
		{"gingham-236", "gingham", "BBW-GING-236", "Gingham splash 236 ml", "ACTIVA", id("variante/gingham-236"), 1790000},
		{"tune-510bt-negro", "tune-510bt", "JBL-T510BT-BK", "JBL Tune 510BT negro", "ACTIVA", id("variante/tune-510bt-negro"), 8900000},
		{"tune-510bt-azul", "tune-510bt", "JBL-T510BT-BL", "JBL Tune 510BT azul", "ACTIVA", id("variante/tune-510bt-azul"), 8900000},
		{"go-4-negro", "go-4", "JBL-GO4-BK", "JBL Go 4 negro", "ACTIVA", id("variante/go-4-negro"), 6900000},
	}
	for _, v := range variantes {
		variante := domain.Variante{ID: v.identificador, ProductoID: id("producto/" + v.producto), Codigo: v.codigo, Nombre: v.nombre, PrecioMinoristaCentavos: v.precio, PrecioMayoristaCentavos: v.precio * 8 / 10, Estado: v.estado, CreadoEn: now, ActualizadoEn: now}
		if err := db.Where("id = ?", variante.ID).Attrs(variante).FirstOrCreate(&domain.Variante{}).Error; err != nil {
			return err
		}
	}
	return nil
}
