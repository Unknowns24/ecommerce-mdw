package pagos

import (
	"time"

	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/google/uuid"
)

type IncidenciaResponse struct {
	ID                  uuid.UUID  `json:"id"`
	ReferenciaExterna   string     `json:"referenciaExterna"`
	Motivo              string     `json:"motivo"`
	MontoCentavos       int64      `json:"montoCentavos"`
	TotalPedidoCentavos int64      `json:"totalPedidoCentavos"`
	CreadoEn            time.Time  `json:"creadoEn"`
	ResueltoPor         *uuid.UUID `json:"resueltoPor,omitempty"`
	ResueltoPorNombre   string     `json:"resueltoPorNombre,omitempty"`
	ResueltoEn          *time.Time `json:"resueltoEn,omitempty"`
	NotaResolucion      *string    `json:"notaResolucion,omitempty"`
}

func AIncidencias(items []domain.IncidenciaPago) []IncidenciaResponse {
	out := make([]IncidenciaResponse, 0, len(items))
	for _, i := range items {
		out = append(out, IncidenciaResponse{ID: i.ID, ReferenciaExterna: i.ReferenciaExterna, Motivo: i.Motivo,
			MontoCentavos: i.MontoCentavos, TotalPedidoCentavos: i.TotalPedidoCentavos, CreadoEn: i.CreadoEn,
			ResueltoPor: i.ResueltoPor, ResueltoPorNombre: i.ResueltoPorNombre, ResueltoEn: i.ResueltoEn, NotaResolucion: i.NotaResolucion})
	}
	return out
}
