package identidad

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/identidad"
	jwt "github.com/Unknowns24/ecommerce-mdw/internal/adapters/jwt"
	apierr "github.com/Unknowns24/ecommerce-mdw/internal/adapters/rest/errors"
	"github.com/Unknowns24/ecommerce-mdw/internal/core/domain"
)

var errCredenciales = errors.New("Correo o contraseña incorrectos")
var ErrBootstrapClave = errors.New("Clave de inicialización inválida")
var hashDummy, _ = bcrypt.GenerateFromPassword([]byte("contraseña dummy sin cuenta"), 12)

func CredencialesIncorrectas() error { return errCredenciales }

type Servicio struct {
	repo         *repo.Repositorio
	emisor       *jwt.Emisor
	bootstrapKey string
}

func Nuevo(r *repo.Repositorio, e *jwt.Emisor, clave string) *Servicio { return &Servicio{r, e, clave} }

func duplicado(err error) bool { var p *pgconn.PgError; return errors.As(err, &p) && p.Code == "23505" }
func (s *Servicio) Registrar(ctx context.Context, nombre, apellido, correo, clave string, admin bool) (domain.Usuario, error) {
	correo = strings.ToLower(strings.TrimSpace(correo))
	if _, err := s.repo.PorCorreo(ctx, correo); err == nil {
		return domain.Usuario{}, apierr.ErrRegla{Mensaje: "Ese correo ya está registrado"}
	} else if !repo.EsNoEncontrado(err) {
		return domain.Usuario{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(clave), 12)
	if err != nil {
		return domain.Usuario{}, err
	}
	origen, rol := "PUBLICO", "cliente"
	if admin {
		origen, rol = "ADMINISTRATIVO", "cliente"
	}
	u := domain.Usuario{ID: uuid.New(), Nombre: nombre, Apellido: apellido, Correo: correo, CredencialHash: string(hash), Estado: "ACTIVO", Origen: origen}
	if err := s.repo.Crear(ctx, &u, rol); err != nil {
		if duplicado(err) {
			return domain.Usuario{}, apierr.ErrRegla{Mensaje: "Ese correo ya está registrado"}
		}
		return domain.Usuario{}, err
	}
	return u, nil
}
func (s *Servicio) Bootstrap(ctx context.Context, nombre, apellido, correo, clave, claveBootstrap string) (domain.Usuario, error) {
	if existe, err := s.repo.ExisteDueno(ctx); err != nil {
		return domain.Usuario{}, err
	} else if existe {
		return domain.Usuario{}, apierr.ErrRegla{Mensaje: "El sistema ya fue inicializado"}
	}
	if s.bootstrapKey == "" || subtle.ConstantTimeCompare([]byte(claveBootstrap), []byte(s.bootstrapKey)) != 1 {
		return domain.Usuario{}, ErrBootstrapClave
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(clave), 12)
	if err != nil {
		return domain.Usuario{}, err
	}
	u := domain.Usuario{ID: uuid.New(), Nombre: nombre, Apellido: apellido, Correo: strings.ToLower(strings.TrimSpace(correo)), CredencialHash: string(hash), Estado: "ACTIVO", Origen: "ADMINISTRATIVO", EsDuenoInicial: true}
	if err := s.repo.Bootstrap(ctx, &u); err != nil {
		if errors.Is(err, repo.ErrYaInicializado) {
			return domain.Usuario{}, apierr.ErrRegla{Mensaje: "El sistema ya fue inicializado"}
		}
		if duplicado(err) {
			return domain.Usuario{}, apierr.ErrRegla{Mensaje: "Ese correo ya está registrado"}
		}
		return domain.Usuario{}, err
	}
	return u, nil
}
func (s *Servicio) Login(ctx context.Context, correo, clave string) (string, error) {
	u, err := s.repo.PorCorreo(ctx, correo)
	hash := hashDummy
	if err == nil {
		hash = []byte(u.CredencialHash)
	}
	coincide := bcrypt.CompareHashAndPassword(hash, []byte(clave)) == nil
	if err != nil && !repo.EsNoEncontrado(err) {
		return "", err
	}
	if err != nil || !coincide || u.Estado != "ACTIVO" {
		return "", errCredenciales
	}
	roles, err := s.repo.RolesDeUsuario(ctx, u.ID)
	if err != nil {
		return "", err
	}
	permisos, err := s.repo.PermisosDeUsuario(ctx, u.ID)
	if err != nil {
		return "", err
	}
	return s.emisor.Emitir(jwt.Claims{Sub: u.ID, Correo: u.Correo, Roles: roles, Permisos: permisos})
}
func (s *Servicio) Cuenta(ctx context.Context, id uuid.UUID) (domain.Usuario, error) {
	return s.repo.PorID(ctx, id)
}
func (s *Servicio) ActualizarCuenta(ctx context.Context, id uuid.UUID, nombre, apellido string, telefono *string) (domain.Usuario, error) {
	if err := s.repo.ActualizarDatos(ctx, id, nombre, apellido, telefono); err != nil {
		return domain.Usuario{}, err
	}
	return s.repo.PorID(ctx, id)
}
func (s *Servicio) CambiarEstado(ctx context.Context, actorID, objetivoID uuid.UUID, estado string) error {
	actor, err := s.repo.PorID(ctx, actorID)
	if err != nil {
		return err
	}
	objetivo, err := s.repo.PorID(ctx, objetivoID)
	if err != nil {
		return err
	}
	if err := PuedeCambiarEstado(actor, objetivo, estado); err != nil {
		return err
	}
	return s.repo.CambiarEstado(ctx, objetivoID, estado)
}
func (s *Servicio) ReemplazarRoles(ctx context.Context, actorID, objetivoID uuid.UUID, roles []string) error {
	actor, err := s.repo.PorID(ctx, actorID)
	if err != nil {
		return err
	}
	objetivo, err := s.repo.PorID(ctx, objetivoID)
	if err != nil {
		return err
	}
	if objetivo.Origen != "ADMINISTRATIVO" {
		return apierr.ErrRegla{Mensaje: "Sólo una cuenta administrativa puede recibir roles"}
	}
	if err := PuedeAsignarRoles(actor, objetivo, roles); err != nil {
		return err
	}
	vistos := map[string]bool{}
	for _, nombre := range roles {
		if vistos[nombre] {
			return apierr.ErrValidacion{Campos: map[string]string{"roles": "No repitas roles"}}
		}
		vistos[nombre] = true
		if _, err := s.repo.PorNombre(ctx, nombre); err != nil {
			if repo.EsNoEncontrado(err) {
				return apierr.ErrValidacion{Campos: map[string]string{"roles": "Contiene un rol desconocido"}}
			}
			return err
		}
	}
	return s.repo.ReemplazarRoles(ctx, objetivoID, roles)
}

type UsuarioListado struct {
	domain.Usuario
	Roles []string `json:"roles"`
}

func (s *Servicio) ListarUsuarios(ctx context.Context, limit, offset int, estado string) ([]UsuarioListado, error) {
	usuarios, err := s.repo.Listar(ctx, limit, offset, estado)
	if err != nil {
		return nil, err
	}
	items := make([]UsuarioListado, 0, len(usuarios))
	for _, u := range usuarios {
		roles, err := s.repo.RolesDeUsuario(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, UsuarioListado{u, roles})
	}
	return items, nil
}

type RolListado struct {
	domain.Rol
	Permisos []string `json:"permisos"`
}

func (s *Servicio) ListarRoles(ctx context.Context, limit, offset int) ([]RolListado, error) {
	roles, err := s.repo.ListarRoles(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	items := make([]RolListado, 0, len(roles))
	for _, r := range roles {
		p, err := s.repo.PermisosDeRol(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, RolListado{r, p})
	}
	return items, nil
}
func (s *Servicio) CrearRol(ctx context.Context, nombre string, permisos []string) (domain.Rol, error) {
	validos, err := s.repo.ListarPermisos(ctx, 100, 0)
	if err != nil {
		return domain.Rol{}, err
	}
	set := map[string]bool{}
	for _, p := range validos {
		set[p.Accion] = true
	}
	for _, p := range permisos {
		if !set[p] {
			return domain.Rol{}, apierr.ErrValidacion{Campos: map[string]string{"permisos": "Contiene un permiso desconocido"}}
		}
	}
	unique := map[string]bool{}
	for _, p := range permisos {
		if unique[p] {
			return domain.Rol{}, apierr.ErrValidacion{Campos: map[string]string{"permisos": "No repitas permisos"}}
		}
		unique[p] = true
	}
	r := domain.Rol{ID: uuid.New(), Nombre: nombre}
	if err := s.repo.CrearRol(ctx, &r, permisos); err != nil {
		if duplicado(err) {
			return domain.Rol{}, apierr.ErrRegla{Mensaje: "Ese rol ya existe"}
		}
		return domain.Rol{}, err
	}
	return r, nil
}
