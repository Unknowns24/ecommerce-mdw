# ADR-004 — Dinero en centavos (`int64`) e identificadores `uuid.UUID`

- Estado: aceptado
- Fecha: 2026-10-04

## Contexto

El sistema maneja importes (precios, costos, totales de pedido) e identificadores que se exponen en URLs públicas (el pedido de un comprador, el enlace privado de un invitado). Dos decisiones de tipo de dato tienen consecuencias que no se pueden corregir después sin migrar datos en producción, así que se fijan ahora.

## Decisión

**Dinero:** todo importe se guarda como `int64` en centavos, en columnas `bigint`, con el campo nombrado `..._centavos` (`precio_minorista_centavos`, `subtotal_centavos`, `total_centavos`). Nunca `float64` ni `float32`.

`float64` no representa exacto un valor como 0,10 en base 2: la representación más cercana tiene un error que se acumula al sumar líneas de un pedido, y con suficientes operaciones el total deja de coincidir con la suma manual. `decimal` (de terceros, o `numeric` de Postgres mapeado a un tipo decimal) evita ese error pero agrega una dependencia y una conversión en cada capa sin necesidad: enteros en centavos ya son exactos, rápidos y no requieren librería.

**Identificadores:** todo id es `uuid.UUID` en Go, columna `uuid` en Postgres (`gen_random_uuid()` o generado en la aplicación). Nunca autoincremental.

Un id autoincremental revela cuántas filas existen (cuántos pedidos tiene el negocio) y permite adivinar URLs vecinas: `/api/pedidos/1042` sugiere que `/api/pedidos/1041` también existe. El enlace privado de un pedido de invitado (`GET /api/pedidos/publico/{token}`) depende explícitamente de que el identificador no sea adivinable; un UUID v4 generado con un generador criptográficamente aleatorio cumple esa propiedad, un contador no.

## Consecuencias

- Toda suma de dinero en el backend opera sobre `int64`; convertir a una representación decimal (`"$15.900,00"`) es responsabilidad exclusiva de la capa de presentación, nunca del backend.
- Todo `CREATE TABLE` usa `bigint` para columnas de dinero y `uuid` para claves primarias y foráneas; no hay excepciones por conveniencia de un módulo.
- Los cuatro módulos respetan esta decisión en sus propias migraciones: no es negociable por dominio, es una regla del sistema completo.
