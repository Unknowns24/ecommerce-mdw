package pedidos

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// LineaCarrito es una línea del carrito tal como se muestra: valorizada con el
// catálogo VIGENTE (el carrito no congela precios). Si la variante dejó de
// existir o se desactivó, Disponible es false, no tiene precio y no suma al
// total estimado.
type LineaCarrito struct {
	ItemID                 uuid.UUID
	VarianteID             uuid.UUID
	Codigo, Nombre         string
	Unidades               int
	PrecioUnitarioCentavos int64
	SubtotalCentavos       int64
	Disponible             bool
}

// VistaCarrito es el carrito con su total estimado.
type VistaCarrito struct {
	Lineas                []LineaCarrito
	TotalEstimadoCentavos int64
}

// GestionCarrito es el carrito del cliente registrado (SPEC-H07). El del
// invitado vive en el navegador (localStorage) y nunca llega acá: para
// comprar, el invitado manda sus items en el body del checkout.
//
// El carrito NO reserva unidades ni congela precios: los totales se recalculan
// con el catálogo vigente en cada lectura. Que algo esté en el carrito no
// garantiza nada hasta el checkout, y por eso el checkout revalida todo.
type GestionCarrito struct {
	carritos out.RepositorioCarrito
	catalogo out.LectorCatalogo
	stock    out.ConsultaStock
}

func NuevaGestionCarrito(carritos out.RepositorioCarrito, catalogo out.LectorCatalogo, stock out.ConsultaStock) *GestionCarrito {
	return &GestionCarrito{carritos: carritos, catalogo: catalogo, stock: stock}
}

// ValidarDisponibilidad decide si se pueden tener `total` unidades de una
// variante cuando hay `disponible`. Pedir más que el disponible es un 409 y no
// un 400: el request está perfecto, es el estado del sistema el que no lo
// permite. Los Datos enumeran variante → disponible para que la pantalla no
// tenga que parsear castellano.
func ValidarDisponibilidad(varianteID uuid.UUID, total, disponible int) error {
	if total <= disponible {
		return nil
	}
	return apierr.ErrRegla{
		Mensaje: fmt.Sprintf("No hay stock suficiente: podés pedir hasta %d unidades de este producto.", disponible),
		Datos:   map[string]int{varianteID.String(): disponible},
	}
}

func validarUnidades(unidades int) error {
	if unidades < 1 {
		return apierr.ErrValidacion{Campos: map[string]string{"unidades": "Las unidades deben ser un entero positivo"}}
	}
	return nil
}

// Ver devuelve el carrito del usuario, o uno vacío si todavía no tiene.
func (g *GestionCarrito) Ver(ctx context.Context, usuarioID uuid.UUID) (VistaCarrito, error) {
	c, err := g.carritos.PorUsuario(ctx, usuarioID)
	if errors.Is(err, apierr.ErrNoEncontrado) {
		return VistaCarrito{Lineas: []LineaCarrito{}}, nil
	}
	if err != nil {
		return VistaCarrito{}, err
	}
	return g.vista(ctx, c)
}

// AgregarItem agrega una variante; si ya estaba, suma las unidades. Devuelve el
// carrito y si la línea es nueva.
func (g *GestionCarrito) AgregarItem(ctx context.Context, usuarioID, varianteID uuid.UUID, unidades int) (VistaCarrito, bool, error) {
	if err := validarUnidades(unidades); err != nil {
		return VistaCarrito{}, false, err
	}
	// 404 si la variante no existe o está inactiva.
	if _, err := g.catalogo.VariantePorID(ctx, varianteID); err != nil {
		return VistaCarrito{}, false, err
	}
	c, err := g.carritos.CrearSiNoExiste(ctx, usuarioID)
	if err != nil {
		return VistaCarrito{}, false, err
	}

	yaEnCarrito := 0
	for _, it := range c.Items {
		if it.VarianteID == varianteID {
			yaEnCarrito += it.Unidades
		}
	}
	disponible, err := g.stock.Disponible(ctx, varianteID)
	if err != nil {
		return VistaCarrito{}, false, err
	}
	// Lo que ya tenía más lo que agrega: no se puede superar el disponible.
	if err := ValidarDisponibilidad(varianteID, yaEnCarrito+unidades, disponible); err != nil {
		return VistaCarrito{}, false, err
	}

	if _, err := g.carritos.AgregarItem(ctx, c.ID, varianteID, unidades); err != nil {
		return VistaCarrito{}, false, err
	}
	v, err := g.Ver(ctx, usuarioID)
	return v, yaEnCarrito == 0, err
}

