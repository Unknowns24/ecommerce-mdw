package identidad

import "github.com/Unknowns24/ecommerce-mdw/internal/core/domain"

// RespuestaCuenta es la respuesta de GET /api/cuenta: los datos del usuario y lo que
// su sesión puede hacer. Los roles y permisos son los del token, no los de la
// base: un cambio de roles aplica al volver a ingresar (ADR-006), y la
// interfaz no tiene que ofrecer algo que la sesión todavía no puede hacer.
type RespuestaCuenta struct {
	domain.Usuario
	Roles    []string `json:"roles"`
	Permisos []string `json:"permisos"`
}

func NuevaCuenta(u domain.Usuario, roles, permisos []string) RespuestaCuenta {
	if roles == nil {
		roles = []string{}
	}
	if permisos == nil {
		permisos = []string{}
	}
	return RespuestaCuenta{Usuario: u, Roles: roles, Permisos: permisos}
}
