package pedidos

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// AvisoReintegro es el texto que acompaña la cancelación de un pedido pagado:
// el sistema no devuelve plata, la devolución la hace una persona en Mercado
// Pago. Lo pide la spec (SPEC-H18).
const AvisoReintegro = "El reintegro debe realizarse manualmente en Mercado Pago. Este sistema no ejecuta devoluciones."

// Gestion son las transiciones de un pedido que hace la administración:
// despachar, entregar, cobrar en efectivo y cancelar. No son un CRUD: cada una
// es una operación del negocio con su regla (estados.go), y el cliente NO
// elige el estado de destino, solo pide que ocurra la operación.
//
// Todas hacen lo mismo, en este orden:
//  1. buscar el pedido (404 si no existe);
//  2. preguntarle a la función pura de estados.go si corresponde (409 con un
//     mensaje que dice qué hacer si no);
//  3. recién entonces, dentro de una transacción, cambiar el estado.
type Gestion struct {
	pedidos out.GestorPedidos
	stock   out.GestorStock
	tx      out.Transaccionador
	ahora   func() time.Time
}

func NuevaGestion(pedidos out.GestorPedidos, stock out.GestorStock, tx out.Transaccionador, ahora func() time.Time) *Gestion {
	return &Gestion{pedidos: pedidos, stock: stock, tx: tx, ahora: ahora}
}

// Listar devuelve una página de todos los pedidos, con filtros por estado y fecha.
func (g *Gestion) Listar(ctx context.Context, f out.FiltrosPedidos, pagina, tamano int) (PaginaPedidos, error) {
	pag, tam, limit, offset := NormalizarPaginacion(pagina, tamano)
	pedidos, total, err := g.pedidos.ListarAdmin(ctx, f, limit, offset)
	if err != nil {
		return PaginaPedidos{}, err
	}
	return PaginaPedidos{Pedidos: pedidos, Total: total, Pagina: pag, Tamano: tam}, nil
}

// Detalle devuelve el pedido completo, con comprador y pago.
func (g *Gestion) Detalle(ctx context.Context, id uuid.UUID) (domain.Pedido, error) {
	return g.pedidos.PorID(ctx, id)
}

// Despachar marca un envío confirmado y pagado como DESPACHADO.
func (g *Gestion) Despachar(ctx context.Context, id uuid.UUID) (domain.Pedido, error) {
	return g.transicionar(ctx, id, PuedeDespachar, func(tx *gorm.DB, p domain.Pedido) error {
		return g.pedidos.CambiarEstado(ctx, tx, p.ID, p.EstadoPedido, domain.PedidoDespachado)
	})
}

// Entregar completa el pedido: un envío desde DESPACHADO, un retiro desde
// CONFIRMADO y pagado.
func (g *Gestion) Entregar(ctx context.Context, id uuid.UUID) (domain.Pedido, error) {
	return g.transicionar(ctx, id, PuedeEntregar, func(tx *gorm.DB, p domain.Pedido) error {
		return g.pedidos.CambiarEstado(ctx, tx, p.ID, p.EstadoPedido, domain.PedidoCompletado)
	})
}

// CobrarEnEfectivo registra, en una sola acción, el cobro y la entrega de un
// retiro en efectivo: marca el pago como cobrado, convierte la reserva de stock
// en venta y completa el pedido. O las tres cosas o ninguna.
func (g *Gestion) CobrarEnEfectivo(ctx context.Context, id uuid.UUID) (domain.Pedido, error) {
	return g.transicionar(ctx, id, PuedeCompletarEnEfectivo, func(tx *gorm.DB, p domain.Pedido) error {
		if err := g.pedidos.MarcarPagoAprobado(ctx, tx, p.ID); err != nil {
			return err
		}
		if err := g.stock.Confirmar(ctx, tx, p.ID); err != nil {
			return err
		}
		return g.pedidos.CambiarEstado(ctx, tx, p.ID, p.EstadoPedido, domain.PedidoCompletado)
	})
}

// ResultadoCancelacion es el pedido cancelado y, si estaba pagado, el aviso de
// que el reintegro se hace fuera del sistema.
type ResultadoCancelacion struct {
	Pedido domain.Pedido
	Aviso  string
}

// Cancelar cancela un pedido PENDIENTE_DE_PAGO o CONFIRMADO (nunca uno
// despachado o completado), libera el stock y guarda quién, cuándo y por qué.
// El responsable sale de la sesión, nunca del body.
//
// Es idempotente: cancelar dos veces no devuelve el stock dos veces. Lo
// garantizan dos cosas: la segunda vez el estado ya es CANCELADO y
// PuedeCancelar la rechaza (409), y Liberar del inventario es idempotente.
// Además el cambio de estado es lo PRIMERO que se hace dentro de la
// transacción: si dos administradores cancelan a la vez, solo uno logra el
// UPDATE (CambiarEstado exige el estado de origen) y el otro recibe 409 antes
// de tocar el stock.
func (g *Gestion) Cancelar(ctx context.Context, id, responsableID uuid.UUID, motivo string) (ResultadoCancelacion, error) {
	var estabaPagado bool
	pedido, err := g.transicionar(ctx, id, PuedeCancelar, func(tx *gorm.DB, p domain.Pedido) error {
		estabaPagado = p.EstadoPago == domain.PagoAprobado
		if err := g.pedidos.CambiarEstado(ctx, tx, p.ID, p.EstadoPedido, domain.PedidoCancelado); err != nil {
			return err
		}
		if err := g.pedidos.RegistrarCancelacion(ctx, tx, p.ID, responsableID, motivo, g.ahora()); err != nil {
			return err
		}
		return g.stock.Liberar(ctx, tx, p.ID)
	})
	if err != nil {
		return ResultadoCancelacion{}, err
	}

	res := ResultadoCancelacion{Pedido: pedido}
	if estabaPagado {
		res.Aviso = AvisoReintegro
	}
	return res, nil
}

// transicionar es el esqueleto común: busca, pregunta a la regla pura, hace el
// trabajo en una transacción y devuelve el pedido actualizado.
func (g *Gestion) transicionar(
	ctx context.Context,
	id uuid.UUID,
	regla func(domain.Pedido) error,
	trabajo func(tx *gorm.DB, p domain.Pedido) error,
) (domain.Pedido, error) {
	p, err := g.pedidos.PorID(ctx, id)
	if err != nil {
		return domain.Pedido{}, err // 404 si no existe
	}
	if err := regla(p); err != nil {
		return domain.Pedido{}, err // 409 con el mensaje de qué hacer
	}
	if err := g.tx.EnTransaccion(ctx, func(tx *gorm.DB) error { return trabajo(tx, p) }); err != nil {
		return domain.Pedido{}, err
	}
	return g.pedidos.PorID(ctx, id)
}
