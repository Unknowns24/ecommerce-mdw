package inventario

import (
	"net/http"

	"github.com/google/uuid"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/inventario"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

const endpointAltaProveedor = "POST /api/admin/proveedores"
const endpointListarProveedores = "GET /api/admin/proveedores"

func (h *handler) altaProveedor(w http.ResponseWriter, r *http.Request) {
	var body dto.AltaProveedorDTO
	if campos := decodificarYValidar(r, &body); campos != nil {
		apierr.Responder(w, endpointAltaProveedor, apierr.ErrValidacion{Campos: campos})
		return
	}

	proveedor := &domain.Proveedor{
		ID:       uuid.New(),
		Nombre:   body.Nombre,
		Correo:   body.Correo,
		Telefono: body.Telefono,
		Web:      body.Web,
	}
	if err := h.proveedores.Crear(r.Context(), proveedor); err != nil {
		apierr.Responder(w, endpointAltaProveedor, err)
		return
	}

	apierr.JSON(w, http.StatusCreated, proveedorARespuesta(*proveedor))
}

func (h *handler) listarProveedores(w http.ResponseWriter, r *http.Request) {
	limit, offset := paginacion(r)
	proveedores, err := h.proveedores.Listar(r.Context(), limit, offset)
	if err != nil {
		apierr.Responder(w, endpointListarProveedores, err)
		return
	}

	respuesta := make([]dto.ProveedorRespuesta, len(proveedores))
	for i, p := range proveedores {
		respuesta[i] = proveedorARespuesta(p)
	}
	apierr.JSON(w, http.StatusOK, map[string]any{"items": respuesta})
}

func proveedorARespuesta(p domain.Proveedor) dto.ProveedorRespuesta {
	return dto.ProveedorRespuesta{
		ID:       p.ID.String(),
		Nombre:   p.Nombre,
		Correo:   p.Correo,
		Telefono: p.Telefono,
		Web:      p.Web,
		CreadoEn: p.CreadoEn,
	}
}
