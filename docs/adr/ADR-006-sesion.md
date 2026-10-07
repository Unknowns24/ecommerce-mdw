# ADR-006 — Login propio y JWT con permisos adentro, sin consultar la base en cada request

- Estado: aceptado
- Fecha: 2026-10-04

## Contexto

El sistema necesita autenticar compradores y administradores, y autorizar cada operación administrativa por permiso, en todos los caminos — incluidos los que no pasan por la interfaz. El contrato de sesión y permisos (`middleware.RequiereSesion`, `middleware.RequierePermiso`) es la pieza bloqueante de la que dependen los otros tres módulos, así que su diseño se fija temprano y no se renegocia por módulo.

## Decisión

**Proveedor de identidad:** login propio (correo y contraseña con bcrypt, costo 12), no OAuth de terceros. Ganamos no depender de un proveedor externo para la condición de aprobación del parcial ("si el login no funciona, no hay parcial") y control total sobre el bootstrap del dueño inicial, que es una regla de negocio propia sin equivalente en un flujo de OAuth estándar. Resignamos la verificación de identidad real que un proveedor externo aporta (nadie confirma que el correo pertenece a quien dice ser) y tener que mantener nosotros mismos el hasheo y la recuperación de contraseña.

**Estrategia de sesión:** JWT firmado con HS256 (`APP_SECRET_KEY`), con los permisos del usuario adentro del token (claim `permisos`), no una sesión consultada contra la base en cada request. Ganamos cero consultas a la base por request para autorizar — cada handler protegido resuelve `RequierePermiso` con el token ya parseado — y un middleware sin estado, fácil de razonar y de testear sin levantar la base. Resignamos revocación inmediata: si el dueño le cambia los roles a alguien o lo desactiva, ese cambio recién aplica cuando esa persona vuelve a iniciar sesión, porque el token viejo sigue siendo válido hasta que expira (24 h) o se vuelve a emitir. Un sistema que necesitara revocar permisos al instante (por ejemplo, despedir a un empleado y cortarle el acceso ya mismo) no debería elegir este diseño sin agregar una lista de revocación.

## Consecuencias

- `middleware.UsuarioDeContexto` es la única fuente de identidad para un handler; ningún DTO, header propio o parámetro de query puede aportar `usuarioID`, `rol` ni `permiso` — es la regla 2 de los contratos compartidos, y este ADR es la razón técnica de por qué existe.
- La ventana de "rol viejo todavía activo" dura como máximo lo que dura el token (`PAYMENT_RESERVATION_TTL` no aplica acá; la expiración del JWT es un valor propio del middleware, documentado por Yasmín junto con su implementación).
- El JWT vence a las 24 horas. `POST /api/auth/logout` borra la cookie `sesion` del navegador; un token Bearer ya copiado sigue válido hasta su vencimiento. El cliente debe descartarlo al cerrar sesión.
- Si en una iteración futura se necesita revocación inmediata, habrá que consultar el estado de la sesión o una lista de tokens revocados en cada request protegida, o usar tokens de vida mucho más corta con un mecanismo de renovación. Consultar sólo al iniciar sesión no revoca un JWT ya emitido.
