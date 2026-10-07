package inventario

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/inventario"
	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/inventario"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

const endpointStockVariante = "GET /api/admin/variantes/{id}/stock"
const endpointMovimientosVariante = "GET /api/admin/variantes/{id}/movimientos"
const endpointAjusteStock = "POST /api/admin/variantes/{id}/ajustes"

func (h *handler) stockVariante(w http.ResponseWriter, r *http.Request) {
	varianteID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierr.Responder(w, endpointStockVariante, apierr.ErrNoEncontrado)
		return
	}

	ctx := r.Context()
	existe, err := varianteExiste(ctx, h.db, varianteID)
	if err != nil {
		apierr.Responder(w, endpointStockVariante, err)
		return
	}
	if !existe {
		apierr.Responder(w, endpointStockVariante, apierr.ErrNoEncontrado)
		return
	}

	disponible, err := h.stock.Disponible(ctx, varianteID)
	if err != nil {
		apierr.Responder(w, endpointStockVariante, err)
		return
	}
	lotes, err := h.lotes.PorVariante(ctx, varianteID, "FIFO")
	if err != nil {
		apierr.Responder(w, endpointStockVariante, err)
		return
	}

	respuesta := dto.StockVarianteRespuesta{
		VarianteID: varianteID.String(),
		Disponible: disponible,
		Lotes:      make([]dto.LoteRespuesta, len(lotes)),
	}
	for i, l := range lotes {
		respuesta.Lotes[i] = dto.LoteRespuesta{
			ID:                    l.ID.String(),
			VarianteID:            l.VarianteID.String(),
			ProveedorID:           l.ProveedorID.String(),
			UnidadesIngresadas:    l.UnidadesIngresadas,
			CostoUnitarioCentavos: l.CostoUnitarioCentavos,
			FechaIngreso:          l.FechaIngreso,
		}
	}
	apierr.JSON(w, http.StatusOK, respuesta)
}

func (h *handler) movimientosVariante(w http.ResponseWriter, r *http.Request) {
	varianteID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierr.Responder(w, endpointMovimientosVariante, apierr.ErrNoEncontrado)
		return
	}

	ctx := r.Context()
	existe, err := varianteExiste(ctx, h.db, varianteID)
	if err != nil {
		apierr.Responder(w, endpointMovimientosVariante, err)
		return
	}
	if !existe {
		apierr.Responder(w, endpointMovimientosVariante, apierr.ErrNoEncontrado)
		return
	}

	limit, offset := paginacion(r)
	movimientos, err := h.movimientos.PorVariante(ctx, varianteID, limit, offset)
	if err != nil {
		apierr.Responder(w, endpointMovimientosVariante, err)
		return
	}

	respuesta := make([]dto.MovimientoRespuesta, len(movimientos))
	for i, m := range movimientos {
		respuesta[i] = dto.MovimientoRespuesta{
			ID:       m.ID.String(),
			LoteID:   m.LoteID.String(),
			Tipo:     string(m.Tipo),
			Unidades: m.Unidades,
			Motivo:   m.Motivo,
			CreadoEn: m.CreadoEn,
		}
	}
	apierr.JSON(w, http.StatusOK, map[string]any{"items": respuesta})
}

func (h *handler) ajusteStock(w http.ResponseWriter, r *http.Request) {
	varianteID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierr.Responder(w, endpointAjusteStock, apierr.ErrNoEncontrado)
		return
	}

	var body dto.AjusteStockDTO
	if campos := decodificarYValidar(r, &body); campos != nil {
		apierr.Responder(w, endpointAjusteStock, apierr.ErrValidacion{Campos: campos})
		return
	}
	loteID, err := uuid.Parse(body.LoteID)
	if err != nil {
		apierr.Responder(w, endpointAjusteStock, apierr.ErrValidacion{Campos: map[string]string{
			"loteId": "Tiene que ser un identificador válido",
		}})
		return
	}

	ctx := r.Context()
	usuario, ok := middleware.UsuarioDeContexto(ctx)
	if !ok {
		apierr.Responder(w, endpointAjusteStock, apierr.ErrNoAutenticado)
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		lotesTx := repo.NuevoRepositorioLotes(tx)
		lote, err := lotesTx.PorIDBloqueando(ctx, loteID)
		if err != nil {
			return apierr.ErrNoEncontrado
		}
		if lote.VarianteID != varianteID {
			return apierr.ErrNoEncontrado
		}

		neto, err := repo.NuevoRepositorioMovimientos(tx).NetoPorLotes(ctx, []uuid.UUID{loteID})
		if err != nil {
			return err
		}
		activas, err := repo.NuevoRepositorioReservas(tx).ActivasPorLotes(ctx, []uuid.UUID{loteID})
		if err != nil {
			return err
		}

		actual := lote.UnidadesIngresadas + neto[loteID]
		resultante := actual + body.Unidades
		if resultante < 0 {
			return apierr.ErrRegla{Mensaje: "El ajuste dejaría stock negativo"}
		}
		if resultante < activas[loteID] {
			return apierr.ErrRegla{Mensaje: "El ajuste consumiría unidades reservadas en otro pedido"}
		}

		responsableID := usuario.ID
		movimiento := &domain.MovimientoStock{
			ID:            uuid.New(),
			LoteID:        loteID,
			Tipo:          domain.MovimientoAjuste,
			Unidades:      body.Unidades,
			Motivo:        body.Motivo,
			ResponsableID: &responsableID,
		}
		return repo.NuevoRepositorioMovimientos(tx).Crear(ctx, movimiento)
	})
	if err != nil {
		apierr.Responder(w, endpointAjusteStock, err)
		return
	}

	apierr.JSON(w, http.StatusCreated, map[string]string{"estado": "ajuste registrado"})
}
