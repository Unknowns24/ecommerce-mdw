package pedidos

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// gestorFalso guarda los pedidos en memoria y respeta la regla clave del
// repositorio real: CambiarEstado solo actúa si el estado actual coincide con
// el de origen (si no, 409).
type gestorFalso struct {
	pedidos map[uuid.UUID]*domain.Pedido
	// alLeer se ejecuta justo después de leer un pedido: sirve para simular
	// que OTRA persona lo cambió entre la lectura y la escritura.
	alLeer                    func(*domain.Pedido)
	ultimoLimit, ultimoOffset int
	ultimoFiltro              out.FiltrosPedidos
}

func (g *gestorFalso) PorID(_ context.Context, id uuid.UUID) (domain.Pedido, error) {
	p, ok := g.pedidos[id]
	if !ok {
		return domain.Pedido{}, apierr.ErrNoEncontrado
	}
	copia := *p
	if g.alLeer != nil {
		alLeer := g.alLeer
		g.alLeer = nil
		alLeer(p)
	}
	return copia, nil
}

func (g *gestorFalso) ListarAdmin(_ context.Context, f out.FiltrosPedidos, limit, offset int) ([]domain.Pedido, int64, error) {
	g.ultimoFiltro, g.ultimoLimit, g.ultimoOffset = f, limit, offset
	var res []domain.Pedido
	for _, p := range g.pedidos {
		if f.Estado == nil || p.EstadoPedido == *f.Estado {
			res = append(res, *p)
		}
	}
	return res, int64(len(res)), nil
}

func (g *gestorFalso) CambiarEstado(_ context.Context, _ *gorm.DB, id uuid.UUID, desde, nuevo domain.EstadoPedido) error {
	p := g.pedidos[id]
	if p.EstadoPedido != desde {
		return apierr.ErrRegla{Mensaje: "El pedido cambió de estado mientras lo procesabas. Actualizá la pantalla e intentá de nuevo."}
	}
	p.EstadoPedido = nuevo
	return nil
}

func (g *gestorFalso) MarcarPagoAprobado(_ context.Context, _ *gorm.DB, id uuid.UUID) error {
	p := g.pedidos[id]
	if p.EstadoPago != domain.PagoPendiente {
		return apierr.ErrRegla{Mensaje: "El cobro de este pedido ya fue registrado."}
	}
	p.EstadoPago = domain.PagoAprobado
	return nil
}

func (g *gestorFalso) RegistrarCancelacion(_ context.Context, _ *gorm.DB, id, responsableID uuid.UUID, motivo string, ahora time.Time) error {
	p := g.pedidos[id]
	p.CanceladoPor, p.CanceladoEn, p.MotivoCancelacion = &responsableID, &ahora, &motivo
	return nil
}

type stockGestionFalso struct{ confirmaciones, liberaciones int }

func (s *stockGestionFalso) Confirmar(context.Context, *gorm.DB, uuid.UUID) error {
	s.confirmaciones++
	return nil
}
func (s *stockGestionFalso) Liberar(context.Context, *gorm.DB, uuid.UUID) error {
	s.liberaciones++
	return nil
}

type escenarioGestion struct {
	gestion *Gestion
	gestor  *gestorFalso
	stock   *stockGestionFalso
	tx      *txFalsa
	ahora   time.Time
}

func nuevoEscenarioGestion(pedidos ...*domain.Pedido) *escenarioGestion {
	e := &escenarioGestion{
		gestor: &gestorFalso{pedidos: map[uuid.UUID]*domain.Pedido{}},
		stock:  &stockGestionFalso{},
		tx:     &txFalsa{},
		ahora:  time.Date(2026, 10, 6, 15, 30, 0, 0, time.UTC),
	}
	for _, p := range pedidos {
		e.gestor.pedidos[p.ID] = p
	}
	e.gestion = NuevaGestion(e.gestor, e.stock, e.tx, func() time.Time { return e.ahora })
	return e
}

func nuevoPedido(modo domain.ModoEntrega, estado domain.EstadoPedido, pago domain.EstadoPago, medio domain.MedioPago) *domain.Pedido {
	return &domain.Pedido{ID: uuid.New(), ModoEntrega: modo, EstadoPedido: estado, EstadoPago: pago, MedioPago: medio}
}

// ---- despachar / entregar ----

func TestGestion_Despachar_EnvioPagado(t *testing.T) {
	p := nuevoPedido(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago)
	e := nuevoEscenarioGestion(p)

	res, err := e.gestion.Despachar(context.Background(), p.ID)
	if err != nil || res.EstadoPedido != domain.PedidoDespachado {
		t.Fatalf("debía quedar DESPACHADO: %v, %v", res.EstadoPedido, err)
	}
}

