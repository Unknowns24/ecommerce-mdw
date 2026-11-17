## Qué hace este PR

<Una o dos oraciones, con la historia (SPEC-Hxx) y la fila del mapa de pantallas o de `docs/api.md` que toca.>

## Revisión de código — backend

- [ ] ¿Algún endpoint quedó sin middleware de sesión y sin decir que es público a propósito?
- [ ] ¿Alguna consulta devuelve datos de un usuario sin su id en el WHERE?
- [ ] ¿Aparece `rol` o `usuarioID` en un DTO, un header propio o la query?
- [ ] ¿Algún listado sin `Limit` o sin `Select`?
- [ ] ¿Alguna regla de negocio quedó adentro de un handler?
- [ ] ¿Las rutas coinciden con `docs/api.md` y `docs/api/`?
- [ ] ¿`docs/api/*.http` tiene los tres casos de cada endpoint nuevo?
- [ ] ¿`go build`, `go vet` y `go test` están en verde?

## Revisión de código — frontend

- [ ] ¿La dirección de cada pantalla nueva está en el mapa de pantallas (spec, sección 5)?
- [ ] ¿La pantalla resuelve sus cinco estados, con textos distintos para vacío y sin resultados?
- [ ] ¿Cada Server Action verifica la sesión cuando la operación la exige?
- [ ] ¿Algún id de usuario, rol o precio viaja en el formulario en vez de salir de la sesión o la URL?
- [ ] ¿Errores 400 debajo del campo y 409 arriba del formulario? ¿El formulario conserva lo escrito al fallar?
- [ ] ¿El botón se deshabilita mientras guarda? ¿Hay `revalidatePath` de cada pantalla que quedó vieja?
- [ ] ¿Algún `.css` nuevo, `style={{}}` o borde/radio escrito a mano en `src/app/`?
- [ ] ¿Algún Client Component llama a la API o algún `NEXT_PUBLIC_` lleva un secreto?
- [ ] Recorrido con teclado y a 375px. Si es un formulario del flujo, también en un teléfono real contra producción: ¿qué se vio?
- [ ] ¿`npm run lint`, `npm run typecheck` y `npm run build` están en verde?

## Pruebas

- [ ] Si cambió una regla de negocio (spec, sección 6), cambió su prueba unitaria en Go.
- [ ] Si cambió quién puede hacer qué, o hay un endpoint nuevo, tiene su fila en `docs/api.md` y su caso en `frontend/e2e/permisos.spec.ts`.
- [ ] Si cambió el flujo principal (spec, sección 5), cambió `frontend/e2e/flujo-principal.spec.ts`.
- [ ] Ninguna prueba se arregló cambiando el valor esperado.
- [ ] Si es una prueba nueva: ¿la vieron fallar? Contar cómo.
