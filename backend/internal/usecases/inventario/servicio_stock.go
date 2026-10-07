package inventario

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/inventario"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

// ItemReserva is one line of a reservation request: how many units of a
// variant to commit toward a pedido.
type ItemReserva struct {
	VarianteID uuid.UUID
	Unidades   int
}

// ServicioStock is the frozen contract the other modules consume (Nicolás's
// checkout reserves, confirms and releases through it; Genaro's public
// catalog reads Disponible). Its signature does not change without
// coordinating with them.
type ServicioStock struct {
	db *gorm.DB
}

func NuevoServicioStock(db *gorm.DB) *ServicioStock {
	return &ServicioStock{db: db}
}

// Disponible devuelve las unidades vendibles de una variante: existencias de
// lotes menos reservas activas.
func (s *ServicioStock) Disponible(ctx context.Context, varianteID uuid.UUID) (int, error) {
	metodologia, err := metodologiaDeVariante(ctx, s.db, varianteID)
	if err != nil {
		return 0, err
	}
	lotes, err := lotesDisponiblesDeVariante(ctx, s.db, varianteID, metodologia, false)
	if err != nil {
		return 0, err
	}
	return Disponible(lotes), nil
}

// Reservar crea una fila ACTIVA en reserva_stock por cada ítem, asignando
// lotes por FIFO/LIFO según la metodología de cada producto. Corre dentro de
// la transacción que le pasa el checkout: lee los lotes con bloqueo para que
// dos checkouts simultáneos no reserven la misma última unidad. Si falta
// stock para cualquier ítem, no reserva nada y devuelve apierr.ErrRegla con
// el mapa variante → unidades disponibles que el checkout enumera al
// comprador.
func (s *ServicioStock) Reservar(ctx context.Context, tx *gorm.DB, pedidoID uuid.UUID, items []ItemReserva) error {
	type asignacionPendiente struct {
		loteID   uuid.UUID
		unidades int
	}

	var pendientes []asignacionPendiente
	faltantes := map[string]int{}

	for _, item := range items {
		metodologia, err := metodologiaDeVariante(ctx, tx, item.VarianteID)
		if err != nil {
			return err
		}

		lotes, err := lotesDisponiblesDeVariante(ctx, tx, item.VarianteID, metodologia, true)
		if err != nil {
			return err
		}

		asignaciones, err := AsignarLotes(lotes, item.Unidades, metodologia)
		if err != nil {
			if errors.Is(err, ErrStockInsuficiente) {
				faltantes[item.VarianteID.String()] = Disponible(lotes)
				continue
			}
			return err
		}

		for _, asignacion := range asignaciones {
			pendientes = append(pendientes, asignacionPendiente{loteID: asignacion.LoteID, unidades: asignacion.Unidades})
		}
	}

	if len(faltantes) > 0 {
		return apierr.ErrRegla{Mensaje: "No hay stock suficiente", Datos: faltantes}
	}

	reservas := repo.NuevoRepositorioReservas(tx)
	for _, p := range pendientes {
		reserva := &domain.ReservaStock{
			ID:       uuid.New(),
			PedidoID: pedidoID,
			LoteID:   p.loteID,
			Unidades: p.unidades,
			Estado:   domain.ReservaActiva,
		}
		if err := reservas.Crear(ctx, reserva); err != nil {
			return err
		}
	}

	return nil
}

