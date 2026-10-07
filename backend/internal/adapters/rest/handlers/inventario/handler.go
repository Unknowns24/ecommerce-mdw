package inventario

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	repo "github.com/Unknowns24/ecommerce-mdw/internal/adapters/db/postgresql/repositories/inventario"
	usecases "github.com/Unknowns24/ecommerce-mdw/internal/usecases/inventario"
)

var validar = validator.New()

type handler struct {
	db          *gorm.DB
	lotes       *repo.RepositorioLotes
	movimientos *repo.RepositorioMovimientos
	reservas    *repo.RepositorioReservas
	proveedores *repo.RepositorioProveedores
	stock       *usecases.ServicioStock
}

func nuevoHandler(db *gorm.DB) *handler {
	return &handler{
		db:          db,
		lotes:       repo.NuevoRepositorioLotes(db),
		movimientos: repo.NuevoRepositorioMovimientos(db),
		reservas:    repo.NuevoRepositorioReservas(db),
		proveedores: repo.NuevoRepositorioProveedores(db),
		stock:       usecases.NuevoServicioStock(db),
	}
}

// decodificarYValidar lee el body como JSON en dto y lo valida con
// validator/v10, traduciendo el primer problema de forma a un mapa
// campo → mensaje en castellano.
func decodificarYValidar(r *http.Request, dto any) map[string]string {
	if err := json.NewDecoder(r.Body).Decode(dto); err != nil {
		return map[string]string{"body": "El cuerpo no es un JSON válido"}
	}
	if err := validar.Struct(dto); err != nil {
		campos := map[string]string{}
		for _, fe := range err.(validator.ValidationErrors) {
			campos[fe.Field()] = mensajeValidacion(fe)
		}
		return campos
	}
	return nil
}

func mensajeValidacion(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Es requerido"
	case "uuid":
		return "Tiene que ser un identificador válido"
	case "gt":
		return "Tiene que ser mayor a " + fe.Param()
	case "gte":
		return "Tiene que ser mayor o igual a " + fe.Param()
	case "email":
		return "Tiene que ser un correo válido"
	case "max":
		return "Supera el largo máximo permitido"
	case "min":
		return "No cumple el largo mínimo"
	default:
		return "Dato inválido"
	}
}

// paginacion lee pagina/porPagina de la query con los defaults del sistema
// y devuelve limit/offset ya acotados.
func paginacion(r *http.Request) (limit, offset int) {
	pagina := 1
	if v, err := strconv.Atoi(r.URL.Query().Get("pagina")); err == nil && v >= 1 {
		pagina = v
	}
	porPagina := 20
	if v, err := strconv.Atoi(r.URL.Query().Get("porPagina")); err == nil && v >= 1 && v <= 100 {
		porPagina = v
	}
	return porPagina, (pagina - 1) * porPagina
}
