@AGENTS.md

## Reglas permanentes para el asistente (copia personal de Agustín)

Valen además de AGENTS.md. Si alguna choca con AGENTS.md, `docs/specs.md`, `docs/contratos-frontend.md` o un ADR, mandan esos documentos.

1. En tareas de frontend, usar la skill Frontend Design (plugin `frontend-design@claude-plugins-official`) y la dirección visual de `DESIGN.md`.
2. Respetar la identidad visual del proyecto: los tokens viven en `frontend/src/app/globals.css` y se describen en `DESIGN.md`.
3. Priorizar rendimiento y experiencia de uso; medir antes de afirmar que algo mejoró.
4. No generar componentes innecesarios: reutilizar `components/ui/`, `components/formulario.tsx`, `Campo`, `SinDatos` y `ErrorDeCarga`.
5. No usar diseños genéricos de IA (ver «Qué evitamos» en `DESIGN.md`).
6. Mantener la arquitectura y las convenciones existentes (AGENTS.md, ADR-001 a ADR-007).
7. Antes de dar por terminada una funcionalidad: `make check` en backend; `lint`, `typecheck` y `build` en frontend; y las pruebas e2e afectadas (`npm run test:e2e`).
8. Inspeccionar visualmente las páginas modificadas (Playwright, a 375px y a escritorio) antes de entregarlas.
