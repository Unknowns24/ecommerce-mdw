package domain

import (
	"github.com/google/uuid"
	"time"
)

type Usuario struct {
	ID             uuid.UUID `gorm:"column:id;primaryKey" json:"id"`
	Nombre         string    `gorm:"column:nombre" json:"nombre"`
	Apellido       string    `gorm:"column:apellido" json:"apellido"`
	Correo         string    `gorm:"column:correo" json:"correo"`
	CredencialHash string    `gorm:"column:credencial_hash" json:"-"`
	Telefono       *string   `gorm:"column:telefono" json:"telefono,omitempty"`
	Estado         string    `gorm:"column:estado" json:"estado"`
	Origen         string    `gorm:"column:origen" json:"origen"`
	EsDuenoInicial bool      `gorm:"column:es_dueno_inicial" json:"es_dueno_inicial"`
	CreadoEn       time.Time `gorm:"column:creado_en;autoCreateTime" json:"creado_en"`
	ActualizadoEn  time.Time `gorm:"column:actualizado_en;autoUpdateTime" json:"actualizado_en"`
}

func (Usuario) TableName() string { return "usuario" }

type Rol struct {
	ID       uuid.UUID `gorm:"column:id;primaryKey" json:"id"`
	Nombre   string    `gorm:"column:nombre" json:"nombre"`
	CreadoEn time.Time `gorm:"column:creado_en;autoCreateTime" json:"creado_en"`
}

func (Rol) TableName() string { return "rol" }

type Permiso struct {
	ID          uuid.UUID `gorm:"column:id;primaryKey" json:"id"`
	Accion      string    `gorm:"column:accion" json:"accion"`
	Descripcion string    `gorm:"column:descripcion" json:"descripcion"`
}

func (Permiso) TableName() string { return "permiso" }

type UsuarioRol struct {
	UsuarioID uuid.UUID `gorm:"column:usuario_id;primaryKey"`
	RolID     uuid.UUID `gorm:"column:rol_id;primaryKey"`
}

func (UsuarioRol) TableName() string { return "usuario_rol" }

type RolPermiso struct {
	RolID     uuid.UUID `gorm:"column:rol_id;primaryKey"`
	PermisoID uuid.UUID `gorm:"column:permiso_id;primaryKey"`
}

func (RolPermiso) TableName() string { return "rol_permiso" }