// Confirmar pasa las reservas ACTIVA del pedido a CONSUMIDA y genera un
// MovimientoStock de tipo SALIDA por cada una. Es idempotente: si ya están
// consumidas (no quedan ACTIVA), no hace nada y no devuelve error — una
// notificación repetida de Mercado Pago no puede descontar dos veces.
func (s *ServicioStock) Confirmar(ctx context.Context, tx *gorm.DB, pedidoID uuid.UUID) error {
	reservasRepo := repo.NuevoRepositorioReservas(tx)
	activas, err := reservasRepo.ActivasPorPedido(ctx, pedidoID)
	if err != nil {
		return err
	}

	movimientosRepo := repo.NuevoRepositorioMovimientos(tx)
	for _, reserva := range activas {
		if err := reservasRepo.CambiarEstado(ctx, reserva.ID, domain.ReservaConsumida); err != nil {
			return err
		}
		movimiento := &domain.MovimientoStock{
			ID:       uuid.New(),
			LoteID:   reserva.LoteID,
			Tipo:     domain.MovimientoSalida,
			Unidades: reserva.Unidades,
			Motivo:   "Venta confirmada",
			PedidoID: &pedidoID,
		}
		if err := movimientosRepo.Crear(ctx, movimiento); err != nil {
			return err
		}
	}
	return nil
}

// Liberar pasa las reservas ACTIVA del pedido a LIBERADA. Es idempotente y
// no genera movimiento: liberar una reserva no es una entrada física.
func (s *ServicioStock) Liberar(ctx context.Context, tx *gorm.DB, pedidoID uuid.UUID) error {
	reservasRepo := repo.NuevoRepositorioReservas(tx)
	activas, err := reservasRepo.ActivasPorPedido(ctx, pedidoID)
	if err != nil {
		return err
	}
	for _, reserva := range activas {
		if err := reservasRepo.CambiarEstado(ctx, reserva.ID, domain.ReservaLiberada); err != nil {
			return err
		}
	}
	return nil
}

// metodologiaDeVariante consulta la tabla de catálogo directamente (misma
// base, módulo de Genaro) porque el lector congelado de catálogo no expone
// la metodología de rotación del producto, que es justamente lo que esta
// capa necesita para decidir FIFO o LIFO.
func metodologiaDeVariante(ctx context.Context, db *gorm.DB, varianteID uuid.UUID) (string, error) {
	var metodologia string
	err := db.WithContext(ctx).
		Table("variante").
		Select("producto.metodologia_rotacion").
		Joins("JOIN producto ON producto.id = variante.producto_id").
		Where("variante.id = ?", varianteID).
		Scan(&metodologia).Error
	if err != nil {
		return "", err
	}
	if metodologia == "" {
		return "", apierr.ErrNoEncontrado
	}
	return metodologia, nil
}

// lotesDisponiblesDeVariante arma la lista de LoteDisponible de una variante:
// unidades netas (ingresadas + ajustes - salidas) y unidades ya
// comprometidas en reservas ACTIVA. Con bloquear=true lee los lotes con
// SELECT ... FOR UPDATE, para usar dentro de la transacción del checkout.
func lotesDisponiblesDeVariante(ctx context.Context, db *gorm.DB, varianteID uuid.UUID, metodologia string, bloquear bool) ([]LoteDisponible, error) {
	lotesRepo := repo.NuevoRepositorioLotes(db)
	var lotes []domain.Lote
	var err error
	if bloquear {
		lotes, err = lotesRepo.PorVarianteBloqueando(ctx, varianteID, metodologia)
	} else {
		lotes, err = lotesRepo.PorVariante(ctx, varianteID, metodologia)
	}
	if err != nil {
		return nil, err
	}
	if len(lotes) == 0 {
		return nil, nil
	}

	loteIDs := make([]uuid.UUID, len(lotes))
	for i, lote := range lotes {
		loteIDs[i] = lote.ID
	}

	neto, err := repo.NuevoRepositorioMovimientos(db).NetoPorLotes(ctx, loteIDs)
	if err != nil {
		return nil, err
	}
	activas, err := repo.NuevoRepositorioReservas(db).ActivasPorLotes(ctx, loteIDs)
	if err != nil {
		return nil, err
	}

	disponibles := make([]LoteDisponible, len(lotes))
	for i, lote := range lotes {
		disponibles[i] = LoteDisponible{
			LoteID:       lote.ID,
			FechaIngreso: lote.FechaIngreso,
			Unidades:     lote.UnidadesIngresadas + neto[lote.ID],
			Reservadas:   activas[lote.ID],
		}
	}
	return disponibles, nil
}
