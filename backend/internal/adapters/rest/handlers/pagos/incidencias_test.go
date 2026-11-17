package pagos

import (
	"context"
	"errors"
	"testing"
	"time"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/pagos"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/external/mp"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	stock "github.com/Unknowns24/ecommerce-mdw/internal/usecases/inventario"
	reglas "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pagos"
	"github.com/google/uuid"
)

func TestIncidenciasPersistenSinDuplicarNiReabrir(t *testing.T) {
	db := conectarParaTest(t)
	ctx := context.Background()
	pedido := pedidoConReserva(t, db, time.Now().Add(-time.Minute))
	actor := uuid.New()
	if err := db.Exec("INSERT INTO usuario (id) VALUES (?)", actor).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM incidencia_pago WHERE pedido_id=?", pedido)
		db.Exec("DELETE FROM usuario WHERE id=?", actor)
	})
	pago := mp.Pago{ID: uuid.NewString(), Estado: "approved", MontoCentavos: 100000, ReferenciaExterna: pedido.String()}
	repositorio := repo.NuevoRepositorioIncidencias(db)
	servicio := reglas.NuevaGestionIncidencias(repositorio, func() time.Time { return time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC) })
	procesar := func() {
		t.Helper()
		if err := procesarNotificacion(ctx, db, stock.NuevoServicioStock(db), repo.NuevoRepositorioIntentos(db), pedido, pago); err != nil {
			t.Fatal(err)
		}
	}
	procesar()
	procesar()
	items, err := repositorio.Listar(ctx, pedido)
	if err != nil || len(items) != 1 {
		t.Fatalf("incidencias=%v, error=%v", items, err)
	}
	i := items[0]
	if i.Motivo != "MONTO_DISTINTO" || i.MontoCentavos != 100000 || i.TotalPedidoCentavos != 150000 {
		t.Fatalf("incidencia incorrecta: %+v", i)
	}
	if err := servicio.Resolver(ctx, uuid.New(), i.ID, actor, "Verificado"); !errors.Is(err, apierr.ErrNoEncontrado) {
		t.Fatalf("otro pedido: %v", err)
	}
	if err := servicio.Resolver(ctx, pedido, i.ID, actor, "  Reintegro externo verificado  "); err != nil {
		t.Fatal(err)
	}
	if err := servicio.Resolver(ctx, pedido, i.ID, actor, "No debe sobrescribir"); err != nil {
		t.Fatal(err)
	}
	procesar()
	items, err = repositorio.Listar(ctx, pedido)
	if err != nil || len(items) != 1 {
		t.Fatalf("incidencias=%v, error=%v", items, err)
	}
	i = items[0]
	if i.ResueltoPor == nil || *i.ResueltoPor != actor || i.ResueltoEn == nil || i.NotaResolucion == nil || *i.NotaResolucion != "Reintegro externo verificado" {
		t.Fatalf("se perdió la primera resolución: %+v", i)
	}
	if p, r := estadoDe(t, db, pedido); p != "PENDIENTE_DE_PAGO" || r != "ACTIVA" {
		t.Fatalf("cerrar una incidencia alteró pedido/stock: %s/%s", p, r)
	}
}

func TestSegundoCobroGeneraIncidenciaSinOtraSalida(t *testing.T) {
	db := conectarParaTest(t)
	ctx := context.Background()
	pedido := pedidoConReserva(t, db, time.Now().Add(time.Minute))
	pago := mp.Pago{ID: uuid.NewString(), Estado: "approved", MontoCentavos: 150000, ReferenciaExterna: pedido.String()}
	s := stock.NuevoServicioStock(db)
	for _, id := range []string{pago.ID, pago.ID, uuid.NewString()} {
		pago.ID = id
		if err := procesarNotificacion(ctx, db, s, repo.NuevoRepositorioIntentos(db), pedido, pago); err != nil {
			t.Fatal(err)
		}
	}
	items, err := repo.NuevoRepositorioIncidencias(db).Listar(ctx, pedido)
	if err != nil || len(items) != 1 || items[0].Motivo != "PEDIDO_NO_PENDIENTE" {
		t.Fatalf("incidencias=%v, error=%v", items, err)
	}
	var salidas int64
	if err := db.Model(&domain.MovimientoStock{}).Where("pedido_id=? AND tipo='SALIDA'", pedido).Count(&salidas).Error; err != nil {
		t.Fatal(err)
	}
	if salidas != 1 {
		t.Fatalf("salidas=%d", salidas)
	}
}

type pagosEnProceso struct{ consultados map[string]bool }

func (f pagosEnProceso) PagosDelPedido(_ context.Context, id string) ([]mp.Pago, error) {
	f.consultados[id] = true
	return []mp.Pago{{Estado: "in_process"}}, nil
}

func TestConciliacionRotaMasDeUnLoteDePendientes(t *testing.T) {
	db := conectarParaTest(t)
	ahora := time.Now()
	ids := make([]uuid.UUID, 0, loteVencimiento+1)
	for range loteVencimiento + 1 {
		ids = append(ids, pedidoConReserva(t, db, ahora.Add(-time.Minute)))
	}
	f := pagosEnProceso{consultados: map[string]bool{}}
	for n := range 2 {
		if _, err := vencerPendientes(context.Background(), db, stock.NuevoServicioStock(db), f, ahora.Add(time.Duration(n)*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range ids {
		if !f.consultados[id.String()] {
			t.Fatalf("el pedido %s quedó fuera de ambos lotes", id)
		}
	}
}
