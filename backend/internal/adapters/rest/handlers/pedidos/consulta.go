package pedidos

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/pedidos"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
)

// longitudToken es el largo de un token_acceso válido: 32 bytes en base64 URL
// sin relleno son 43 caracteres. Cualquier otra cosa no puede ser un token.
const longitudToken = 43

// ListarMisPedidos atiende GET /api/pedidos (SPEC-H16): los pedidos del
// usuario de la sesión, paginados con ?pagina=1&tamano=20. El usuario sale del
// contexto de la sesión: el listado no acepta un usuarioId por la query.
func (h *Handler) ListarMisPedidos(w http.ResponseWriter, r *http.Request) {
	const endpoint = "GET /api/pedidos"

	usuario, ok := h.usuario(r.Context())
	if !ok {
		apierr.Responder(w, endpoint, apierr.ErrNoAutenticado)
		return
	}

	pagina, err := enteroDeQuery(r, "pagina")
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	tamano, err := enteroDeQuery(r, "tamano")
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}

	res, err := h.consulta.MisPedidos(r.Context(), usuario.ID, pagina, tamano)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.APaginaPedidos(res))
}

// ObtenerPedido atiende GET /api/pedidos/{id} (SPEC-H16). Si el pedido es de
// otro usuario devuelve 404, no 403: el problema es la pertenencia, no el rol.
// Para quien pregunta, el pedido del vecino no existe; un 403 confirmaría que
// existe, y en un listado de ids eso ya es información gratis.
func (h *Handler) ObtenerPedido(w http.ResponseWriter, r *http.Request) {
	const endpoint = "GET /api/pedidos/{id}"

	usuario, ok := h.usuario(r.Context())
	if !ok {
		apierr.Responder(w, endpoint, apierr.ErrNoAutenticado)
		return
	}

	// Un id que ni siquiera es un uuid tampoco existe: mismo 404, así no se
	// distingue "mal escrito" de "no es tuyo".
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierr.Responder(w, endpoint, apierr.ErrNoEncontrado)
		return
	}

	pedido, err := h.consulta.Detalle(r.Context(), id, usuario.ID)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.APedidoDetalle(pedido))
}

// ObtenerPedidoPublico atiende GET /api/pedidos/publico/{token}: el pedido del
// enlace privado del invitado. Es público A PROPÓSITO (no pide sesión: el
// secreto es el token), y por eso devuelve una vista reducida: el DTO
// PedidoPublicoResponse enumera los campos a mano y el repositorio tiene su
// propio Select. Reusar los de las pantallas propias sería el bug: el día que
// alguien le agregue un campo al pedido, se publicaría en internet sin que
// nadie lo decidiera.
func (h *Handler) ObtenerPedidoPublico(w http.ResponseWriter, r *http.Request) {
	const endpoint = "GET /api/pedidos/publico/{token}"

	// El enlace es privado: que ningún caché ni sitio externo lo guarde.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")

	// El token es un secreto: no se loguea ni se devuelve en errores. Si no
	// tiene el largo de un token, no puede ser uno y ni se consulta la base.
	token := chi.URLParam(r, "token")
	if len(token) != longitudToken {
		apierr.Responder(w, endpoint, apierr.ErrNoEncontrado)
		return
	}

	pedido, err := h.consulta.Publico(r.Context(), token)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusOK, dto.APedidoPublico(pedido))
}

// enteroDeQuery lee un parámetro entero opcional de la query. Ausente es 0
// (el caso de uso lo reemplaza por el valor por defecto); un texto que no es
// número es un 400 con el nombre del campo.
func enteroDeQuery(r *http.Request, nombre string) (int, error) {
	texto := r.URL.Query().Get(nombre)
	if texto == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(texto)
	if err != nil {
		return 0, apierr.ErrValidacion{Campos: map[string]string{nombre: "Tiene que ser un número entero"}}
	}
	return n, nil
}
