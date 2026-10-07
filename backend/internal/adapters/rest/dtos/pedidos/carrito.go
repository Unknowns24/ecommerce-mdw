package pedidos

import (
	"fmt"

	"github.com/google/uuid"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

// AgregarItemRequest es el body de POST /api/carrito/items. No trae precio: el
// carrito siempre se valoriza con el catálogo del servidor.
type AgregarItemRequest struct {
	VarianteID string `json:"varianteId"`
	Unidades   int    `json:"unidades"`
}

// AEntrada valida la forma del body: un uuid y unidades enteras entre 1 y 1000.
// Que no superen el stock disponible es una regla de estado (409), no de forma.
func (r AgregarItemRequest) AEntrada() (uuid.UUID, int, error) {
	campos := map[string]string{}
	id, err := uuid.Parse(r.VarianteID)
	if err != nil {
		campos["varianteId"] = "Identificador de producto inválido"
	}
	if msg := validarUnidadesDTO(r.Unidades); msg != "" {
		campos["unidades"] = msg
	}
	if len(campos) > 0 {
		return uuid.Nil, 0, apierr.ErrValidacion{Campos: campos}
	}
	return id, r.Unidades, nil
}

// CambiarUnidadesRequest es el body de PATCH /api/carrito/items/{id}.
type CambiarUnidadesRequest struct {
	Unidades int `json:"unidades"`
}

func (r CambiarUnidadesRequest) AEntrada() (int, error) {
	if msg := validarUnidadesDTO(r.Unidades); msg != "" {
		return 0, apierr.ErrValidacion{Campos: map[string]string{"unidades": msg}}
	}
	return r.Unidades, nil
}

func validarUnidadesDTO(u int) string {
	if u < 1 || u > maxUnidades {
		return fmt.Sprintf("Las unidades deben ser un entero entre 1 y %d", maxUnidades)
	}
	return ""
}

// CarritoResponse es el carrito con su total estimado. "Estimado" porque el
// carrito no congela precios ni reserva unidades: el total definitivo lo fija
// el checkout.
type CarritoResponse struct {
	Items                 []LineaCarritoResponse `json:"items"`
	TotalEstimadoCentavos int64                  `json:"totalEstimadoCentavos"`
}

type LineaCarritoResponse struct {
	ID                     uuid.UUID `json:"id"`
	VarianteID             uuid.UUID `json:"varianteId"`
	Codigo                 string    `json:"codigo"`
	Nombre                 string    `json:"nombre"`
	Unidades               int       `json:"unidades"`
	PrecioUnitarioCentavos int64     `json:"precioUnitarioCentavos"`
	SubtotalCentavos       int64     `json:"subtotalCentavos"`
	Disponible             bool      `json:"disponible"`
}

func ACarrito(v uc.VistaCarrito) CarritoResponse {
	items := make([]LineaCarritoResponse, len(v.Lineas))
	for i, l := range v.Lineas {
		items[i] = LineaCarritoResponse{
			ID: l.ItemID, VarianteID: l.VarianteID, Codigo: l.Codigo, Nombre: l.Nombre, Unidades: l.Unidades,
			PrecioUnitarioCentavos: l.PrecioUnitarioCentavos, SubtotalCentavos: l.SubtotalCentavos, Disponible: l.Disponible,
		}
	}
	return CarritoResponse{Items: items, TotalEstimadoCentavos: v.TotalEstimadoCentavos}
}
