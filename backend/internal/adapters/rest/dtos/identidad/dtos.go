package identidad

type Registro struct {
	Nombre     string `json:"nombre" validate:"required"`
	Apellido   string `json:"apellido" validate:"required"`
	Correo     string `json:"correo" validate:"required,email"`
	Contrasena string `json:"contraseña" validate:"required,min=8"`
}
type Bootstrap struct {
	Registro
	ClaveBootstrap string `json:"clave_bootstrap" validate:"required"`
}
type Login struct {
	Correo     string `json:"correo" validate:"required,email"`
	Contrasena string `json:"contraseña" validate:"required"`
}
type Cuenta struct {
	Nombre   string  `json:"nombre" validate:"required"`
	Apellido string  `json:"apellido" validate:"required"`
	Telefono *string `json:"telefono"`
}
type Estado struct {
	Estado string `json:"estado" validate:"required,oneof=ACTIVO INACTIVO"`
}
type Rol struct {
	Nombre   string   `json:"nombre" validate:"required"`
	Permisos []string `json:"permisos"`
}
type Roles struct {
	Roles []string `json:"roles" validate:"required,min=1"`
}
