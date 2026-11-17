package schema

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Migra después de pedidos: las incidencias conservan sus FKs e historial.
func MigracionesIncidenciasPagos() []*gormigrate.Migration {
	return []*gormigrate.Migration{{ID: "20261117_05_incidencias_pagos", Migrate: func(tx *gorm.DB) error {
		for _, sql := range []string{
			`ALTER TABLE pedido ADD COLUMN pago_revisado_en timestamptz`,
			`CREATE INDEX idx_pedido_revision_pago ON pedido (pago_revisado_en, vence_en, id) WHERE estado_pedido = 'PENDIENTE_DE_PAGO'`,
			`CREATE TABLE incidencia_pago (
				id uuid PRIMARY KEY,
				pedido_id uuid NOT NULL REFERENCES pedido(id) ON DELETE RESTRICT,
				referencia_externa varchar(255) NOT NULL,
				motivo varchar(40) NOT NULL CHECK (motivo IN ('MONTO_DISTINTO','PEDIDO_NO_PENDIENTE')),
				monto_centavos bigint NOT NULL CHECK (monto_centavos >= 0),
				total_pedido_centavos bigint NOT NULL CHECK (total_pedido_centavos >= 0),
				creado_en timestamptz NOT NULL DEFAULT now(),
				resuelto_por uuid REFERENCES usuario(id) ON DELETE RESTRICT,
				resuelto_en timestamptz,
				nota_resolucion varchar(1000),
				UNIQUE (pedido_id, referencia_externa, motivo),
				CHECK ((resuelto_por IS NULL AND resuelto_en IS NULL AND nota_resolucion IS NULL) OR
				(resuelto_por IS NOT NULL AND resuelto_en IS NOT NULL AND nota_resolucion IS NOT NULL AND btrim(nota_resolucion) <> ''))
			)`,
			`CREATE INDEX idx_incidencia_pago_pedido ON incidencia_pago (pedido_id, creado_en)`,
		} {
			if err := tx.Exec(sql).Error; err != nil {
				return err
			}
		}
		return nil
	}, Rollback: func(tx *gorm.DB) error {
		if err := tx.Exec("DROP TABLE incidencia_pago").Error; err != nil {
			return err
		}
		return tx.Exec("ALTER TABLE pedido DROP COLUMN pago_revisado_en").Error
	}}}
}
