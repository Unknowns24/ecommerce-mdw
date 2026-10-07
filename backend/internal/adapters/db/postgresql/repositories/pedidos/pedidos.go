package pedidos

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/ports/out"
)

// RepositorioPedidos concentra todas las consultas de pedidos.
type RepositorioPedidos struct{ db *gorm.DB }

func NuevoRepositorioPedidos(db *gorm.DB) *RepositorioPedidos {
	return &RepositorioPedidos{db: db}
}

// FiltrosAdmin son los filtros opcionales del listado administrativo (los define
// el puerto para que el caso de uso no dependa de este paquete).
type FiltrosAdmin = out.FiltrosPedidos

// Crear guarda el pedido y su detalle dentro de la transacción tx que abre el
// checkout: si algo falla después (por ejemplo, el stock no alcanza), todo se
// revierte junto.
func (r *RepositorioPedidos) Crear(ctx context.Context, tx *gorm.DB, p *domain.Pedido, detalles []domain.DetallePedido) error {
	tx = tx.WithContext(ctx)
	// Omit("Detalles"): el detalle se inserta abajo con su pedido_id; sin esto
	// GORM intentaría guardar también p.Detalles y quedaría duplicado.
	if err := tx.Omit("Detalles").Create(p).Error; err != nil {
		return err
	}
	for i := range detalles {
		detalles[i].PedidoID = p.ID
	}
	if len(detalles) > 0 {
		if err := tx.Create(&detalles).Error; err != nil {
			return err
		}
	}
	p.Detalles = detalles
	return nil
}

// PorIDYUsuario es la consulta que define la seguridad del módulo: las dos
// condiciones van en el WHERE. No es que el pedido ajeno se rechace después de
// traerlo: es que no existe para esa llamada. El 404 sale solo y no distingue
// "no existe" de "no es tuyo", que es justo lo que queremos. Cualquier
// endpoint nuevo que reuse esta función hereda la protección.
func (r *RepositorioPedidos) PorIDYUsuario(ctx context.Context, id, usuarioID uuid.UUID) (domain.Pedido, error) {
	var p domain.Pedido
	err := r.db.WithContext(ctx).
		Preload("Detalles").
		Where("id = ? AND usuario_id = ?", id, usuarioID).
		First(&p).Error
	return p, traducir(err)
}

// PorID busca un pedido sin filtrar por comprador. SOLO para endpoints
// administrativos protegidos con el permiso pedidos.gestionar: nunca debe
// usarse en una ruta de cliente.
func (r *RepositorioPedidos) PorID(ctx context.Context, id uuid.UUID) (domain.Pedido, error) {
	var p domain.Pedido
	err := r.db.WithContext(ctx).Preload("Detalles").Where("id = ?", id).First(&p).Error
	return p, traducir(err)
}

