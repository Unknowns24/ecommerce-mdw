package schema

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// MigracionesInventario returns the inventory and payments module's
// migrations. It runs third, after identidad and catalogo: lote references
// variante (catálogo) and movimiento_stock references usuario (identidad),
// both of which must already exist. pedido_id columns carry no foreign key
// because the pedidos module migrates last; that integrity is enforced by
// the inventory service, not by the schema.
//
// 20261004_03_inventario ya está mergeada y no se edita: un cambio de
// esquema se agrega acá como una migración nueva, nunca modificando una
// existente.
func MigracionesInventario() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID:       "20261004_03_inventario",
			Migrate:  migrarInventario,
			Rollback: revertirInventario,
		},
		{
			ID:       "20261004_03b_pagos",
			Migrate:  migrarPagos,
			Rollback: revertirPagos,
		},
	}
}

func migrarInventario(tx *gorm.DB) error {
	sentencias := []string{
		`CREATE TABLE proveedor (
			id uuid PRIMARY KEY,
			nombre varchar(255) NOT NULL,
			correo varchar(255),
			telefono varchar(50),
			web varchar(255),
			creado_en timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE lote (
			id uuid PRIMARY KEY,
			variante_id uuid NOT NULL,
			proveedor_id uuid NOT NULL,
			unidades_ingresadas integer NOT NULL CHECK (unidades_ingresadas > 0),
			costo_unitario_centavos bigint NOT NULL CHECK (costo_unitario_centavos >= 0),
			fecha_ingreso timestamptz NOT NULL,
			creado_en timestamptz NOT NULL DEFAULT now(),
			-- RESTRICT: una variante con lotes ingresados no se borra, se desactiva
			CONSTRAINT fk_lote_variante FOREIGN KEY (variante_id) REFERENCES variante(id) ON DELETE RESTRICT,
			-- RESTRICT: borrar un proveedor no puede llevarse puestos sus lotes
			CONSTRAINT fk_lote_proveedor FOREIGN KEY (proveedor_id) REFERENCES proveedor(id) ON DELETE RESTRICT
		)`,
		`CREATE INDEX idx_lote_variante_id ON lote (variante_id)`,
		`CREATE INDEX idx_lote_proveedor_id ON lote (proveedor_id)`,
		`CREATE TABLE movimiento_stock (
			id uuid PRIMARY KEY,
			lote_id uuid NOT NULL,
			tipo varchar(20) NOT NULL CHECK (tipo IN ('INGRESO', 'AJUSTE', 'SALIDA')),
			unidades integer NOT NULL,
			motivo varchar(255) NOT NULL,
			responsable_id uuid,
			pedido_id uuid,
			creado_en timestamptz NOT NULL DEFAULT now(),
			-- RESTRICT: borrar un lote no puede llevarse su historial de movimientos
			CONSTRAINT fk_movimiento_lote FOREIGN KEY (lote_id) REFERENCES lote(id) ON DELETE RESTRICT,
			-- RESTRICT: un usuario con movimientos registrados no se borra, se desactiva
			CONSTRAINT fk_movimiento_responsable FOREIGN KEY (responsable_id) REFERENCES usuario(id) ON DELETE RESTRICT
		)`,
		`CREATE INDEX idx_movimiento_lote_id ON movimiento_stock (lote_id)`,
		`CREATE TABLE reserva_stock (
			id uuid PRIMARY KEY,
			pedido_id uuid NOT NULL,
			lote_id uuid NOT NULL,
			unidades integer NOT NULL CHECK (unidades > 0),
			estado varchar(20) NOT NULL CHECK (estado IN ('ACTIVA', 'CONSUMIDA', 'LIBERADA')),
			vence_en timestamptz,
			creado_en timestamptz NOT NULL DEFAULT now(),
			-- RESTRICT: borrar un lote no puede llevarse sus reservas, son historia
			CONSTRAINT fk_reserva_lote FOREIGN KEY (lote_id) REFERENCES lote(id) ON DELETE RESTRICT
		)`,
		`CREATE INDEX idx_reserva_pedido_id ON reserva_stock (pedido_id)`,
		`CREATE INDEX idx_reserva_lote_estado ON reserva_stock (lote_id, estado)`,
	}

	for _, sentencia := range sentencias {
		if err := tx.Exec(sentencia).Error; err != nil {
			return err
		}
	}
	return nil
}

func revertirInventario(tx *gorm.DB) error {
	return tx.Exec(`DROP TABLE IF EXISTS reserva_stock, movimiento_stock, lote, proveedor CASCADE`).Error
}

// migrarPagos crea intento_pago. El índice único en (proveedor,
// referencia_externa) es lo que hace imposible procesar dos veces la misma
// notificación de Mercado Pago. pedido_id no lleva FK por el mismo motivo
// que reserva_stock.pedido_id: pedido migra después.
func migrarPagos(tx *gorm.DB) error {
	sentencias := []string{
		`CREATE TABLE intento_pago (
			id uuid PRIMARY KEY,
			pedido_id uuid NOT NULL,
			proveedor varchar(20) NOT NULL CHECK (proveedor IN ('MERCADO_PAGO', 'EFECTIVO')),
			referencia_externa varchar(255) NOT NULL,
			estado varchar(20) NOT NULL CHECK (estado IN ('PENDIENTE', 'APROBADO', 'RECHAZADO')),
			monto_centavos bigint NOT NULL CHECK (monto_centavos >= 0),
			creado_en timestamptz NOT NULL DEFAULT now(),
			actualizado_en timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE UNIQUE INDEX idx_intento_pago_proveedor_referencia ON intento_pago (proveedor, referencia_externa)`,
		`CREATE INDEX idx_intento_pago_pedido_id ON intento_pago (pedido_id)`,
	}
	for _, sentencia := range sentencias {
		if err := tx.Exec(sentencia).Error; err != nil {
			return err
		}
	}
	return nil
}

func revertirPagos(tx *gorm.DB) error {
	return tx.Exec(`DROP TABLE IF EXISTS intento_pago CASCADE`).Error
}
