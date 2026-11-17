package pagos

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
	"github.com/google/uuid"
)

type RepositorioIncidencias interface {
	Listar(context.Context, uuid.UUID) ([]domain.IncidenciaPago, error)
	Resolver(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, time.Time) error
}

type GestionIncidencias struct {
	repo  RepositorioIncidencias
	ahora func() time.Time
}

func NuevaGestionIncidencias(repo RepositorioIncidencias, ahora func() time.Time) *GestionIncidencias {
	return &GestionIncidencias{repo: repo, ahora: ahora}
}

func (s *GestionIncidencias) Listar(ctx context.Context, pedido uuid.UUID) ([]domain.IncidenciaPago, error) {
	return s.repo.Listar(ctx, pedido)
}

func (s *GestionIncidencias) Resolver(ctx context.Context, pedido, incidencia, actor uuid.UUID, nota string) error {
	nota, err := ValidarNotaResolucion(nota)
	if err != nil {
		return err
	}
	return s.repo.Resolver(ctx, pedido, incidencia, actor, nota, s.ahora())
}

func ValidarNotaResolucion(nota string) (string, error) {
	nota = strings.TrimSpace(nota)
	if nota == "" || utf8.RuneCountInString(nota) > 1000 {
		return "", apierr.ErrValidacion{Campos: map[string]string{"nota": "Indicá cómo se resolvió la incidencia (hasta 1000 caracteres)."}}
	}
	return nota, nil
}

func MotivoIncidencia(decision Confirmacion) string {
	switch decision {
	case NoConfirmarMontoDistinto:
		return "MONTO_DISTINTO"
	case NoConfirmarPedidoNoPendiente:
		return "PEDIDO_NO_PENDIENTE"
	default:
		return ""
	}
}
