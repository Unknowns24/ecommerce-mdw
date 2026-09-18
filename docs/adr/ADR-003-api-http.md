# ADR-003 — API HTTP con Chi

- Estado: aceptado
- Fecha: 2026-09-18

## Contexto

La tienda, el panel administrativo y los webhooks necesitan endpoints HTTP con rutas agrupables, middleware estándar y posibilidad de pruebas con `net/http`.

## Decisión

Se usará [`github.com/go-chi/chi/v5`](https://github.com/go-chi/chi) como framework/router HTTP. Chi se integra sobre `net/http`; handlers, middleware y pruebas usarán las interfaces estándar de Go.

Las rutas se organizarán por frontera: públicas de catálogo/checkout, autenticación/cuenta, administrativas y webhooks. La autorización se aplica en middleware y se revalida por operación/recurso en los casos de uso, conforme a `specs.md`.

## Consecuencias

- La capa HTTP solo traduce HTTP a comandos/consultas y respuestas; las reglas de precio, stock, RBAC y estados no viven en handlers.
- Los webhooks verifican autenticidad y usan operaciones idempotentes; una URL de retorno del navegador no confirma pagos.
- Se establece un formato de errores coherente al diseñar el primer endpoint, sin inventarlo anticipadamente en este ADR.
