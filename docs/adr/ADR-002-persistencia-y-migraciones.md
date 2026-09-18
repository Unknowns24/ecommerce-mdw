# ADR-002 — GORM y migraciones versionadas con gormigrate

- Estado: aceptado
- Fecha: 2026-09-18

## Contexto

El sistema necesita persistir pedidos históricos, reservas de stock y trazabilidad. Esas operaciones requieren transacciones, restricciones de integridad y una evolución de esquema repetible entre entornos.

## Decisión

Se utilizará [`gorm.io/gorm`](https://gorm.io/) como ORM. Las migraciones se implementarán con [`github.com/go-gormigrate/gormigrate/v2`](https://github.com/go-gormigrate/gormigrate), con identificadores versionados, ordenados y reversibles cuando sea técnicamente seguro.

Las migraciones vivirán en el backend, se ejecutarán de forma explícita desde la composición de la aplicación o un comando dedicado, y no se sustituirán por `AutoMigrate` al arrancar el servidor. Los modelos de GORM no son el contrato público de la API ni reemplazan las copias históricas requeridas en pedidos.

## Consecuencias

- Las mutaciones de stock, reserva, cupo de cupón y transición de pedido que deben ser atómicas se ejecutarán dentro de una transacción de GORM.
- Las reglas de concurrencia se apoyarán además en restricciones e idioma SQL apropiado del motor elegido; el ORM no garantiza por sí solo la integridad de inventario.
- Cada cambio de esquema requiere una migración nueva. Una migración aplicada no se edita retroactivamente.
- La elección de motor y driver de base de datos sigue pendiente; `DATABASE_DSN` permite desacoplar la configuración, pero no define dialecto ni estrategia de despliegue.
