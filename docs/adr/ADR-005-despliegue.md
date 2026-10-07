# ADR-005 — Despliegue en VPS propia, no en Vercel

- Estado: aceptado
- Fecha: 2026-10-04

## Contexto

El backend es un binario de Go que expone una API HTTP y necesita una base PostgreSQL persistente con migraciones versionadas corridas de forma explícita (ADR-002). La cátedra exige una URL pública fija corriendo en producción como condición de aprobación del Parcial I, y puntúa "Producción" en su rúbrica, con un nivel destacado para quien tenga *preview deployments* por Pull Request.

Vercel es la plataforma sugerida por la materia para el ejemplo de referencia, pero está pensada para funciones serverless de corta duración y frontends; no ofrece out-of-the-box un proceso de larga duración con un pool de conexiones persistente a Postgres, que es exactamente lo que este backend necesita.

## Decisión

Se despliega en una VPS propia: un contenedor Docker con el binario de Go (`backend/Dockerfile`, build multi-stage con `golang:1.26-alpine`, imagen final `alpine`, `CGO_ENABLED=0`, usuario no root), Postgres 16 en el mismo `docker-compose.yml` con volumen persistente, y Nginx como reverse proxy con TLS (Let's Encrypt) hacia el contenedor de la API.

El pipeline de CI/CD (`.github/workflows/cicd.yaml`) tiene dos jobs: `verificar` corre en cada push y cada PR (`go build`, `go vet`, `go test`, `gofmt -l`); `desplegar` corre sólo en push a `main` y hace SSH a la VPS, `git pull`, `docker compose build`, corre `./api migrate` contra la base de producción y recién después `docker compose up -d`. Las migraciones se corren antes de levantar el servidor nuevo, nunca al arrancar el proceso (ADR-002).

## Consecuencias

**Ganamos:** control total del runtime de Go (no hay cold starts ni límites de duración de función), una sola máquina sirve tanto la API como la base de datos sin saltos de red entre servicios, y el comando `./api migrate` corre como un paso explícito y auditable del pipeline en lugar de magia de plataforma.

**Resignamos:** no hay *preview deployment* automático por Pull Request — el nivel "destacado" del criterio de Producción de la rúbrica, que en Vercel sale gratis, en una VPS requiere armar un entorno efímero por rama (reservado para una iteración futura, ver `docs/plan-de-trabajo.md`). La URL pública es una sola y no cambia en todo el cuatrimestre: se documenta en `readme.md`, sección Producción.

La elección del motor de base de datos (PostgreSQL) y su driver se completa en [ADR-002](ADR-002-persistencia-y-migraciones.md#actualización--2026-10-04).