// La demo del martes: despachar un pedido que todavía no fue pagado es un 409,
// no un 500, y no toca nada.
func TestGestion_Despachar_NoPagado_409SinTocarNada(t *testing.T) {
	p := nuevoPedido(domain.EntregaEnvio, domain.PedidoPendienteDePago, domain.PagoPendiente, domain.PagoMercadoPago)
	e := nuevoEscenarioGestion(p)

	_, err := e.gestion.Despachar(context.Background(), p.ID)
	if !esRegla(err) {
		t.Fatalf("esperaba ErrRegla (409), vino %v", err)
	}
	if e.tx.abiertas != 0 || p.EstadoPedido != domain.PedidoPendienteDePago {
		t.Error("no debía abrirse la transacción ni cambiar el estado")
	}
}

func TestGestion_Despachar_Inexistente_404(t *testing.T) {
	e := nuevoEscenarioGestion()
	if _, err := e.gestion.Despachar(context.Background(), uuid.New()); !errors.Is(err, apierr.ErrNoEncontrado) {
		t.Fatalf("esperaba 404, vino %v", err)
	}
}

func TestGestion_Entregar(t *testing.T) {
	envio := nuevoPedido(domain.EntregaEnvio, domain.PedidoDespachado, domain.PagoAprobado, domain.PagoMercadoPago)
	retiro := nuevoPedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago)
	sinDespachar := nuevoPedido(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago)
	e := nuevoEscenarioGestion(envio, retiro, sinDespachar)

	for _, p := range []*domain.Pedido{envio, retiro} {
		if res, err := e.gestion.Entregar(context.Background(), p.ID); err != nil || res.EstadoPedido != domain.PedidoCompletado {
			t.Errorf("debía completarse: %v, %v", res.EstadoPedido, err)
		}
	}
	if _, err := e.gestion.Entregar(context.Background(), sinDespachar.ID); !esRegla(err) {
		t.Errorf("un envío sin despachar no se entrega: %v", err)
	}
}

// ---- cobro en efectivo ----

func TestGestion_CobrarEnEfectivo_CobroEntregaYStockEnUnaAccion(t *testing.T) {
	p := nuevoPedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoPendiente, domain.PagoEfectivo)
	e := nuevoEscenarioGestion(p)

	res, err := e.gestion.CobrarEnEfectivo(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.EstadoPedido != domain.PedidoCompletado || res.EstadoPago != domain.PagoAprobado {
		t.Errorf("debía quedar COMPLETADO y cobrado: %s / %s", res.EstadoPedido, res.EstadoPago)
	}
	if e.stock.confirmaciones != 1 || e.tx.confirmadas != 1 {
		t.Errorf("debía convertir la reserva en venta una vez: confirmaciones=%d", e.stock.confirmaciones)
	}

	// Repetir el cobro no duplica nada.
	if _, err := e.gestion.CobrarEnEfectivo(context.Background(), p.ID); !esRegla(err) {
		t.Errorf("el segundo cobro debe dar 409: %v", err)
	}
	if e.stock.confirmaciones != 1 {
		t.Error("el segundo cobro no debe confirmar stock otra vez")
	}
}

// ---- cancelar ----

func TestGestion_Cancelar_NoPagado_LiberaStockYGuardaQuienCuandoPorQue(t *testing.T) {
	p := nuevoPedido(domain.EntregaRetiro, domain.PedidoPendienteDePago, domain.PagoPendiente, domain.PagoMercadoPago)
	e := nuevoEscenarioGestion(p)
	admin := uuid.New()

	res, err := e.gestion.Cancelar(context.Background(), p.ID, admin, "El cliente lo pidió por WhatsApp")
	if err != nil {
		t.Fatal(err)
	}
	if res.Pedido.EstadoPedido != domain.PedidoCancelado || e.stock.liberaciones != 1 {
		t.Errorf("debía quedar CANCELADO y liberar stock: %s, liberaciones=%d", res.Pedido.EstadoPedido, e.stock.liberaciones)
	}
	if res.Aviso != "" {
		t.Errorf("un pedido no pagado no lleva aviso de reintegro: %q", res.Aviso)
	}
	c := res.Pedido
	if c.CanceladoPor == nil || *c.CanceladoPor != admin || c.CanceladoEn == nil || !c.CanceladoEn.Equal(e.ahora) ||
		c.MotivoCancelacion == nil || *c.MotivoCancelacion != "El cliente lo pidió por WhatsApp" {
		t.Errorf("debía guardar responsable, fecha y motivo: %+v", c)
	}
}

