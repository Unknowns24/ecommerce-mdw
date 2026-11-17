# Tareas hasta el Parcial II (17/11)

Un archivo por integrante. Cada uno se puede leer solo —y pasárselo tal cual al asistente de IA— porque dice qué es tuyo, qué consumís de los demás (siempre por un contrato ya escrito, nunca esperando a una persona), qué hacés y en qué orden, y cuándo cada tarea está terminada.

| Integrante | Archivo                     | Módulo                                                      |
| ---------- | --------------------------- | ----------------------------------------------------------- |
| Genaro     | [genaro.md](genaro.md)      | Catálogo, sistema visual y base del frontend                |
| Yasmín     | [yasmin.md](yasmin.md)      | Identidad, panel y pruebas de permisos                      |
| Nicolás    | [nicolas.md](nicolas.md)    | Carrito, checkout, consulta de pedidos y prueba del flujo   |
| Agustín    | [agustin.md](agustin.md)    | Back-office de pedidos y stock, infraestructura y CI        |

## Por qué nadie espera a nadie

1. **Las direcciones de las pantallas ya están fijadas** en el mapa de pantallas (`docs/specs.md`, sección 5). Para enlazar a una pantalla de otro alcanza con su dirección: no hace falta que exista.
2. **Los endpoints ya existen** y están en `docs/api.md` y `docs/api/*.http`. Ninguna pantalla espera backend nuevo, salvo `GET /api/tienda/configuracion`, que es del mismo dueño de la pantalla que lo usa.
3. **Las piezas compartidas del frontend tienen firma congelada** en `docs/contratos-frontend.md`. Los cuatro lugares donde un módulo usa algo de otro (sesión, menú de sesión, botón y enlace del carrito) nacen como *stub* en la base del día 0 y su dueño los completa sin cambiar la firma.
4. **Lo único que se espera es la base del día 0** (PR de Genaro, viernes 06/11). Mientras tanto, los otros tres tienen trabajo que no la necesita: revisar su arreglo de backend, el harness de pruebas, la lógica del carrito y la infraestructura.
5. **Las pruebas se escriben contra la spec y el contrato**, no contra el código de otro: se pueden escribir en paralelo.

La única dependencia que no se puede evitar es la última: la prueba de punta a punta del flujo (`e2e/flujo-principal.spec.ts`, de Nicolás) recorre pantallas de los cuatro, y por eso es la última tarea del calendario.

## Reglas para todos

- Leé `AGENTS.md` antes de escribir código. Si usás un asistente, va a leerlo solo.
- Una rama por tarea (`feat/<modulo>/<descripcion>`), PR chico contra `develop`, una revisión de alguien que no lo escribió. Un PR por persona por día hábil, y una revisión a un compañero por día (rotación en `docs/plan-de-trabajo.md`).
- Commits con tu propia cuenta. La nota es individual.
- Cada pantalla, antes del PR: los cinco estados (si lee datos), los dos errores en su lugar (si escribe), recorrido con teclado y a 375px. El checklist del template de PR lo pide.
- No edites carpetas ajenas (`docs/contratos-frontend.md`, sección 1). Si necesitás algo de otro módulo, Issue a su dueño.
- Lo que está en «Fuera de alcance» de la spec no se hace, aunque sobre tiempo, hasta que el flujo esté completo y probado.
- **Domingo 15/11: congelamiento.** Desde ahí solo se mergean arreglos. Ensayo: cada uno explica el módulo de otro.
