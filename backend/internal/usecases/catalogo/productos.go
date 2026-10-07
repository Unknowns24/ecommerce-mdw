package catalogo

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/catalogo"
	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/catalogo"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

type ServicioProductos struct{ repo *repo.Productos }

func NuevoServicioProductos(db *gorm.DB) *ServicioProductos {
	return &ServicioProductos{repo: repo.NuevoProductos(db)}
}

type DetalleProducto struct {
	domain.Producto
	Categorias []uuid.UUID       `json:"categorias"`
	Variantes  []domain.Variante `json:"variantes"`
}

func (s *ServicioProductos) Crear(ctx context.Context, in dto.CrearProducto) (DetalleProducto, error) {
	if err := validarProducto(in.MarcaID, in.Nombre, in.MetodologiaRotacion, in.Categorias); err != nil {
		return DetalleProducto{}, err
	}
	if err := s.validarReferencias(ctx, in.MarcaID, in.Categorias); err != nil {
		return DetalleProducto{}, err
	}
	now := time.Now().UTC()
	p := domain.Producto{ID: uuid.New(), MarcaID: in.MarcaID, Nombre: strings.TrimSpace(in.Nombre), Descripcion: in.Descripcion, Estado: "ACTIVO", MetodologiaRotacion: in.MetodologiaRotacion, CreadoEn: now, ActualizadoEn: now}
	if err := s.repo.Crear(ctx, &p, in.Categorias); err != nil {
		return DetalleProducto{}, err
	}
	return s.PorID(ctx, p.ID)
}

func (s *ServicioProductos) PorID(ctx context.Context, id uuid.UUID) (DetalleProducto, error) {
	p, err := s.repo.PorID(ctx, id)
	if err != nil {
		return DetalleProducto{}, err
	}
	cats, err := s.repo.Categorias(ctx, id)
	if err != nil {
		return DetalleProducto{}, err
	}
	vars, err := s.repo.Variantes(ctx, id)
	if err != nil {
		return DetalleProducto{}, err
	}
	return DetalleProducto{Producto: p, Categorias: cats, Variantes: vars}, nil
}

func (s *ServicioProductos) Listar(ctx context.Context, pagina, porPagina int) (map[string]any, error) {
	items, total, err := s.repo.Listar(ctx, porPagina, (pagina-1)*porPagina)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "pagina": pagina, "porPagina": porPagina, "total": total}, nil
}

func (s *ServicioProductos) Editar(ctx context.Context, id uuid.UUID, in dto.EditarProducto) (DetalleProducto, error) {
	p, err := s.repo.PorID(ctx, id)
	if err != nil {
		return DetalleProducto{}, err
	}
	changes := map[string]any{}
	if in.MarcaID != nil {
		p.MarcaID = *in.MarcaID
		changes["marca_id"] = p.MarcaID
	}
	if in.Nombre != nil {
		p.Nombre = strings.TrimSpace(*in.Nombre)
		changes["nombre"] = p.Nombre
	}
	if in.Descripcion != nil {
		changes["descripcion"] = *in.Descripcion
	}
	if in.MetodologiaRotacion != nil {
		p.MetodologiaRotacion = *in.MetodologiaRotacion
		changes["metodologia_rotacion"] = p.MetodologiaRotacion
	}
	cats := []uuid.UUID{}
	if in.Categorias != nil {
		cats = *in.Categorias
	} else {
		cats, err = s.repo.Categorias(ctx, id)
		if err != nil {
			return DetalleProducto{}, err
		}
	}
	if err := validarProducto(p.MarcaID, p.Nombre, p.MetodologiaRotacion, cats); err != nil {
		return DetalleProducto{}, err
	}
	if err := s.validarReferencias(ctx, p.MarcaID, cats); err != nil {
		return DetalleProducto{}, err
	}
	changes["actualizado_en"] = time.Now().UTC()
	if _, err := s.repo.Actualizar(ctx, id, changes, in.Categorias); err != nil {
		return DetalleProducto{}, err
	}
	return s.PorID(ctx, id)
}

func (s *ServicioProductos) Desactivar(ctx context.Context, id uuid.UUID) error {
	return s.repo.CambiarEstado(ctx, id, "INACTIVO")
}

func (s *ServicioProductos) validarReferencias(ctx context.Context, marcaID uuid.UUID, categorias []uuid.UUID) error {
	ok, err := s.repo.MarcaActiva(ctx, marcaID)
	if err != nil {
		return err
	}
	if !ok {
		return apierr.ErrValidacion{Campos: map[string]string{"marcaId": "marca activa inexistente"}}
	}
	ok, err = s.repo.CategoriasActivas(ctx, categorias)
	if err != nil {
		return err
	}
	if !ok {
		return apierr.ErrValidacion{Campos: map[string]string{"categorias": "categoría activa inexistente"}}
	}
	return nil
}

func validarProducto(marcaID uuid.UUID, nombre, rotacion string, categorias []uuid.UUID) error {
	c := map[string]string{}
	if marcaID == uuid.Nil {
		c["marcaId"] = "requerido"
	}
	if len(strings.TrimSpace(nombre)) == 0 || len(nombre) > 255 {
		c["nombre"] = "entre 1 y 255 caracteres"
	}
	if rotacion != "FIFO" && rotacion != "LIFO" {
		c["metodologiaRotacion"] = "FIFO o LIFO"
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range categorias {
		if id == uuid.Nil || seen[id] {
			c["categorias"] = "ids únicos y válidos"
		}
		seen[id] = true
	}
	if len(c) > 0 {
		return apierr.ErrValidacion{Campos: c}
	}
	return nil
}