func TestGestion_Cancelar_Pagado_AvisaQueElReintegroEsManual(t *testing.T) {
	p := nuevoPedido(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago)
	e := nuevoEscenarioGestion(p)

	res, err := e.gestion.Cancelar(context.Background(), p.ID, uuid.New(), "Devolución")
	if err != nil {
		t.Fatal(err)
	}
	if res.Aviso != "El reintegro debe realizarse manualmente en Mercado Pago. Este sistema no ejecuta devoluciones." {
		t.Errorf("aviso inesperado: %q", res.Aviso)
	}
}

// Pregunta de la defensa: "¿qué pasa si cancelo dos veces el mismo pedido?"
func TestGestion_Cancelar_DosVeces_NoDevuelveElStockDosVeces(t *testing.T) {
	p := nuevoPedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoPendiente, domain.PagoEfectivo)
	e := nuevoEscenarioGestion(p)

	if _, err := e.gestion.Cancelar(context.Background(), p.ID, uuid.New(), "primero"); err != nil {
		t.Fatal(err)
	}
	_, err := e.gestion.Cancelar(context.Background(), p.ID, uuid.New(), "segundo")
	if !esRegla(err) {
		t.Fatalf("la segunda cancelación debe dar 409: %v", err)
	}
	if e.stock.liberaciones != 1 {
		t.Errorf("el stock se libera UNA sola vez, no %d", e.stock.liberaciones)
	}
	if *p.MotivoCancelacion != "primero" {
		t.Error("la segunda cancelación no debe pisar quién/por qué canceló")
	}
}

func TestGestion_Cancelar_Completado_Nunca(t *testing.T) {
	p := nuevoPedido(domain.EntregaEnvio, domain.PedidoCompletado, domain.PagoAprobado, domain.PagoMercadoPago)
	e := nuevoEscenarioGestion(p)

	if _, err := e.gestion.Cancelar(context.Background(), p.ID, uuid.New(), "x"); !esRegla(err) {
		t.Fatalf("un pedido COMPLETADO no se cancela: %v", err)
	}
	if e.stock.liberaciones != 0 || p.EstadoPedido != domain.PedidoCompletado {
		t.Error("no debe liberar stock ni cambiar el estado")
	}
}

// Dos administradores cancelan a la vez: ambos leen CONFIRMADO, pero solo uno
// logra el UPDATE. El otro recibe 409 ANTES de tocar el stock.
func TestGestion_Cancelar_Simultanea_SoloUnaGana(t *testing.T) {
	p := nuevoPedido(domain.EntregaRetiro, domain.PedidoConfirmado, domain.PagoPendiente, domain.PagoEfectivo)
	e := nuevoEscenarioGestion(p)
	e.gestor.alLeer = func(real *domain.Pedido) { real.EstadoPedido = domain.PedidoCancelado } // el otro admin ganó

	_, err := e.gestion.Cancelar(context.Background(), p.ID, uuid.New(), "yo llegué segundo")
	if !esRegla(err) {
		t.Fatalf("esperaba 409 por cambio concurrente, vino %v", err)
	}
	if e.stock.liberaciones != 0 {
		t.Error("el que pierde la carrera no debe liberar stock")
	}
}

// ---- listado ----

func TestGestion_Listar_PasaFiltrosYPaginaSegura(t *testing.T) {
	e := nuevoEscenarioGestion(
		nuevoPedido(domain.EntregaEnvio, domain.PedidoConfirmado, domain.PagoAprobado, domain.PagoMercadoPago),
		nuevoPedido(domain.EntregaEnvio, domain.PedidoCancelado, domain.PagoPendiente, domain.PagoMercadoPago),
	)
	estado := domain.PedidoCancelado

	pagina, err := e.gestion.Listar(context.Background(), out.FiltrosPedidos{Estado: &estado}, 2, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if len(pagina.Pedidos) != 1 || pagina.Pedidos[0].EstadoPedido != domain.PedidoCancelado {
		t.Errorf("el filtro por estado no se aplicó: %+v", pagina.Pedidos)
	}
	if e.gestor.ultimoLimit != TamanoPaginaMaximo || e.gestor.ultimoOffset != TamanoPaginaMaximo {
		t.Errorf("tamaño topeado en %d: limit=%d offset=%d", TamanoPaginaMaximo, e.gestor.ultimoLimit, e.gestor.ultimoOffset)
	}
}
