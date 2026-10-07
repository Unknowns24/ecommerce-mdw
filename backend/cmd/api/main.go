// Command api is the composition root of the backend: it loads the
// configuration, opens the database connection, mounts every module's
// routes exactly once, and starts the HTTP server.
package main

import (
	"log"
	"net/http"
	"os"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Unknowns24/ecommerce-mdw/config"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql"
	database "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/migrations"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/handlers/catalogo"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/handlers/identidad"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/handlers/inventario"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/handlers/pagos"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/handlers/pedidos"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/router"
)

// version se completa en build con -ldflags "-X main.version=<git-sha-corto>".
var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuración inválida: %v", err)
	}

	db, err := postgresql.Abrir(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("no se pudo conectar a la base de datos: %v", err)
	}

	// Las migraciones no corren al arrancar el servidor: es un subcomando
	// explícito (./api migrate). Correrlas en el arranque haría que un
	// deploy con el server ya corriendo migre la base de la nube sin
	// querer. Ver ADR-002 y ADR-005.
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		migrador := database.NewMigrator(db, database.Todas())
		if err := migrador.Migrate(); err != nil {
			log.Fatalf("migración falló: %v", err)
		}
		log.Println("migraciones aplicadas correctamente")
		return
	}

	mw := middleware.Nuevo(cfg.AppSecretKey)
	r := router.NewRouter()
	r.Use(chimw.RequestID, chimw.RealIP, chimw.Recoverer, chimw.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Get("/api/salud", salud)

	identidad.Montar(r, db, cfg, mw)
	catalogo.Montar(r, db, cfg, mw)
	inventario.Montar(r, db, cfg, mw)
	pagos.Montar(r, db, cfg, mw)
	pedidos.Montar(r, db, cfg, mw)

	log.Printf("escuchando en :%s (version=%s)", cfg.AppPort, version)
	if err := http.ListenAndServe(":"+cfg.AppPort, r); err != nil {
		log.Fatalf("el servidor se detuvo: %v", err)
	}
}

func salud(w http.ResponseWriter, r *http.Request) {
	apierr.JSON(w, http.StatusOK, map[string]string{"estado": "ok", "version": version})
}
