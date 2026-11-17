package pagos

import (
	"context"
	"errors"
	"time"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RepositorioIncidencias struct{ db *gorm.DB }

func NuevoRepositorioIncidencias(db *gorm.DB) *RepositorioIncidencias {
	return &RepositorioIncidencias{db: db}
}

func (r *RepositorioIncidencias) Registrar(ctx context.Context, incidencia *domain.IncidenciaPago) error {
	return r.db.WithContext(ctx).Omit("ResueltoPorNombre").Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "pedido_id"}, {Name: "referencia_externa"}, {Name: "motivo"}}, DoNothing: true,
	}).Create(incidencia).Error
}

func (r *RepositorioIncidencias) Listar(ctx context.Context, pedidoID uuid.UUID) ([]domain.IncidenciaPago, error) {
	var existe bool
	if err := r.db.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM pedido WHERE id=?)", pedidoID).Scan(&existe).Error; err != nil {
		return nil, err
	}
	if !existe {
		return nil, apierr.ErrNoEncontrado
	}
	items := make([]domain.IncidenciaPago, 0)
	err := r.db.WithContext(ctx).Table("incidencia_pago i").
		Select("i.id, i.pedido_id, i.referencia_externa, i.motivo, i.monto_centavos, i.total_pedido_centavos, i.creado_en, i.resuelto_por, i.resuelto_en, i.nota_resolucion, COALESCE(u.nombre || ' ' || u.apellido, '') AS resuelto_por_nombre").
		Joins("LEFT JOIN usuario u ON u.id=i.resuelto_por").Where("i.pedido_id=?", pedidoID).
		Order("(i.resuelto_en IS NOT NULL), i.creado_en DESC, i.id").Limit(100).Scan(&items).Error
	return items, err
}

// El WHERE incluye el pedido de la URL y no permite sobrescribir una resolución.
func (r *RepositorioIncidencias) Resolver(ctx context.Context, pedidoID, id, actor uuid.UUID, nota string, ahora time.Time) error {
	res := r.db.WithContext(ctx).Model(&domain.IncidenciaPago{}).
		Where("id=? AND pedido_id=? AND resuelto_en IS NULL", id, pedidoID).
		Updates(map[string]any{"resuelto_por": actor, "resuelto_en": ahora, "nota_resolucion": nota})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}
	var existente domain.IncidenciaPago
	err := r.db.WithContext(ctx).Select("id").Where("id=? AND pedido_id=?", id, pedidoID).Take(&existente).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apierr.ErrNoEncontrado
	}
	return err
}

func (r *RepositorioIncidencias) PendientesVencidos(ctx context.Context, ahora time.Time, limite int) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Table("pedido").Select("id").
		Where("medio_pago='MERCADO_PAGO' AND estado_pedido='PENDIENTE_DE_PAGO' AND vence_en <= ?", ahora).
		Order("COALESCE(pago_revisado_en, '-infinity'::timestamptz), vence_en, id").Limit(limite).Pluck("id", &ids).Error
	return ids, err
}

func (r *RepositorioIncidencias) RegistrarRevision(ctx context.Context, pedidoID uuid.UUID, ahora time.Time) error {
	return r.db.WithContext(ctx).Table("pedido").Where("id=?", pedidoID).Update("pago_revisado_en", ahora).Error
}
