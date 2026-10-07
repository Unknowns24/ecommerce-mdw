package out

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

// ItemReservaStock es una línea de reserva: cuántas unidades de una variante
// se comprometen para un pedido. Replica el contrato congelado de Agustín
// (usecases/inventario.ItemReserva).
type ItemReservaStock struct {
	VarianteID uuid.UUID
	Unidades   int
}

// ReservadorStock reserva unidades dentro de la transacción del checkout. Si
// no alcanzan, devuelve apierr.ErrRegla con Datos = mapa variante → unidades
// disponibles. Lo implementa ServicioStock (Agustín) a través de un adaptador.
type ReservadorStock interface {
	Reservar(ctx context.Context, tx *gorm.DB, pedidoID uuid.UUID, items []ItemReservaStock) error
}

// CreadorDePedidos guarda un pedido con su detalle dentro de la transacción tx.
type CreadorDePedidos interface {
	Crear(ctx context.Context, tx *gorm.DB, p *domain.Pedido, detalles []domain.DetallePedido) error
}

// LectorConfiguracion entrega la configuración vigente de la tienda.
type LectorConfiguracion interface {
	Actual(ctx context.Context) (domain.ConfiguracionTienda, error)
	TarifasOrdenadas(ctx context.Context) ([]domain.TarifaDistancia, error)
}

// Transaccionador ejecuta fn dentro de una transacción: si fn devuelve error,
// se revierte todo lo que hizo; si no, se confirma.
type Transaccionador interface {
	EnTransaccion(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// LectorPedidos son las consultas de lectura que el comprador puede hacer. Las
// de datos propios llevan el id del comprador: no existe una forma de pedir "el
// pedido X" sin decir de quién es.
type LectorPedidos interface {
	PorIDYUsuario(ctx context.Context, id, usuarioID uuid.UUID) (domain.Pedido, error)
	PorToken(ctx context.Context, token string) (domain.PedidoPublico, error)
	ListarDeUsuario(ctx context.Context, usuarioID uuid.UUID, limit, offset int) ([]domain.Pedido, int64, error)
}

// FiltrosPedidos son los filtros opcionales del listado administrativo.
type FiltrosPedidos struct {
	Estado *domain.EstadoPedido
	Desde  *time.Time // incluido
	Hasta  *time.Time // excluido
}

// GestorPedidos son las operaciones de la administración sobre los pedidos.
// Nada de esto filtra por comprador: solo se usa en rutas con el permiso
// pedidos.gestionar.
type GestorPedidos interface {
	PorID(ctx context.Context, id uuid.UUID) (domain.Pedido, error)
	ListarAdmin(ctx context.Context, f FiltrosPedidos, limit, offset int) ([]domain.Pedido, int64, error)
	CambiarEstado(ctx context.Context, tx *gorm.DB, id uuid.UUID, desde, nuevo domain.EstadoPedido) error
	MarcarPagoAprobado(ctx context.Context, tx *gorm.DB, id uuid.UUID) error
	RegistrarCancelacion(ctx context.Context, tx *gorm.DB, id, responsableID uuid.UUID, motivo string, ahora time.Time) error
}

// GestorStock es lo que las transiciones necesitan del inventario. Confirmar
// convierte la reserva en venta (no descuenta dos veces); Liberar devuelve las
// unidades reservadas. Ambos son idempotentes (contrato de Agustín).
type GestorStock interface {
	Confirmar(ctx context.Context, tx *gorm.DB, pedidoID uuid.UUID) error
	Liberar(ctx context.Context, tx *gorm.DB, pedidoID uuid.UUID) error
}

// RepositorioCarrito guarda el carrito de los clientes registrados. Todas las
// operaciones sobre una línea llevan el id del carrito: una línea de otro
// carrito, para quien llama, no existe.
type RepositorioCarrito interface {
	PorUsuario(ctx context.Context, usuarioID uuid.UUID) (domain.Carrito, error)
	CrearSiNoExiste(ctx context.Context, usuarioID uuid.UUID) (domain.Carrito, error)
	AgregarItem(ctx context.Context, carritoID, varianteID uuid.UUID, unidades int) (domain.ItemCarrito, error)
	CambiarUnidades(ctx context.Context, carritoID, itemID uuid.UUID, unidades int) error
	QuitarItem(ctx context.Context, carritoID, itemID uuid.UUID) error
	Vaciar(ctx context.Context, carritoID uuid.UUID) error
}

// ConsultaStock informa cuántas unidades vendibles tiene una variante
// (existencias menos reservas). Lo implementa ServicioStock.Disponible
// (Agustín), cuya firma coincide exactamente con esta.
type ConsultaStock interface {
	Disponible(ctx context.Context, varianteID uuid.UUID) (int, error)
}
