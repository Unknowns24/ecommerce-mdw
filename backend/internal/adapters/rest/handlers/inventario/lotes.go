package inventario

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/inventario"
	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/inventario"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

const endpointIngresoLote = "POST /api/admin/lotes"

func (h *handler) ingresoLote(w http.ResponseWriter, r *http.Request) {
	var body dto.IngresoLoteDTO
	if campos := decodificarYValidar(r, &body); campos != nil {
		apierr.Responder(w, endpointIngresoLote, apierr.ErrValidacion{Campos: campos})
		return
	}

	varianteID, errVariante := uuid.Parse(body.VarianteID)
	proveedorID, errProveedor := uuid.Parse(body.ProveedorID)
	if errVariante != nil || errProveedor != nil {
		apierr.Responder(w, endpointIngresoLote, apierr.ErrNoEncontrado)
		return
	}
	fecha, err := time.Parse(time.RFC3339, body.Fecha)
	if err != nil {
		apierr.Responder(w, endpointIngresoLote, apierr.ErrValidacion{Campos: map[string]string{
			"fecha": "Tiene que ser una fecha en formato RFC3339",
		}})
		return
	}

	ctx := r.Context()
	existeVariante, err := varianteExiste(ctx, h.db, varianteID)
	if err != nil {
		apierr.Responder(w, endpointIngresoLote, err)
		return
	}
	if !existeVariante {
		apierr.Responder(w, endpointIngresoLote, apierr.ErrNoEncontrado)
		return
	}
	if _, err := h.proveedores.PorID(ctx, proveedorID); err != nil {
		apierr.Responder(w, endpointIngresoLote, apierr.ErrNoEncontrado)
		return
	}

	lote := &domain.Lote{
		ID:                    uuid.New(),
		VarianteID:            varianteID,
		ProveedorID:           proveedorID,
		UnidadesIngresadas:    body.Unidades,
		CostoUnitarioCentavos: body.CostoUnitarioCentavos,
		FechaIngreso:          fecha,
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := repo.NuevoRepositorioLotes(tx).Crear(ctx, lote); err != nil {
			return err
		}
		movimiento := &domain.MovimientoStock{
			ID:       uuid.New(),
			LoteID:   lote.ID,
			Tipo:     domain.MovimientoIngreso,
			Unidades: body.Unidades,
			Motivo:   "Ingreso de lote",
		}
		return repo.NuevoRepositorioMovimientos(tx).Crear(ctx, movimiento)
	})
	if err != nil {
		apierr.Responder(w, endpointIngresoLote, err)
		return
	}

	apierr.JSON(w, http.StatusCreated, dto.LoteRespuesta{
		ID:                    lote.ID.String(),
		VarianteID:            lote.VarianteID.String(),
		ProveedorID:           lote.ProveedorID.String(),
		UnidadesIngresadas:    lote.UnidadesIngresadas,
		CostoUnitarioCentavos: lote.CostoUnitarioCentavos,
		FechaIngreso:          lote.FechaIngreso,
	})
}

// varianteExiste consulta la tabla de catálogo directamente: el lector
// congelado de Genaro (catalogo.VariantePorID) vive en su propio paquete de
// usecases y trae de vuelta un error de dominio en lugar de un booleano, así
// que para el único chequeo de existencia que este endpoint necesita alcanza
// con una consulta a la tabla.
func varianteExiste(ctx context.Context, db *gorm.DB, id uuid.UUID) (bool, error) {
	var existe bool
	err := db.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM variante WHERE id = ?)", id).Scan(&existe).Error
	return existe, err
}