// PorToken busca por el secreto del enlace privado del invitado y devuelve
// SOLO la vista pública.
//
// Tiene su propio Select, más chico que el de las pantallas propias. Reusar el
// de "mis pedidos" sería el bug: el día que alguien le agregue un campo al
// pedido (el correo, el DNI, el usuario), ese campo se publicaría en internet
// sin que nadie lo decidiera. Acá se enumeran, a mano, las únicas columnas que
// salen. El token es un secreto: nunca se escribe en un log.
func (r *RepositorioPedidos) PorToken(ctx context.Context, token string) (domain.PedidoPublico, error) {
	db := r.db.WithContext(ctx)

	// Struct plano (sin la lista de detalles) para que GORM solo mapee columnas.
	var fila struct {
		Numero           int64
		ModoEntrega      domain.ModoEntrega
		Domicilio        *string
		DistanciaMetros  *int
		EnvioCentavos    int64
		SubtotalCentavos int64
		TotalCentavos    int64
		EstadoPedido     domain.EstadoPedido
		EstadoPago       domain.EstadoPago
		MedioPago        domain.MedioPago
		VenceEn          *time.Time
		CreadoEn         time.Time
	}
	res := db.Table("pedido").
		Select("numero, modo_entrega, domicilio, distancia_metros, envio_centavos, "+
			"subtotal_centavos, total_centavos, estado_pedido, estado_pago, medio_pago, vence_en, creado_en").
		Where("token_acceso = ?", token).
		Limit(1).
		Scan(&fila)
	if res.Error != nil {
		return domain.PedidoPublico{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.PedidoPublico{}, apierr.ErrNoEncontrado
	}

	var detalles []domain.DetallePublico
	err := db.Table("detalle_pedido").
		Select("codigo, nombre, unidades, precio_unitario_centavos, subtotal_centavos").
		Where("pedido_id = (SELECT id FROM pedido WHERE token_acceso = ?)", token).
		Order("nombre").
		Scan(&detalles).Error
	if err != nil {
		return domain.PedidoPublico{}, err
	}

	return domain.PedidoPublico{
		Numero:           fila.Numero,
		ModoEntrega:      fila.ModoEntrega,
		Domicilio:        fila.Domicilio,
		DistanciaMetros:  fila.DistanciaMetros,
		EnvioCentavos:    fila.EnvioCentavos,
		SubtotalCentavos: fila.SubtotalCentavos,
		TotalCentavos:    fila.TotalCentavos,
		EstadoPedido:     fila.EstadoPedido,
		EstadoPago:       fila.EstadoPago,
		MedioPago:        fila.MedioPago,
		VenceEn:          fila.VenceEn,
		CreadoEn:         fila.CreadoEn,
		Detalles:         detalles,
	}, nil
}

// ListarDeUsuario devuelve los pedidos del usuario (más nuevos primero) y el
// total para paginar. pedido(usuario_id) tiene índice: sin él, cada pantalla
// recorrería la tabla entera.
func (r *RepositorioPedidos) ListarDeUsuario(ctx context.Context, usuarioID uuid.UUID, limit, offset int) ([]domain.Pedido, int64, error) {
	base := r.db.WithContext(ctx).Model(&domain.Pedido{}).Where("usuario_id = ?", usuarioID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var pedidos []domain.Pedido
	err := base.Order("creado_en DESC, numero DESC").Limit(limit).Offset(offset).Find(&pedidos).Error
	return pedidos, total, err
}

// ListarAdmin lista todos los pedidos con filtros opcionales por estado y
// fecha. Solo para endpoints con pedidos.gestionar.
func (r *RepositorioPedidos) ListarAdmin(ctx context.Context, f FiltrosAdmin, limit, offset int) ([]domain.Pedido, int64, error) {
	base := r.db.WithContext(ctx).Model(&domain.Pedido{})
	if f.Estado != nil {
		base = base.Where("estado_pedido = ?", *f.Estado)
	}
	if f.Desde != nil {
		base = base.Where("creado_en >= ?", *f.Desde)
	}
	if f.Hasta != nil {
		base = base.Where("creado_en < ?", *f.Hasta)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var pedidos []domain.Pedido
	err := base.Order("creado_en DESC, numero DESC").Limit(limit).Offset(offset).Find(&pedidos).Error
	return pedidos, total, err
}

// CambiarEstado pasa el pedido de `desde` a `nuevo`. El UPDATE incluye el
// estado actual en el WHERE: si otra request lo cambió entre que se leyó el
// pedido y se escribió, no se actualiza ninguna fila y se devuelve un 409. Así
// dos cancelaciones simultáneas no pueden ejecutarse las dos.
func (r *RepositorioPedidos) CambiarEstado(ctx context.Context, tx *gorm.DB, id uuid.UUID, desde, nuevo domain.EstadoPedido) error {
	res := tx.WithContext(ctx).Model(&domain.Pedido{}).
		Where("id = ? AND estado_pedido = ?", id, desde).
		Update("estado_pedido", nuevo)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apierr.ErrRegla{Mensaje: "El pedido cambió de estado mientras lo procesabas. Actualizá la pantalla e intentá de nuevo."}
	}
	return nil
}

// MarcarPagoAprobado registra el cobro (por ejemplo, el efectivo al retirar).
// Solo actúa si el pago estaba PENDIENTE, así registrar el cobro dos veces no
// hace nada la segunda.
func (r *RepositorioPedidos) MarcarPagoAprobado(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	res := tx.WithContext(ctx).Model(&domain.Pedido{}).
		Where("id = ? AND estado_pago = ?", id, domain.PagoPendiente).
		Update("estado_pago", domain.PagoAprobado)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apierr.ErrRegla{Mensaje: "El cobro de este pedido ya fue registrado."}
	}
	return nil
}

// RegistrarCancelacion guarda quién canceló, cuándo y por qué. El responsable
// sale de la sesión, nunca del body; la fecha entra por parámetro.
func (r *RepositorioPedidos) RegistrarCancelacion(ctx context.Context, tx *gorm.DB, id, responsableID uuid.UUID, motivo string, ahora time.Time) error {
	return tx.WithContext(ctx).Model(&domain.Pedido{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"cancelado_por":      responsableID,
			"cancelado_en":       ahora,
			"motivo_cancelacion": motivo,
		}).Error
}
