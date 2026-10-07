package catalogo

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

type Filtros struct {
	Texto       string
	CategoriaID *uuid.UUID
	MarcaID     *uuid.UUID
	Orden       string
}

type VariantePublica struct {
	ID                      uuid.UUID `json:"id"`
	ProductoID              uuid.UUID `json:"-"`
	Codigo                  string    `json:"codigo"`
	Nombre                  string    `json:"nombre"`
	Descripcion             string    `json:"descripcion,omitempty"`
	Marca                   string    `json:"marca"`
	PrecioMinoristaCentavos int64     `json:"precioMinoristaCentavos"`
	PrecioMayoristaCentavos int64     `json:"precioMayoristaCentavos,omitempty"`
	ImagenPrincipal         string    `json:"imagenPrincipal"`
	Disponible              bool      `json:"disponible" gorm:"-"`
}

type CategoriaPublica struct {
	ID      uuid.UUID           `json:"id"`
	Nombre  string              `json:"nombre"`
	PadreID *uuid.UUID          `json:"-"`
	Hijas   []*CategoriaPublica `json:"hijas,omitempty" gorm:"-"`
}

type RepositorioPublico struct{ db *gorm.DB }

func NuevoRepositorioPublico(db *gorm.DB) *RepositorioPublico { return &RepositorioPublico{db: db} }

func (r *RepositorioPublico) base(ctx context.Context, f Filtros) *gorm.DB {
	q := r.db.WithContext(ctx).Table("variante AS v").Joins("JOIN producto AS p ON p.id = v.producto_id").Joins("JOIN marca AS m ON m.id = p.marca_id").Where("v.estado = ? AND p.estado = ? AND m.estado = ?", "ACTIVA", "ACTIVO", "ACTIVA")
	if f.Texto != "" {
		literal := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.ToLower(f.Texto))
		q = q.Where(`lower(v.nombre) LIKE ? ESCAPE '\'`, "%"+literal+"%")
	}
	if f.MarcaID != nil {
		q = q.Where("p.marca_id = ?", *f.MarcaID)
	}
	if f.CategoriaID != nil {
		q = q.Where(`EXISTS (WITH RECURSIVE descendientes AS (SELECT id FROM categoria WHERE id = ? UNION ALL SELECT c.id FROM categoria c JOIN descendientes d ON c.padre_id = d.id) SELECT 1 FROM producto_categoria pc JOIN descendientes d ON d.id = pc.categoria_id WHERE pc.producto_id = p.id)`, *f.CategoriaID)
	}
	return q
}

func (r *RepositorioPublico) Listar(ctx context.Context, f Filtros, limit, offset int) ([]VariantePublica, int64, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, 0, apierr.ErrValidacion{Campos: map[string]string{"paginacion": "fuera de rango"}}
	}
	var total int64
	if err := r.base(ctx, f).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	ordenes := map[string]string{"nombre_asc": "v.nombre ASC, v.id ASC", "precio_asc": "v.precio_minorista_centavos ASC, v.id ASC", "precio_desc": "v.precio_minorista_centavos DESC, v.id ASC"}
	orden, ok := ordenes[f.Orden]
	if !ok {
		return nil, 0, apierr.ErrValidacion{Campos: map[string]string{"orden": "inválido"}}
	}
	items := make([]VariantePublica, 0)
	err := r.base(ctx, f).Select("v.id, v.producto_id, v.codigo, v.nombre, v.precio_minorista_centavos, m.nombre AS marca, COALESCE(i.referencia_imagen, '') AS imagen_principal").Joins("LEFT JOIN imagen_variante AS i ON i.variante_id = v.id AND i.es_principal = true").Order(orden).Limit(limit).Offset(offset).Scan(&items).Error
	return items, total, err
}

func (r *RepositorioPublico) PorID(ctx context.Context, id uuid.UUID) (VariantePublica, error) {
	var item VariantePublica
	err := r.base(ctx, Filtros{}).Select("v.id, v.producto_id, v.codigo, v.nombre, v.descripcion, v.precio_minorista_centavos, v.precio_mayorista_centavos, m.nombre AS marca, COALESCE(i.referencia_imagen, '') AS imagen_principal").Joins("LEFT JOIN imagen_variante AS i ON i.variante_id = v.id AND i.es_principal = true").Where("v.id = ?", id).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, apierr.ErrNoEncontrado
	}
	return item, err
}

