# Frontend de BC Importados

Desde `frontend/`, instalá las dependencias y configurá la URL de la API de Go en `.env.local`.

```bash
npm ci
cp .env.example .env.local
npm run dev
```

Abrí [la tienda local](http://localhost:3000). El cliente de la API lee `API_BASE_URL` en el servidor de Next.js.

```bash
npm run lint
npm run typecheck
npm run build
```

El build genera `.next/standalone/server.js`. El [Dockerfile](Dockerfile) ejecuta ese servidor y copia `public/` y `.next/static/` a la imagen.
