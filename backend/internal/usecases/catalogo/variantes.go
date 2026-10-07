package catalogo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/catalogo"
	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/catalogo"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type ServicioVariantes struct{ repo *repo.Variantes }

func NuevoServicioVariantes(db *gorm.DB) *ServicioVariantes {
	return &ServicioVariantes{repo: repo.NuevoVariantes(db)}
}

func (s *ServicioVariantes) Crear(ctx context.Context, in dto.CrearVariante) (domain.Variante, error) {
	if err := validarVariante(in.Codigo, in.Nombre, in.PrecioMinoristaCentavos, in.PrecioMayoristaCentavos); err != nil {
		return domain.Variante{}, err
	}
	ok, err := s.repo.ProductoActivo(ctx, in.ProductoID)
	if err != nil {
		return domain.Variante{}, err
	}
	if !ok {
		return domain.Variante{}, apierr.ErrValidacion{Campos: map[string]string{"productoId": "producto activo inexistente"}}
	}
	now := time.Now().UTC()
	v := domain.Variante{ID: uuid.New(), ProductoID: in.ProductoID, Codigo: strings.TrimSpace(in.Codigo), Nombre: strings.TrimSpace(in.Nombre), Descripcion: in.Descripcion, PrecioMinoristaCentavos: in.PrecioMinoristaCentavos, PrecioMayoristaCentavos: in.PrecioMayoristaCentavos, Estado: "ACTIVA", CreadoEn: now, ActualizadoEn: now}
	if err := s.repo.Crear(ctx, &v); err != nil {
		return domain.Variante{}, conflictoCodigo(err)
	}
	return v, nil
}

func (s *ServicioVariantes) PorID(ctx context.Context, id uuid.UUID) (domain.Variante, error) {
	return s.repo.PorID(ctx, id)
}

func (s *ServicioVariantes) Listar(ctx context.Context, pagina, porPagina int) (map[string]any, error) {
	items, total, err := s.repo.Listar(ctx, porPagina, (pagina-1)*porPagina)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "pagina": pagina, "porPagina": porPagina, "total": total}, nil
}

func (s *ServicioVariantes) Editar(ctx context.Context, id uuid.UUID, in dto.EditarVariante) (domain.Variante, error) {
	v, err := s.repo.PorID(ctx, id)
	if err != nil {
		return v, err
	}
	changes := map[string]any{}
	if in.Codigo != nil {
		v.Codigo = strings.TrimSpace(*in.Codigo)
		changes["codigo"] = v.Codigo
	}
	if in.Nombre != nil {
		v.Nombre = strings.TrimSpace(*in.Nombre)
		changes["nombre"] = v.Nombre
	}
	if in.Descripcion != nil {
		changes["descripcion"] = *in.Descripcion
	}
	if in.PrecioMinoristaCentavos != nil {
		v.PrecioMinoristaCentavos = *in.PrecioMinoristaCentavos
		changes["precio_minorista_centavos"] = v.PrecioMinoristaCentavos
	}
	if in.PrecioMayoristaCentavos != nil {
		v.PrecioMayoristaCentavos = *in.PrecioMayoristaCentavos
		changes["precio_mayorista_centavos"] = v.PrecioMayoristaCentavos
	}
	if err := validarVariante(v.Codigo, v.Nombre, v.PrecioMinoristaCentavos, v.PrecioMayoristaCentavos); err != nil {
		return domain.Variante{}, err
	}
	changes["actualizado_en"] = time.Now().UTC()
	v, err = s.repo.Actualizar(ctx, id, changes)
	return v, conflictoCodigo(err)
}

func (s *ServicioVariantes) Desactivar(ctx context.Context, id uuid.UUID) error {
	return s.repo.CambiarEstado(ctx, id, "INACTIVA")
}

func validarVariante(codigo, nombre string, minorista, mayorista int64) error {
	c := map[string]string{}
	if len(strings.TrimSpace(codigo)) == 0 || len(codigo) > 100 {
		c["codigo"] = "entre 1 y 100 caracteres"
	}
	if len(strings.TrimSpace(nombre)) == 0 || len(nombre) > 255 {
		c["nombre"] = "entre 1 y 255 caracteres"
	}
	if minorista < 0 {
		c["precioMinoristaCentavos"] = "no negativo"
	}
	if mayorista < 0 {
		c["precioMayoristaCentavos"] = "no negativo"
	}
	if len(c) > 0 {
		return apierr.ErrValidacion{Campos: c}
	}
	return nil
}

func conflictoCodigo(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apierr.ErrRegla{Mensaje: "El código de variante ya existe"}
	}
	return err
}
