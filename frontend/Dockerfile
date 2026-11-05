# syntax=docker/dockerfile:1
# Imagen de producción del frontend (ADR-007). Necesita output: "standalone"
# en next.config.ts: el build deja en .next/standalone un server.js con solo
# las dependencias que usa, y la imagen final no lleva node_modules completos.

FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund

FROM node:22-alpine AS build
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME=0.0.0.0
RUN addgroup -S bc && adduser -S bc -G bc
COPY --from=build --chown=bc:bc /app/public ./public
COPY --from=build --chown=bc:bc /app/.next/standalone ./
COPY --from=build --chown=bc:bc /app/.next/static ./.next/static
USER bc
EXPOSE 3000
# API_BASE_URL se lee en tiempo de ejecución, en el servidor de Next: la
# pone docker-compose.yml y no queda horneada en la imagen.
CMD ["node", "server.js"]