// CambiarUnidades fija las unidades de una línea del carrito del usuario. Una
// línea que no está en SU carrito no existe: 404.
func (g *GestionCarrito) CambiarUnidades(ctx context.Context, usuarioID, itemID uuid.UUID, unidades int) (VistaCarrito, error) {
	if err := validarUnidades(unidades); err != nil {
		return VistaCarrito{}, err
	}
	c, item, err := g.itemPropio(ctx, usuarioID, itemID)
	if err != nil {
		return VistaCarrito{}, err
	}
	disponible, err := g.stock.Disponible(ctx, item.VarianteID)
	if err != nil {
		return VistaCarrito{}, err
	}
	if err := ValidarDisponibilidad(item.VarianteID, unidades, disponible); err != nil {
		return VistaCarrito{}, err
	}
	if err := g.carritos.CambiarUnidades(ctx, c.ID, itemID, unidades); err != nil {
		return VistaCarrito{}, err
	}
	return g.Ver(ctx, usuarioID)
}

// QuitarItem borra una línea del carrito del usuario.
func (g *GestionCarrito) QuitarItem(ctx context.Context, usuarioID, itemID uuid.UUID) error {
	c, _, err := g.itemPropio(ctx, usuarioID, itemID)
	if err != nil {
		return err
	}
	return g.carritos.QuitarItem(ctx, c.ID, itemID)
}

// Vaciar borra todas las líneas. Sin carrito no hay nada que vaciar: es un éxito.
func (g *GestionCarrito) Vaciar(ctx context.Context, usuarioID uuid.UUID) error {
	c, err := g.carritos.PorUsuario(ctx, usuarioID)
	if errors.Is(err, apierr.ErrNoEncontrado) {
		return nil
	}
	if err != nil {
		return err
	}
	return g.carritos.Vaciar(ctx, c.ID)
}

// itemPropio busca el carrito del usuario y una línea DENTRO de él.
func (g *GestionCarrito) itemPropio(ctx context.Context, usuarioID, itemID uuid.UUID) (domain.Carrito, domain.ItemCarrito, error) {
	c, err := g.carritos.PorUsuario(ctx, usuarioID)
	if err != nil {
		return domain.Carrito{}, domain.ItemCarrito{}, err // sin carrito: 404
	}
	for _, it := range c.Items {
		if it.ID == itemID {
			return c, it, nil
		}
	}
	return domain.Carrito{}, domain.ItemCarrito{}, apierr.ErrNoEncontrado
}

// vista valoriza el carrito con el catálogo vigente.
func (g *GestionCarrito) vista(ctx context.Context, c domain.Carrito) (VistaCarrito, error) {
	lineas := make([]LineaCarrito, 0, len(c.Items))
	vendibles := make([]LineaCalculada, 0, len(c.Items))

	for _, it := range c.Items {
		linea := LineaCarrito{ItemID: it.ID, VarianteID: it.VarianteID, Unidades: it.Unidades}
		v, err := g.catalogo.VariantePorID(ctx, it.VarianteID)
		switch {
		case errors.Is(err, apierr.ErrNoEncontrado):
			// La variante se desactivó o se borró: la línea queda visible pero
			// marcada, sin precio, y no suma al total.
		case err != nil:
			return VistaCarrito{}, err
		default:
			linea.Codigo, linea.Nombre, linea.Disponible = v.Codigo, v.Nombre, true
			linea.PrecioUnitarioCentavos = v.PrecioMinoristaCentavos
			linea.SubtotalCentavos = int64(it.Unidades) * v.PrecioMinoristaCentavos
			vendibles = append(vendibles, LineaCalculada{SubtotalCentavos: linea.SubtotalCentavos})
		}
		lineas = append(lineas, linea)
	}

	total, _ := CalcularTotal(vendibles, 0)
	return VistaCarrito{Lineas: lineas, TotalEstimadoCentavos: total}, nil
}
