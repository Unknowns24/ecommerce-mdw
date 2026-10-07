// Package pedidos contiene las consultas a la base del carrito, el pedido y
// la configuración de la tienda. Las consultas viven acá y en ningún otro
// lado: el handler nunca arma un db.Where, y las reglas de negocio viven en
// usecases.
package pedidos

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

// traducir convierte "la fila no existe" de GORM en el error del contrato de
// la API; cualquier otro error se devuelve tal cual (terminará en un 500).
func traducir(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apierr.ErrNoEncontrado
	}
	return err
}

// RepositorioCarritos guarda el carrito de los clientes registrados. El del
// invitado vive en el navegador y nunca llega a la base.
type RepositorioCarritos struct{ db *gorm.DB }

func NuevoRepositorioCarritos(db *gorm.DB) *RepositorioCarritos {
	return &RepositorioCarritos{db: db}
}

// PorUsuario devuelve el carrito del usuario con sus líneas, o
// ErrNoEncontrado si todavía no tiene.
func (r *RepositorioCarritos) PorUsuario(ctx context.Context, usuarioID uuid.UUID) (domain.Carrito, error) {
	var c domain.Carrito
	err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("creado_en") }).
		Where("usuario_id = ?", usuarioID).
		First(&c).Error
	return c, traducir(err)
}

// CrearSiNoExiste devuelve el carrito del usuario, creándolo si hace falta.
// Usa INSERT ... ON CONFLICT DO NOTHING: si dos requests llegan a la vez, el
// UNIQUE de usuario_id deja pasar uno solo y el otro simplemente lo lee.
func (r *RepositorioCarritos) CrearSiNoExiste(ctx context.Context, usuarioID uuid.UUID) (domain.Carrito, error) {
	nuevo := domain.Carrito{UsuarioID: &usuarioID}
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "usuario_id"}}, DoNothing: true}).
		Create(&nuevo).Error
	if err != nil {
		return domain.Carrito{}, err
	}
	return r.PorUsuario(ctx, usuarioID)
}

// AgregarItem suma unidades a la línea de esa variante, o la crea si no
// estaba. Es un solo INSERT ... ON CONFLICT DO UPDATE: no hay un "leer y
// después escribir" donde dos requests simultáneos puedan pisarse.
func (r *RepositorioCarritos) AgregarItem(ctx context.Context, carritoID, varianteID uuid.UUID, unidades int) (domain.ItemCarrito, error) {
	item := domain.ItemCarrito{CarritoID: carritoID, VarianteID: varianteID, Unidades: unidades}
	err := r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "carrito_id"}, {Name: "variante_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"unidades": gorm.Expr("item_carrito.unidades + EXCLUDED.unidades"),
			}),
		},
		clause.Returning{},
	).Create(&item).Error
	if err != nil {
		return domain.ItemCarrito{}, err
	}
	return item, r.tocar(ctx, carritoID)
}

// CambiarUnidades fija las unidades de una línea. El WHERE lleva el carrito:
// una línea de otro carrito, para esta llamada, no existe.
func (r *RepositorioCarritos) CambiarUnidades(ctx context.Context, carritoID, itemID uuid.UUID, unidades int) error {
	res := r.db.WithContext(ctx).Model(&domain.ItemCarrito{}).
		Where("id = ? AND carrito_id = ?", itemID, carritoID).
		Update("unidades", unidades)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apierr.ErrNoEncontrado
	}
	return r.tocar(ctx, carritoID)
}

// QuitarItem borra una línea del carrito (es una selección editable, no
// historial de ventas).
func (r *RepositorioCarritos) QuitarItem(ctx context.Context, carritoID, itemID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND carrito_id = ?", itemID, carritoID).
		Delete(&domain.ItemCarrito{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apierr.ErrNoEncontrado
	}
	return r.tocar(ctx, carritoID)
}

// Vaciar borra todas las líneas del carrito.
func (r *RepositorioCarritos) Vaciar(ctx context.Context, carritoID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("carrito_id = ?", carritoID).Delete(&domain.ItemCarrito{}).Error; err != nil {
		return err
	}
	return r.tocar(ctx, carritoID)
}

// tocar actualiza la fecha de última modificación del carrito.
func (r *RepositorioCarritos) tocar(ctx context.Context, carritoID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.Carrito{}).
		Where("id = ?", carritoID).
		Update("actualizado_en", time.Now()).Error
}
