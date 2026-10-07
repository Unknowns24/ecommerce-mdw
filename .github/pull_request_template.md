## Revisión de código

- [ ] ¿Algún endpoint quedó sin middleware de sesión y sin decir que es público a propósito?
- [ ] ¿Alguna consulta devuelve datos de un usuario sin su id en el WHERE?
- [ ] ¿Aparece `rol` o `usuarioID` en un DTO, un header propio o la query?
- [ ] ¿Algún listado sin `Limit` o sin `Select`?
- [ ] ¿Alguna regla de negocio quedó adentro de un handler?
- [ ] ¿Las rutas coinciden con `docs/api/`?
- [ ] ¿`docs/api/*.http` tiene los tres casos de cada endpoint nuevo?
- [ ] ¿`go build`, `go vet` y `go test` están en verde?
