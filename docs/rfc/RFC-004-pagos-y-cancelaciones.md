# RFC-004 — Conciliación de pagos y cancelaciones

- Estado: abierto

## Preguntas a resolver

- Medios de pago admitidos, vigencia de la operación externa y consulta al vencer la reserva.
- Política ante una aprobación tardía después de liberar unidades.
- Estados definitivos de pedido/pago/factura.
- Cancelación después de despacho o entrega, restitución física y tratamiento fiscal.
- Contrato de idempotencia y datos requeridos por la API fiscal.

## Límite vigente

Se mantienen las garantías de `specs.md`: no se confirma desde el navegador, no se completa con stock negativo, no se duplica un cobro/movimiento/uso de cupón y no se declara un reintegro automático. La aprobación tardía se registra como incidencia para resolución controlada.
