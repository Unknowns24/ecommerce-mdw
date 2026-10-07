package schema

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// MigracionesPedidos crea las tablas del carrito, el pedido y su detalle. Corre
// última de las cuatro porque tiene FKs a usuario (identidad) y variante
// (catálogo), y porque las reservas de stock de inventario apuntan a pedido.
//
// Una migración mergeada no se edita nunca: un error se corrige con otra
// migración nueva.
func MigracionesPedidos() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "20261004_04_pedidos",
			Migrate: func(tx *gorm.DB) error {
				for _, sentencia := range crearPedidos {
					if err := tx.Exec(sentencia).Error; err != nil {
						return err
					}
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				for _, sentencia := range borrarPedidos {
					if err := tx.Exec(sentencia).Error; err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}

var crearPedidos = []string{
	// Configuración de la tienda y sus tarifas por distancia. El pedido copia
	// el costo de envío que se cotizó, así que cambiar una tarifa nunca
	// altera pedidos ya creados.
	`CREATE TABLE configuracion_tienda (
		id                      uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
		efectivo_habilitado     boolean     NOT NULL DEFAULT false,
		distancia_maxima_metros integer     NOT NULL CHECK (distancia_maxima_metros >= 0),
		creado_en               timestamptz NOT NULL DEFAULT now()
	)`,

	// onDelete CASCADE: las tarifas son parte de la configuración, no
	// historial; el pedido ya guardó el costo que se cobró.
	`CREATE TABLE tarifa_distancia (
		id               uuid    PRIMARY KEY DEFAULT gen_random_uuid(),
		configuracion_id uuid    NOT NULL REFERENCES configuracion_tienda (id) ON DELETE CASCADE,
		desde_metros     integer NOT NULL CHECK (desde_metros >= 0),
		hasta_metros     integer NOT NULL,
		costo_centavos   bigint  NOT NULL CHECK (costo_centavos >= 0),
		CHECK (hasta_metros > desde_metros)
	)`,

	// Carrito del cliente registrado. onDelete CASCADE hacia usuario: el
	// carrito es una selección editable, no historial; si la cuenta se va,
	// el carrito se va con ella. UNIQUE: un usuario tiene un solo carrito (y
	// ese UNIQUE ya crea el índice carrito(usuario_id)).
	`CREATE TABLE carrito (
		id             uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
		usuario_id     uuid        UNIQUE REFERENCES usuario (id) ON DELETE CASCADE,
		creado_en      timestamptz NOT NULL DEFAULT now(),
		actualizado_en timestamptz NOT NULL DEFAULT now()
	)`,

	// onDelete CASCADE hacia carrito (el carrito es una selección editable) y
	// hacia variante (una línea de carrito no es historial de ventas).
	// UNIQUE (carrito_id, variante_id): una variante aparece una sola vez por
	// carrito; "agregar una que ya está" suma unidades, y esta restricción
	// impide duplicar la línea aunque dos requests lleguen a la vez.
	`CREATE TABLE item_carrito (
		id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
		carrito_id  uuid        NOT NULL REFERENCES carrito (id) ON DELETE CASCADE,
		variante_id uuid        NOT NULL REFERENCES variante (id) ON DELETE CASCADE,
		unidades    integer     NOT NULL CHECK (unidades > 0),
		creado_en   timestamptz NOT NULL DEFAULT now(),
		UNIQUE (carrito_id, variante_id)
	)`,
	`CREATE INDEX idx_item_carrito_carrito_id ON item_carrito (carrito_id)`,

	// Número legible y secuencial del pedido (el id es un uuid no adivinable;
	// el número es el que se le dice al cliente por teléfono).
	`CREATE SEQUENCE pedido_numero_seq`,

	// onDelete RESTRICT hacia usuario: un usuario con pedidos no se borra, se
	// desactiva; el historial de ventas no se va con la baja de una cuenta.
	// usuario_id es nulo cuando compró un invitado.
	//
	// Las listas cerradas van como CHECK, nunca como varchar libre, y las
	// reglas que el negocio no negocia también quedan en la base: el envío
	// exige domicilio y distancia, el efectivo solo existe con retiro, y el
	// total es siempre subtotal + envío.
	`CREATE TABLE pedido (
		id                 uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
		numero             bigint       NOT NULL UNIQUE DEFAULT nextval('pedido_numero_seq'),
		usuario_id         uuid         REFERENCES usuario (id) ON DELETE RESTRICT,
		comprador_nombre   varchar(100) NOT NULL,
		comprador_apellido varchar(100) NOT NULL,
		comprador_correo   varchar(254) NOT NULL,
		comprador_telefono varchar(30)  NOT NULL,
		comprador_dni      varchar(20)  NOT NULL,
		modo_entrega       varchar(10)  NOT NULL CHECK (modo_entrega IN ('RETIRO', 'ENVIO')),
		domicilio          varchar(300),
		distancia_metros   integer      CHECK (distancia_metros >= 0),
		envio_centavos     bigint       NOT NULL CHECK (envio_centavos >= 0),
		subtotal_centavos  bigint       NOT NULL CHECK (subtotal_centavos >= 0),
		total_centavos     bigint       NOT NULL CHECK (total_centavos >= 0),
		estado_pedido      varchar(20)  NOT NULL CHECK (estado_pedido IN
			('PENDIENTE_DE_PAGO', 'CONFIRMADO', 'DESPACHADO', 'COMPLETADO', 'CANCELADO')),
		estado_pago        varchar(10)  NOT NULL CHECK (estado_pago IN
			('PENDIENTE', 'APROBADO', 'RECHAZADO')),
		medio_pago         varchar(15)  NOT NULL CHECK (medio_pago IN ('MERCADO_PAGO', 'EFECTIVO')),
		token_acceso       varchar(64)  NOT NULL UNIQUE,
		vence_en           timestamptz,
		cancelado_por      uuid         REFERENCES usuario (id) ON DELETE RESTRICT,
		cancelado_en       timestamptz,
		motivo_cancelacion varchar(500),
		creado_en         timestamptz  NOT NULL DEFAULT now(),
		actualizado_en     timestamptz  NOT NULL DEFAULT now(),
		CHECK (modo_entrega = 'RETIRO' OR (domicilio IS NOT NULL AND distancia_metros IS NOT NULL)),
		CHECK (medio_pago <> 'EFECTIVO' OR modo_entrega = 'RETIRO'),
		CHECK (total_centavos = subtotal_centavos + envio_centavos)
	)`,
	`ALTER SEQUENCE pedido_numero_seq OWNED BY pedido.numero`,
	// pedido(usuario_id) es la FK por la que se filtra en cada "mis pedidos":
	// sin índice, cada pantalla recorre la tabla entera.
	`CREATE INDEX idx_pedido_usuario_id    ON pedido (usuario_id)`,
	`CREATE INDEX idx_pedido_estado_pedido ON pedido (estado_pedido)`,
	`CREATE INDEX idx_pedido_creado_en     ON pedido (creado_en)`,

	// DetallePedido es la N-N con datos propios entre pedido y variante.
	// codigo, nombre y precio_unitario_centavos son una copia inmutable de lo
	// que se compró ese día, no una copia del catálogo.
	// onDelete RESTRICT hacia pedido: los pedidos no se borran, se cancelan.
	// onDelete RESTRICT hacia variante: un producto vendido no se borra del
	// catálogo (se desactiva).
	`CREATE TABLE detalle_pedido (
		id                       uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
		pedido_id                uuid         NOT NULL REFERENCES pedido (id) ON DELETE RESTRICT,
		variante_id              uuid         NOT NULL REFERENCES variante (id) ON DELETE RESTRICT,
		codigo                   varchar(60)  NOT NULL,
		nombre                   varchar(200) NOT NULL,
		unidades                 integer      NOT NULL CHECK (unidades > 0),
		precio_unitario_centavos bigint       NOT NULL CHECK (precio_unitario_centavos >= 0),
		subtotal_centavos        bigint       NOT NULL CHECK (subtotal_centavos >= 0),
		CHECK (subtotal_centavos = unidades * precio_unitario_centavos)
	)`,
	`CREATE INDEX idx_detalle_pedido_pedido_id ON detalle_pedido (pedido_id)`,
}

// El rollback borra en el orden inverso al de creación, para no violar las FK.
var borrarPedidos = []string{
	`DROP TABLE detalle_pedido`,
	`DROP TABLE pedido`,
	`DROP SEQUENCE IF EXISTS pedido_numero_seq`,
	`DROP TABLE item_carrito`,
	`DROP TABLE carrito`,
	`DROP TABLE tarifa_distancia`,
	`DROP TABLE configuracion_tienda`,
}
