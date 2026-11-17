package pagos

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	usecases "github.com/Unknowns24/ecommerce-mdw/internal/usecases/inventario"
)

type pagosFalsos struct {
	estados []string
	err     error
}

func (f pagosFalsos) EstadosDePagosDelPedido(context.Context, string) ([]string, error) {
	return f.estados, f.err
}

// pedidoConReserva crea un pedido de Mercado Pago que vence en venceEn, con
// una unidad reservada de un lote.
func pedidoConReserva(t *testing.T, db *gorm.DB, venceEn time.Time) uuid.UUID {
	t.Helper()
	producto, variante, proveedor, lote, pedido := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	sentencias := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO producto (id) VALUES (?)`, []any{producto}},
		{`INSERT INTO variante (id, producto_id) VALUES (?, ?)`, []any{variante, producto}},
		{`INSERT INTO proveedor (id, nombre) VALUES (?, 'Proveedor de prueba')`, []any{proveedor}},
		{`INSERT INTO lote (id, variante_id, proveedor_id, unidades_ingresadas, costo_unitario_centavos, fecha_ingreso) VALUES (?, ?, ?, 1, 1000, now())`, []any{lote, variante, proveedor}},
		{`INSERT INTO pedido (id, estado_pedido, estado_pago, medio_pago, total_centavos, token_acceso, vence_en) VALUES (?, 'PENDIENTE_DE_PAGO', 'PENDIENTE', 'MERCADO_PAGO', 150000, ?, ?)`, []any{pedido, uuid.NewString(), venceEn}},
		{`INSERT INTO reserva_stock (id, pedido_id, lote_id, unidades, estado) VALUES (?, ?, ?, 1, 'ACTIVA')`, []any{uuid.New(), pedido, lote}},
	}
	for _, s := range sentencias {
		if err := db.Exec(s.sql, s.args...).Error; err != nil {
			t.Fatalf("sembrar pedido con reserva: %v", err)
		}
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM reserva_stock WHERE pedido_id = ?", pedido)
		db.Exec("DELETE FROM pedido WHERE id = ?", pedido)
		db.Exec("DELETE FROM lote WHERE id = ?", lote)
		db.Exec("DELETE FROM proveedor WHERE id = ?", proveedor)
		db.Exec("DELETE FROM variante WHERE id = ?", variante)
		db.Exec("DELETE FROM producto WHERE id = ?", producto)
	})
	return pedido
}

func estadoDe(t *testing.T, db *gorm.DB, pedido uuid.UUID) (estadoPedido, estadoReserva string) {
	t.Helper()
	if err := db.Raw(
		`SELECT p.estado_pedido, r.estado FROM pedido p JOIN reserva_stock r ON r.pedido_id = p.id WHERE p.id = ?`, pedido,
	).Row().Scan(&estadoPedido, &estadoReserva); err != nil {
		t.Fatalf("leer pedido y reserva: %v", err)
	}
	return estadoPedido, estadoReserva
}

func TestVencerPendientes(t *testing.T) {
	db := conectarParaTest(t)
	stock := usecases.NuevoServicioStock(db)
	ctx := context.Background()
	ahora := time.Now()

	casos := []struct {
		nombre          string
		venceEn         time.Time
		pagos           pagosFalsos
		pedidoEsperado  string
		reservaEsperada string
	}{
		{"vencido y sin pagos: se cancela y libera", ahora.Add(-time.Minute), pagosFalsos{}, "CANCELADO", "LIBERADA"},
		{"vencido con un pago rechazado: se cancela", ahora.Add(-time.Minute), pagosFalsos{estados: []string{"rejected"}}, "CANCELADO", "LIBERADA"},
		{"vencido con un pago en proceso: se espera", ahora.Add(-time.Minute), pagosFalsos{estados: []string{"in_process"}}, "PENDIENTE_DE_PAGO", "ACTIVA"},
		{"vencido y Mercado Pago no responde: no se asume que no pagó", ahora.Add(-time.Minute), pagosFalsos{err: errors.New("timeout")}, "PENDIENTE_DE_PAGO", "ACTIVA"},
		{"todavía vigente: no se toca", ahora.Add(time.Minute), pagosFalsos{}, "PENDIENTE_DE_PAGO", "ACTIVA"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			pedido := pedidoConReserva(t, db, c.venceEn)

			if _, err := vencerPendientes(ctx, db, stock, c.pagos, ahora); err != nil {
				t.Fatalf("vencerPendientes: %v", err)
			}

			estadoPedido, estadoReserva := estadoDe(t, db, pedido)
			if estadoPedido != c.pedidoEsperado || estadoReserva != c.reservaEsperada {
				t.Fatalf("pedido %s con reserva %s, esperado %s con %s", estadoPedido, estadoReserva, c.pedidoEsperado, c.reservaEsperada)
			}
		})
	}
}

// Vencer dos veces no hace nada la segunda: la reserva ya está liberada y el
// pedido ya no está pendiente.
func TestVencerPendientesEsIdempotente(t *testing.T) {
	db := conectarParaTest(t)
	stock := usecases.NuevoServicioStock(db)
	ctx := context.Background()
	ahora := time.Now()
	pedido := pedidoConReserva(t, db, ahora.Add(-time.Minute))

	if _, err := vencerPendientes(ctx, db, stock, pagosFalsos{}, ahora); err != nil {
		t.Fatal(err)
	}
	cancelado, err := vencerPedido(ctx, db, stock, pedido, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if cancelado {
		t.Fatal("la segunda pasada no debe volver a cancelar")
	}
}
