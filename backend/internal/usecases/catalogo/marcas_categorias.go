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
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type ServicioClasificacion struct{ repo *repo.MarcasCategorias }

func NuevoServicioClasificacion(db *gorm.DB) *ServicioClasificacion {
	return &ServicioClasificacion{repo: repo.NuevoMarcasCategorias(db)}
}

func validarNombre(nombre string) error {
	if len(strings.TrimSpace(nombre)) == 0 || len(nombre) > 150 {
		return apierr.ErrValidacion{Campos: map[string]string{"nombre": "entre 1 y 150 caracteres"}}
	}
	return nil
}

func (s *ServicioClasificacion) CrearMarca(ctx context.Context, nombre string) (domain.Marca, error) {
	if err := validarNombre(nombre); err != nil {
		return domain.Marca{}, err
	}
	m := domain.Marca{ID: uuid.New(), Nombre: strings.TrimSpace(nombre), Estado: "ACTIVA", CreadoEn: time.Now().UTC()}
	if err := s.repo.CrearMarca(ctx, &m); err != nil {
		return domain.Marca{}, conflictoNombreMarca(err)
	}
	return m, nil
}
func (s *ServicioClasificacion) Marca(ctx context.Context, id uuid.UUID) (domain.Marca, error) {
	return s.repo.Marca(ctx, id)
}
func (s *ServicioClasificacion) ListarMarcas(ctx context.Context, pagina, porPagina int) (map[string]any, error) {
	items, total, err := s.repo.ListarMarcas(ctx, porPagina, (pagina-1)*porPagina)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "pagina": pagina, "porPagina": porPagina, "total": total}, nil
}
func (s *ServicioClasificacion) EditarMarca(ctx context.Context, id uuid.UUID, nombre *string) (domain.Marca, error) {
	if nombre == nil {
		return s.repo.Marca(ctx, id)
	}
	if err := validarNombre(*nombre); err != nil {
		return domain.Marca{}, err
	}
	m, err := s.repo.EditarMarca(ctx, id, map[string]any{"nombre": strings.TrimSpace(*nombre)})
	return m, conflictoNombreMarca(err)
}
func (s *ServicioClasificacion) DesactivarMarca(ctx context.Context, id uuid.UUID) error {
	return s.repo.DesactivarMarca(ctx, id)
}

func (s *ServicioClasificacion) CrearCategoria(ctx context.Context, nombre string, padreID *uuid.UUID) (domain.Categoria, error) {
	if err := validarNombre(nombre); err != nil {
		return domain.Categoria{}, err
	}
	if err := s.validarPadre(ctx, uuid.Nil, padreID); err != nil {
		return domain.Categoria{}, err
	}
	c := domain.Categoria{ID: uuid.New(), Nombre: strings.TrimSpace(nombre), PadreID: padreID, Estado: "ACTIVA", CreadoEn: time.Now().UTC()}
	if err := s.repo.CrearCategoria(ctx, &c); err != nil {
		return domain.Categoria{}, err
	}
	return c, nil
}
func (s *ServicioClasificacion) Categoria(ctx context.Context, id uuid.UUID) (domain.Categoria, error) {
	return s.repo.Categoria(ctx, id)
}
func (s *ServicioClasificacion) ListarCategorias(ctx context.Context, pagina, porPagina int) (map[string]any, error) {
	items, total, err := s.repo.ListarCategorias(ctx, porPagina, (pagina-1)*porPagina)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "pagina": pagina, "porPagina": porPagina, "total": total}, nil
}
func (s *ServicioClasificacion) EditarCategoria(ctx context.Context, id uuid.UUID, nombre *string, padreID **uuid.UUID) (domain.Categoria, error) {
	changes := map[string]any{}
	if nombre != nil {
		if err := validarNombre(*nombre); err != nil {
			return domain.Categoria{}, err
		}
		changes["nombre"] = strings.TrimSpace(*nombre)
	}
	if padreID != nil {
		if err := s.validarPadre(ctx, id, *padreID); err != nil {
			return domain.Categoria{}, err
		}
		changes["padre_id"] = *padreID
	}
	if len(changes) == 0 {
		return s.repo.Categoria(ctx, id)
	}
	return s.repo.EditarCategoria(ctx, id, changes)
}
func (s *ServicioClasificacion) DesactivarCategoria(ctx context.Context, id uuid.UUID) error {
	return s.repo.DesactivarCategoria(ctx, id)
}

func (s *ServicioClasificacion) validarPadre(ctx context.Context, id uuid.UUID, padreID *uuid.UUID) error {
	if padreID == nil {
		return nil
	}
	if *padreID == uuid.Nil || *padreID == id {
		return apierr.ErrValidacion{Campos: map[string]string{"padreId": "padre inválido"}}
	}
	padre, err := s.repo.Categoria(ctx, *padreID)
	if errors.Is(err, apierr.ErrNoEncontrado) {
		return apierr.ErrValidacion{Campos: map[string]string{"padreId": "categoría inexistente"}}
	}
	if err != nil {
		return err
	}
	if padre.Estado != "ACTIVA" {
		return apierr.ErrValidacion{Campos: map[string]string{"padreId": "categoría inactiva"}}
	}
	if id != uuid.Nil {
		cycle, err := s.repo.EsDescendiente(ctx, id, *padreID)
		if err != nil {
			return err
		}
		if cycle {
			return apierr.ErrValidacion{Campos: map[string]string{"padreId": "crea un ciclo"}}
		}
	}
	return nil
}

func conflictoNombreMarca(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apierr.ErrRegla{Mensaje: "La marca ya existe"}
	}
	return err
}
