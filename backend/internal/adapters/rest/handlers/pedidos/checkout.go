package pedidos

import (
	"context"
	"encoding/json"
	"net/http"

	dto "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/dtos/pedidos"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/middleware"
	uc "github.com/Unknowns24/ecommerce-mdw/internal/usecases/pedidos"
)

const maxBody = 1 << 20 // 1 MiB: un pedido razonable pesa unos pocos KB

// Handler agrupa los endpoints del módulo. Solo traduce HTTP: lee el body,
// llama al caso de uso y convierte el resultado en un código de estado. No
// decide reglas de negocio.
type Handler struct {
	checkout *uc.Checkout
	consulta *uc.Consulta
	gestion  *uc.Gestion
	carrito  *uc.GestionCarrito

	// usuario lee la identidad del contexto de la sesión. Es un campo para que
	// los tests puedan simular una sesión sin armar un JWT.
	usuario func(ctx context.Context) (middleware.Usuario, bool)
}

func NuevoHandler(checkout *uc.Checkout, consulta *uc.Consulta, gestion *uc.Gestion, carrito *uc.GestionCarrito) *Handler {
	return &Handler{checkout: checkout, consulta: consulta, gestion: gestion, carrito: carrito, usuario: middleware.UsuarioDeContexto}
}

// CrearPedido atiende POST /api/pedidos (SPEC-H13). La sesión es opcional: el
// invitado compra igual. El orden importa:
//
//  1. ¿hay sesión?          opcional; el invitado compra igual.
//  2. ¿el body está bien?   400 si no (forma del JSON, sin mirar el sistema).
//  3. ¿las reglas dejan?    404/409 (stock, cobertura, efectivo habilitado).
//  4. hacer el trabajo      201.
//
// Los errores esperados viajan como valor y apierr.Responder los traduce: así
// un stock insuficiente es un 409 y no un 500.
func (h *Handler) CrearPedido(w http.ResponseWriter, r *http.Request) {
	const endpoint = "POST /api/pedidos"

	// La identidad sale del contexto de la sesión, nunca del body.
	usuario, hayUsuario := h.usuario(r.Context())

	var req dto.CrearPedidoRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.Responder(w, endpoint, apierr.ErrValidacion{Campos: map[string]string{"body": "El cuerpo debe ser un JSON válido"}})
		return
	}

	datos, err := req.AEntrada()
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	if hayUsuario {
		id := usuario.ID
		datos.UsuarioID = &id
	}

	pedido, err := h.checkout.Crear(r.Context(), datos)
	if err != nil {
		apierr.Responder(w, endpoint, err)
		return
	}
	apierr.JSON(w, http.StatusCreated, dto.APedidoResponse(pedido))
}
