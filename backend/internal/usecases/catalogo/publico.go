package catalogo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/catalogo"
)

type Publico struct{ repo *repo.RepositorioPublico }

func NuevoPublico(db *gorm.DB) *Publico { return &Publico{repo: repo.NuevoRepositorioPublico(db)} }

type PaginaVariantes struct {
	Items     []repo.VariantePublica `json:"items"`
	Pagina    int                    `json:"pagina"`
	PorPagina int                    `json:"porPagina"`
	Total     int64                  `json:"total"`
}

func (s *Publico) Listar(ctx context.Context, f repo.Filtros, pagina, porPagina int) (PaginaVariantes, error) {
	items, total, err := s.repo.Listar(ctx, f, porPagina, (pagina-1)*porPagina)
	if err != nil {
		return PaginaVariantes{}, err
	}
	ids := make([]uuid.UUID, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	disponibles, err := s.repo.DisponibilidadPorVariantes(ctx, ids)
	if err != nil {
		return PaginaVariantes{}, err
	}
	for i := range items {
		items[i].Disponible = disponibles[items[i].ID]
	}
	return PaginaVariantes{Items: items, Pagina: pagina, PorPagina: porPagina, Total: total}, nil
}

type DetalleVariante struct {
	repo.VariantePublica
	Imagenes       []string                `json:"imagenes"`
	Categorias     []repo.CategoriaPublica `json:"categorias"`
	OtrasVariantes []repo.VariantePublica  `json:"otrasVariantes"`
}

func (s *Publico) Detalle(ctx context.Context, id uuid.UUID) (DetalleVariante, error) {
	item, err := s.repo.PorID(ctx, id)
	if err != nil {
		return DetalleVariante{}, err
	}
	imagenes, err := s.repo.Imagenes(ctx, id)
	if err != nil {
		return DetalleVariante{}, err
	}
	categorias, err := s.repo.CategoriasDeProducto(ctx, item.ProductoID)
	if err != nil {
		return DetalleVariante{}, err
	}
	otras, err := s.repo.Otras(ctx, item.ProductoID, id)
	if err != nil {
		return DetalleVariante{}, err
	}
	ids := []uuid.UUID{item.ID}
	for _, otra := range otras {
		ids = append(ids, otra.ID)
	}
	disponibles, err := s.repo.DisponibilidadPorVariantes(ctx, ids)
	if err != nil {
		return DetalleVariante{}, err
	}
	item.Disponible = disponibles[item.ID]
	for i := range otras {
		otras[i].Marca = item.Marca
		otras[i].Disponible = disponibles[otras[i].ID]
	}
	return DetalleVariante{VariantePublica: item, Imagenes: imagenes, Categorias: categorias, OtrasVariantes: otras}, nil
}

func (s *Publico) Marcas(ctx context.Context) ([]map[string]any, error) { return s.repo.Marcas(ctx) }
func (s *Publico) Categorias(ctx context.Context) ([]*repo.CategoriaPublica, error) {
	rows, err := s.repo.Categorias(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]*repo.CategoriaPublica, len(rows))
	for i := range rows {
		rows[i].Hijas = []*repo.CategoriaPublica{}
		byID[rows[i].ID] = &rows[i]
	}
	roots := make([]*repo.CategoriaPublica, 0)
	for i := range rows {
		c := &rows[i]
		if c.PadreID != nil {
			if padre, ok := byID[*c.PadreID]; ok {
				padre.Hijas = append(padre.Hijas, c)
				continue
			}
		}
		roots = append(roots, c)
	}
	return roots, nil
}
