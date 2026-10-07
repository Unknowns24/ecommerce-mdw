package identidad

import (
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

func PuedeCambiarEstado(actor, objetivo domain.Usuario, nuevoEstado string) error {
	if objetivo.EsDuenoInicial {
		return apierr.ErrRegla{Mensaje: "El dueño inicial no puede ser modificado"}
	}
	if actor.ID == objetivo.ID && nuevoEstado == "INACTIVO" {
		return apierr.ErrRegla{Mensaje: "No podés desactivarte a vos mismo"}
	}
	return nil
}
func PuedeAsignarRoles(actor, objetivo domain.Usuario, roles []string) error {
	if objetivo.EsDuenoInicial {
		return apierr.ErrRegla{Mensaje: "El dueño inicial no puede ser modificado"}
	}
	return nil
}