func (r *RepositorioPublico) Imagenes(ctx context.Context, id uuid.UUID) ([]string, error) {
	var rows []struct{ ReferenciaImagen string }
	err := r.db.WithContext(ctx).Table("imagen_variante").Select("referencia_imagen").Where("variante_id = ?", id).Order("orden ASC, id ASC").Limit(100).Scan(&rows).Error
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ReferenciaImagen)
	}
	return out, err
}

func (r *RepositorioPublico) CategoriasDeProducto(ctx context.Context, productoID uuid.UUID) ([]CategoriaPublica, error) {
	var out []CategoriaPublica
	err := r.db.WithContext(ctx).Table("categoria AS c").Select("c.id, c.nombre, c.padre_id").Joins("JOIN producto_categoria AS pc ON pc.categoria_id = c.id").Where("pc.producto_id = ? AND c.estado = ?", productoID, "ACTIVA").Order("c.nombre ASC, c.id ASC").Limit(100).Scan(&out).Error
	if out == nil {
		out = []CategoriaPublica{}
	}
	return out, err
}

func (r *RepositorioPublico) Otras(ctx context.Context, productoID, actualID uuid.UUID) ([]VariantePublica, error) {
	var out []VariantePublica
	err := r.db.WithContext(ctx).Table("variante AS v").Select("v.id, v.producto_id, v.codigo, v.nombre, v.precio_minorista_centavos, COALESCE(i.referencia_imagen, '') AS imagen_principal").Joins("LEFT JOIN imagen_variante AS i ON i.variante_id = v.id AND i.es_principal = true").Where("v.producto_id = ? AND v.id <> ? AND v.estado = ?", productoID, actualID, "ACTIVA").Order("v.nombre ASC, v.id ASC").Limit(100).Scan(&out).Error
	return out, err
}

func (r *RepositorioPublico) Marcas(ctx context.Context) ([]map[string]any, error) {
	var rows []struct {
		ID     uuid.UUID
		Nombre string
	}
	err := r.db.WithContext(ctx).Table("marca").Select("id, nombre").Where("estado = ?", "ACTIVA").Order("nombre ASC, id ASC").Limit(1000).Scan(&rows).Error
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{"id": row.ID, "nombre": row.Nombre})
	}
	return out, err
}

func (r *RepositorioPublico) Categorias(ctx context.Context) ([]CategoriaPublica, error) {
	var rows []CategoriaPublica
	err := r.db.WithContext(ctx).Table("categoria").Select("id, nombre, padre_id").Where("estado = ?", "ACTIVA").Order("nombre ASC, id ASC").Limit(1000).Scan(&rows).Error
	return rows, err
}

// DisponibilidadPorVariantes calcula existencias menos salidas y reservas en una consulta.
func (r *RepositorioPublico) DisponibilidadPorVariantes(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := make(map[uuid.UUID]bool, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		VarianteID uuid.UUID
		Unidades   int64
	}
	err := r.db.WithContext(ctx).Raw(`WITH seleccion AS (
		SELECT id, variante_id, unidades_ingresadas FROM lote WHERE variante_id IN ?
	), movimientos AS (
		SELECT m.lote_id, SUM(CASE WHEN m.tipo = 'AJUSTE' THEN m.unidades WHEN m.tipo = 'SALIDA' THEN -m.unidades ELSE 0 END) AS neto
		FROM movimiento_stock m JOIN seleccion l ON l.id = m.lote_id GROUP BY m.lote_id
	), reservas AS (
		SELECT rs.lote_id, SUM(rs.unidades) AS reservadas
		FROM reserva_stock rs JOIN seleccion l ON l.id = rs.lote_id WHERE rs.estado = 'ACTIVA' GROUP BY rs.lote_id
	)
	SELECT l.variante_id, SUM(l.unidades_ingresadas + COALESCE(m.neto, 0) - COALESCE(rs.reservadas, 0)) AS unidades
	FROM seleccion l LEFT JOIN movimientos m ON m.lote_id = l.id
	LEFT JOIN reservas rs ON rs.lote_id = l.id GROUP BY l.variante_id`, ids).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.VarianteID] = row.Unidades > 0
	}
	return out, nil
}
