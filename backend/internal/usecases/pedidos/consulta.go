package pedidos

import (
	"context"

	"github.com/google/uuid"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// Paginación del listado "mis pedidos".
const (
	TamanoPaginaPorDefecto = 20
	TamanoPaginaMaximo     = 100
)

// NormalizarPaginacion convierte lo que pidió el cliente (página desde 1 y
// tamaño) en límites seguros para la consulta: una página inválida es la 1, un
// tamaño ausente es el por defecto y nunca pasa del máximo (si no, un cliente
// podría pedir "tamaño 1.000.000" y traerse la tabla entera).
func NormalizarPaginacion(pagina, tamano int) (pag, tam, limit, offset int) {
	if pagina < 1 {
		pagina = 1
	}
	switch {
	case tamano < 1:
		tamano = TamanoPaginaPorDefecto
	case tamano > TamanoPaginaMaximo:
		tamano = TamanoPaginaMaximo
	}
	return pagina, tamano, tamano, (pagina - 1) * tamano
}

// PaginaPedidos es una página del listado.
type PaginaPedidos struct {
	Pedidos []domain.Pedido
	Total   int64
	Pagina  int
	Tamano  int
}

// Consulta atiende las lecturas de pedidos del comprador.
type Consulta struct{ pedidos out.LectorPedidos }

func NuevaConsulta(pedidos out.LectorPedidos) *Consulta { return &Consulta{pedidos: pedidos} }

// MisPedidos lista los pedidos del usuario de la sesión, más nuevos primero.
func (c *Consulta) MisPedidos(ctx context.Context, usuarioID uuid.UUID, pagina, tamano int) (PaginaPedidos, error) {
	pag, tam, limit, offset := NormalizarPaginacion(pagina, tamano)
	pedidos, total, err := c.pedidos.ListarDeUsuario(ctx, usuarioID, limit, offset)
	if err != nil {
		return PaginaPedidos{}, err
	}
	return PaginaPedidos{Pedidos: pedidos, Total: total, Pagina: pag, Tamano: tam}, nil
}

// Detalle devuelve un pedido del usuario. La búsqueda lleva el id del
// comprador adentro de la consulta: un pedido ajeno no existe para esta
// llamada y el error es ErrNoEncontrado (404), igual que si el id no existiera.
func (c *Consulta) Detalle(ctx context.Context, id, usuarioID uuid.UUID) (domain.Pedido, error) {
	return c.pedidos.PorIDYUsuario(ctx, id, usuarioID)
}

// Publico devuelve la vista pública del pedido al que pertenece el token.
func (c *Consulta) Publico(ctx context.Context, token string) (domain.PedidoPublico, error) {
	return c.pedidos.PorToken(ctx, token)
}
