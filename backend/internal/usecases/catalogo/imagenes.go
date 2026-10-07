package catalogo

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/catalogo"
	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/catalogo"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type ServicioImagenes struct {
	imagenes  *repo.Imagenes
	variantes *repo.Variantes
}

func NuevoServicioImagenes(db *gorm.DB) *ServicioImagenes {
	return &ServicioImagenes{imagenes: repo.NuevoImagenes(db), variantes: repo.NuevoVariantes(db)}
}

func (s *ServicioImagenes) Crear(ctx context.Context, varianteID uuid.UUID, in dto.CrearImagen) (domain.ImagenVariante, error) {
	if _, err := s.variantes.PorID(ctx, varianteID); err != nil {
		return domain.ImagenVariante{}, err
	}
	if len(strings.TrimSpace(in.ReferenciaImagen)) == 0 || len(in.ReferenciaImagen) > 2048 {
		return domain.ImagenVariante{}, apierr.ErrValidacion{Campos: map[string]string{"referenciaImagen": "entre 1 y 2048 caracteres"}}
	}
	if in.Orden < 0 {
		return domain.ImagenVariante{}, apierr.ErrValidacion{Campos: map[string]string{"orden": "no negativo"}}
	}
	imagen := domain.ImagenVariante{ID: uuid.New(), VarianteID: varianteID, ReferenciaImagen: strings.TrimSpace(in.ReferenciaImagen), Orden: in.Orden}
	if err := s.imagenes.Crear(ctx, &imagen); err != nil {
		return domain.ImagenVariante{}, err
	}
	return imagen, nil
}

func (s *ServicioImagenes) Eliminar(ctx context.Context, id uuid.UUID) error {
	return s.imagenes.Eliminar(ctx, id)
}
func (s *ServicioImagenes) DefinirPrincipal(ctx context.Context, varianteID, imagenID uuid.UUID) error {
	return s.imagenes.DefinirPrincipal(ctx, varianteID, imagenID)
}
