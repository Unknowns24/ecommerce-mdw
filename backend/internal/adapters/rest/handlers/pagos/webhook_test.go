package pagos

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql"
	database "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/migrations"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/migrations/schema"
	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/pagos"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/external/mp"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	usecases "github.com/Unknowns24/ecommerce-mdw/internal/usecases/inventario"
)

// conectarParaTest levanta la conexión real de inventario+pagos contra
// DATABASE_DSN y crea, sólo para este test, el recorte mínimo de las tablas
// de identidad/catálogo/pedidos del que las FKs de inventario dependen.
// Ningún módulo ajeno se edita: esto es una fixture de test, no una
// migración del sistema.
func conectarParaTest(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Skip("DATABASE_DSN no está configurada: se omite el test de integración del webhook")
	}

	db, err := postgresql.Abrir(dsn)
	if err != nil {
		t.Fatalf("Abrir() error = %v", err)
	}

	fixture := []string{
		`CREATE TABLE IF NOT EXISTS usuario (id uuid PRIMARY KEY)`,
		`CREATE TABLE IF NOT EXISTS producto (id uuid PRIMARY KEY, metodologia_rotacion varchar(10) NOT NULL DEFAULT 'FIFO')`,
		`CREATE TABLE IF NOT EXISTS variante (id uuid PRIMARY KEY, producto_id uuid NOT NULL REFERENCES producto(id))`,
		`CREATE TABLE IF NOT EXISTS pedido (
			id uuid PRIMARY KEY,
			usuario_id uuid,
			estado_pedido varchar(30) NOT NULL,
			estado_pago varchar(30) NOT NULL,
			medio_pago varchar(20) NOT NULL,
			total_centavos bigint NOT NULL,
			actualizado_en timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS detalle_pedido (
			id uuid PRIMARY KEY,
			pedido_id uuid NOT NULL REFERENCES pedido(id),
			variante_id uuid NOT NULL,
			codigo varchar(50) NOT NULL,
			nombre varchar(255) NOT NULL,
			unidades integer NOT NULL,
			precio_unitario_centavos bigint NOT NULL
		)`,
	}
	for _, sentencia := range fixture {
		if err := db.Exec(sentencia).Error; err != nil {
			t.Fatalf("crear fixture: %v", err)
		}
	}

	migrador := database.NewMigrator(db, schema.MigracionesInventario())
	if err := migrador.Migrate(); err != nil {
		t.Fatalf("migrar inventario+pagos: %v", err)
	}

	return db
}

func crearPedidoDePrueba(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	pedidoID := uuid.New()
	err := db.Exec(
		`INSERT INTO pedido (id, estado_pedido, estado_pago, medio_pago, total_centavos) VALUES (?, 'PENDIENTE_DE_PAGO', 'PENDIENTE', 'MERCADO_PAGO', 150000)`,
		pedidoID,
	).Error
	if err != nil {
		t.Fatalf("crear pedido de prueba: %v", err)
	}
	return pedidoID
}

// TestProcesarNotificacionEsIdempotente es el caso que importa: la misma
// notificación de Mercado Pago procesada dos veces deja el sistema igual
// que una sola vez — un solo intento_pago APROBADO, un solo UPDATE del
// pedido, sin error en la segunda pasada.
func TestProcesarNotificacionEsIdempotente(t *testing.T) {
	db := conectarParaTest(t)
	pedidoID := crearPedidoDePrueba(t, db)

	stock := usecases.NuevoServicioStock(db)
	intentos := repo.NuevoRepositorioIntentos(db)
	ctx := context.Background()

	pago := mp.Pago{
		ID:                "pago-idempotencia-" + uuid.NewString(),
		Estado:            "approved",
		MontoCentavos:     150000,
		ReferenciaExterna: pedidoID.String(),
	}

	if err := procesarNotificacion(ctx, db, stock, intentos, pedidoID, pago); err != nil {
		t.Fatalf("primera notificación: error = %v", err)
	}
	if err := procesarNotificacion(ctx, db, stock, intentos, pedidoID, pago); err != nil {
		t.Fatalf("segunda notificación (repetida): error = %v", err)
	}

	var totalIntentos int64
	if err := db.Model(&domain.IntentoPago{}).
		Where("proveedor = ? AND referencia_externa = ?", domain.ProveedorMercadoPago, pago.ID).
		Count(&totalIntentos).Error; err != nil {
		t.Fatalf("contar intentos: %v", err)
	}
	if totalIntentos != 1 {
		t.Fatalf("intentos de pago registrados = %d, want 1 (la notificación repetida no debe duplicar)", totalIntentos)
	}

	intento, err := intentos.UltimoPorPedido(ctx, pedidoID)
	if err != nil {
		t.Fatalf("UltimoPorPedido() error = %v", err)
	}
	if intento.Estado != domain.IntentoAprobado {
		t.Fatalf("estado del intento = %q, want APROBADO", intento.Estado)
	}

	var estadoPedido, estadoPago string
	if err := db.Raw("SELECT estado_pedido, estado_pago FROM pedido WHERE id = ?", pedidoID).
		Row().Scan(&estadoPedido, &estadoPago); err != nil {
		t.Fatalf("leer pedido: %v", err)
	}
	if estadoPedido != "CONFIRMADO" || estadoPago != "APROBADO" {
		t.Fatalf("pedido = (%s, %s), want (CONFIRMADO, APROBADO)", estadoPedido, estadoPago)
	}

	t.Cleanup(func() {
		db.Exec("DELETE FROM intento_pago WHERE pedido_id = ?", pedidoID)
		db.Exec("DELETE FROM detalle_pedido WHERE pedido_id = ?", pedidoID)
		db.Exec("DELETE FROM pedido WHERE id = ?", pedidoID)
	})
}
