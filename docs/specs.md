# Especificación del sistema — BC Importados

**Versión:** 0.1 — borrador para revisión del equipo  
**Fecha:** 18 de septiembre de 2026  
**Proyecto:** Metodologías y Desarrollo Web — UAI

Este documento reúne las decisiones del equipo para la primera versión del e-commerce de BC Importados. Conserva las nueve secciones del ejemplo de la materia. El negocio está funcionando, pero no se realizó un relevamiento completo con sus dueños; por eso describe la propuesta del equipo y no una aprobación del cliente.

Las reglas sin marca especial corresponden a decisiones expresadas durante el relevamiento. Los detalles adicionales propuestos para cerrar comportamientos se identifican como **Propuesta para revisión**. Los asuntos que requieren una decisión se indican como **Pendiente**: no deben resolverse silenciosamente al implementar.

Las historias de la sección 4 usan identificadores provisionales `SPEC-Hxx`. No sustituyen todavía la numeración de las 56 HU originales; su actualización y correspondencia se realizará después de revisar esta spec.

## 1. El problema

**Para quién:** clientes minoristas y mayoristas de BC Importados y personas autorizadas para administrar su negocio. Vende perfumes, productos de Victoria’s Secret, cremas, body splash y electrónicos. Cuenta con un único negocio y un local físico.

**Qué hace hoy sin el sistema:** vende por Instagram, comparte un catálogo en Drive y registra stock y ventas en Excel. No se atribuyen pérdidas, errores ni problemas operativos que no hayan sido comprobados.

**Qué mejora:** incorpora un canal en el que el comprador consulta el catálogo, prepara su carrito y realiza un pedido sin coordinar toda la operación por mensajes. Centraliza catálogo, reglas comerciales, pedidos y existencias; permite reservar unidades durante el pago y registrar salidas por ventas externas. Incluye retiro en el local y envío dentro de una distancia vial configurable desde el negocio.

La compra admite invitados y clientes registrados. El registro es opcional. El sistema conserva carritos y favoritos, aplica promociones y cupones, integra pagos y facturación, y permite gestionar pedidos desde un panel con permisos.

## 2. Roles

| Actor o rol                    | Quién es                              | Acceso y responsabilidades                                                                                                                                |
| ------------------------------ | ------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Visitante / comprador invitado | Persona sin cuenta                    | Consulta catálogo, usa carrito y favoritos locales, compra y consulta un pedido mediante enlace privado. Puede solicitar avisos de reposición por correo. |
| Cliente registrado             | Persona que se registra públicamente  | Compra, conserva carrito y favoritos en la base de datos y consulta sus propios pedidos desde su cuenta.                                                  |
| Usuario administrativo         | Cuenta creada desde la administración | Accede a las funciones habilitadas por la unión de permisos de sus roles. No se obtiene esta condición desde el registro público.                         |
| Dueño                          | Cuenta administrativa con rol dueño   | Tiene todos los permisos y control sobre cuentas administrativas y sus roles, con la excepción de la cuenta inicial protegida.                            |
| Dueño inicial                  | Usuario creado por bootstrap          | Tiene rol dueño y protección especial: otros administradores no pueden modificarlo, desactivarlo ni eliminarlo.                                           |

El visitante y el comprador invitado son actores, no roles que requieran una cuenta. El sistema distingue el origen administrativo o público de una cuenta; el mecanismo de persistencia se concretará en el modelo de datos.

**RBAC:** una cuenta administrativa puede tener uno o varios roles. Los permisos efectivos son la unión sin duplicados de sus permisos. El catálogo de permisos corresponde a acciones implementadas; crear un rol no crea nuevas funcionalidades. Solo el dueño administra cuentas administrativas y sus asignaciones de roles. Las demás tareas pueden delegarse mediante permisos específicos.

**Bootstrap:** el registro del dueño inicial solo está disponible cuando no hay ningún usuario con rol dueño y exige una clave bootstrap/master definida en variables de entorno del servidor. El rol dueño es imborrable. El dueño inicial puede cambiar sus propios datos y contraseña, pero no eliminarse, desactivarse ni quitarse ese rol. La protección se valida también en el backend.

**Propuesta para revisión:** impedir además modificaciones a los permisos del rol dueño; serializar el bootstrap para que solicitudes simultáneas no creen dos dueños iniciales. Los cambios de roles y bloqueos deben afectar también a sesiones ya abiertas, sin permitir continuar operando con permisos revocados.

**Pendiente:** determinar si una cuenta administrativa podrá comprar como cliente o necesitará una cuenta pública separada. No afecta la posibilidad de comprar como invitado.

## 3. Entidades

Modelo conceptual; no impone todavía nombres físicos de tablas ni una implementación ORM.

| Entidad                          | Qué representa                                                                                             | Relaciones principales                                                                 |
| -------------------------------- | ---------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| Usuario                          | Cuenta, datos personales, credenciales, estado, origen y condición de dueño inicial                        | Roles N-N; pedidos 1-N cuando el comprador tiene cuenta; carrito y favoritos.          |
| Rol                              | Conjunto nombrado de permisos                                                                              | Usuarios N-N; permisos N-N.                                                            |
| Permiso                          | Acción administrativa autorizable                                                                          | Roles N-N.                                                                             |
| Producto                         | Agrupador de presentaciones de una misma línea; nombre, descripción, marca, estado y metodología FIFO/LIFO | Variantes 1-N; categorías N-N; marca N-1.                                              |
| Variante                         | Unidad comercial vendible: código, nombre, descripción, precios minorista y mayorista, imágenes y estado   | Producto N-1; lotes 1-N; detalles de pedido, carrito y favoritos.                      |
| Imagen de variante               | Imagen principal o adicional de la presentación                                                            | Variante N-1.                                                                          |
| Categoría                        | Organización del catálogo, con subcategorías                                                               | Categoría padre/hijas; productos N-N.                                                  |
| Marca                            | Marca comercial independiente de las categorías                                                            | Productos 1-N.                                                                         |
| Proveedor                        | Nombre obligatorio; correo, teléfono y web opcionales                                                      | Lotes 1-N.                                                                             |
| Lote                             | Ingreso de unidades de una variante, proveedor, fecha y costo por unidad                                   | Variante N-1; proveedor N-1; movimientos y asignaciones a pedidos.                     |
| Movimiento de stock              | Entrada, ajuste o salida, con unidades, fecha, motivo y responsable/origen                                 | Lote N-1; pedido o salida externa cuando corresponda.                                  |
| Reserva de stock                 | Unidades de lotes comprometidas para un pedido                                                             | Pedido 1-N; lotes N-1; estado y vencimiento cuando corresponda.                        |
| Asignación de lotes              | Detalle de qué unidades y costos de lotes respaldan una venta                                              | Detalle de pedido N-1; lote N-1.                                                       |
| Promoción                        | Regla configurable: tipo de beneficio, condiciones, elegibilidad, prioridad y repetición                   | Productos, variantes y categorías incluidos/excluidos; aplicaciones en pedidos.        |
| Cupón                            | Código, tipo de descuento, valor, mínimo, tope, máximo de usos, acumulación y vigencia opcional            | Usos 1-N.                                                                              |
| Uso de cupón                     | Aplicación del cupón a un pedido, consumida o restituida                                                   | Cupón N-1; pedido con un máximo de un cupón.                                           |
| Carrito / ítem                   | Selección editable antes del pedido                                                                        | Usuario cuando está registrado; variantes y cantidades. Invitados: persistencia local. |
| Favorito                         | Variante guardada por una persona                                                                          | Usuario y variante; para invitados se guarda localmente.                               |
| Pedido / detalle                 | Compra web y copia de comprador, productos, cantidades, precios, descuentos y entrega                      | Usuario opcional; detalles 1-N; pagos, reserva, factura, auditoría.                    |
| Intento de pago                  | Operación de Mercado Pago o registro de cobro en efectivo                                                  | Pedido N-1; identificadores externos cuando existen.                                   |
| Factura                          | Referencia y resultado de emisión en el sistema fiscal externo                                             | Pedido; identificador externo y estado de emisión.                                     |
| Auditoría de pedido              | Registro de cambios y responsables                                                                         | Pedido N-1.                                                                            |
| Suscripción de reposición        | Correo que solicita aviso para una variante                                                                | Variante N-1; no exige usuario.                                                        |
| Configuración de tienda / tarifa | Habilitación de efectivo, ubicación del local, distancia máxima y rangos tarifarios                        | Se utiliza en la cotización y copia de entrega del pedido.                             |
| Registro de venta externa        | Salida de stock por venta realizada por otro canal                                                         | Variantes, lotes, unidades y responsable; excluido de estadísticas online.             |

**Propuesta para revisión:** las variantes heredan marca y categorías del producto; cada categoría tiene como máximo un padre y no admite ciclos. Un producto sin alternativas usa una única variante. Guardar importes en pesos argentinos con precisión decimal, sin cálculos monetarios de coma flotante binaria.

Los mililitros pueden figurar en nombre y descripción; no se exige un atributo estructurado específico. Las características técnicas de electrónicos se informan en la descripción.

## 4. Historias de usuario

Las historias administrativas requieren sesión y el permiso correspondiente, salvo las reservadas explícitamente al dueño. Toda restricción se verifica en backend aunque se invoque la operación sin utilizar la interfaz. En compras, «comprador» incluye invitados y clientes registrados.

### SPEC-H01 — Inicializar el dueño

**Como** responsable de instalación, **quiero** registrar el dueño inicial de manera protegida, **para** habilitar la administración del sistema.

- Sin usuarios con rol dueño y con clave bootstrap válida, se crea e identifica al dueño inicial.
- Con una clave inválida, no se crea la cuenta ni se revela la clave esperada.
- Si ya existe un dueño, no se admite otro registro por bootstrap.
- La cuenta y el rol inicial respetan las protecciones indicadas en la sección 2.

### SPEC-H02 — Registrarse, acceder y gestionar la cuenta

**Como** visitante, **quiero** registrarme opcionalmente, **para** conservar mis preferencias y consultar mis compras.

- El registro público crea una cuenta de cliente sin privilegios administrativos; el correo es único.
- El usuario inicia y cierra sesión; puede consultar y editar sus datos permitidos.
- La recuperación se realiza por correo con un enlace temporal de un solo uso, sin revelar si una cuenta existe.
- El cierre invalida la sesión correspondiente y los pedidos de otros usuarios no son accesibles.
- **Pendiente:** campos exactos de perfil, cambio de correo y duración de enlaces/sesiones. Los datos obligatorios del checkout están definidos en SPEC-H13.

### SPEC-H03 — Administrar cuentas y roles

**Como** dueño, **quiero** crear cuentas administrativas y asignarles roles, **para** delegar la operación.

- Puede invitar por correo a establecer una contraseña, listar cuentas, desactivarlas y reactivarlas.
- Puede crear y modificar roles y asignar varios a una cuenta administrativa.
- La autorización usa la unión de permisos, sin duplicados.
- Otros dueños no pueden modificar ni bloquear al dueño inicial.
- El registro público no permite elegir roles administrativos ni declararse dueño.

### SPEC-H04 — Consultar, buscar y ordenar el catálogo

**Como** visitante, **quiero** explorar los productos, **para** encontrar qué comprar.

- El catálogo público muestra tarjetas por variante, con imagen, nombre y precios aplicables o condiciones de precio.
- Incluye búsqueda por nombre sin distinguir mayúsculas/minúsculas, filtros por categoría y marca y orden de precio ascendente/descendente.
- Búsqueda, filtros y orden pueden combinarse; la paginación conserva la selección.
- Las variantes activas agotadas siguen visibles con «SIN STOCK»; las inactivas no se ofrecen.
- Si no hay resultados se informa claramente.
- **Propuesta para revisión:** ordenar por precio minorista y ofrecer filtro «Solo con stock». El desempate de precios debe ser estable.

### SPEC-H05 — Consultar una variante

**Como** visitante, **quiero** ver el detalle de una presentación, **para** conocerla antes de comprar.

- Se muestran descripción, marca, categorías, imágenes, precios y disponibilidad.
- Hay enlaces a las otras variantes activas del mismo producto.
- Una imagen fallida tiene una alternativa visual y no rompe la página.
- Una variante sin stock no admite agregado al carrito ni compra por encargo.

### SPEC-H06 — Guardar favoritos y recibir avisos de reposición

**Como** visitante o cliente, **quiero** guardar variantes y solicitar avisos de stock, **para** volver a los productos que me interesan.

- Se agregan y quitan favoritos; se guardan localmente para invitados y en la base para clientes registrados.
- Una variante agotada permite suscribirse con correo, sin crear una cuenta.
- El aviso se envía por correo cuando vuelve a existir stock disponible.
- **Propuesta para revisión:** una suscripción activa por correo y variante, un aviso por suscripción y posibilidad de baja. No implica suscripción a publicidad general.

### SPEC-H07 — Gestionar el carrito

**Como** comprador, **quiero** agregar, quitar y cambiar cantidades, **para** preparar mi pedido.

- Las cantidades son enteras positivas y no superan la disponibilidad validada.
- Se muestran variante, cantidad, precio aplicado, descuentos, subtotal y total estimado.
- Cada cambio vuelve a evaluar promociones y cupón; se informa si uno deja de aplicar.
- El carrito se conserva al navegar y recargar: almacenamiento local para invitado y base de datos para cliente.
- El carrito no reserva unidades ni congela precios.
- **Propuesta para revisión:** agregar una variante ya presente suma unidades; al iniciar sesión se unen favoritos y carritos sin exceder disponibilidad, informando ajustes.
- **Pendiente:** plazo de conservación y regla exacta de unión de carritos.

### SPEC-H08 — Gestionar catálogo y proveedores

**Como** administrador autorizado, **quiero** gestionar productos, variantes, imágenes, categorías, marcas y proveedores, **para** mantener el catálogo y los ingresos de mercadería.

- Se crean y editan productos y variantes con sus códigos, descripciones y precios minorista/mayorista.
- Se establece imagen principal y se agregan o quitan imágenes adicionales.
- Se gestionan categorías con subcategorías y asociaciones múltiples de productos.
- Se crean, modifican y desactivan marcas y categorías.
- Se registran proveedores con nombre y datos de contacto opcionales, y se consultan sus lotes.
- Desactivar una oferta no elimina sus registros históricos.
- El stock no se cambia por una edición silenciosa del catálogo; usa lotes o ajustes auditables.

### SPEC-H09 — Ingresar lotes y ajustar existencias

**Como** administrador autorizado, **quiero** ingresar lotes y registrar ajustes, **para** mantener las existencias reales.

- Un ingreso registra variante, proveedor, unidades, costo unitario y fecha.
- Cada ajuste identifica lote, cantidad, motivo y responsable.
- Se consultan movimientos y lotes por variante y proveedor.
- No se permite stock negativo ni consumir unidades comprometidas en otro pedido.
- FIFO/LIFO se selecciona por producto y se aplica a sus variantes.
- **Pendiente:** cómo resolver una pérdida física que afecte unidades ya reservadas; no se liberarán silenciosamente pedidos para permitir el ajuste.

### SPEC-H10 — Registrar una venta externa

**Como** administrador autorizado, **quiero** registrar unidades vendidas fuera de la web, **para** descontarlas del stock compartido.

- Selecciona variantes y cantidades; se validan existencias disponibles.
- Se descuentan lotes conforme a FIFO/LIFO y se generan movimientos con responsable y origen externo.
- No utiliza unidades reservadas para pedidos web.
- No aumenta los indicadores de ventas online ni crea un cobro o factura web.
- Su alcance es registrar la salida de stock, no administrar comercialmente la venta externa.

### SPEC-H11 — Configurar promociones

**Como** administrador con permiso, **quiero** definir reglas comerciales, **para** aplicar ofertas sin modificar el código.

- Puede seleccionar productos completos, variantes y categorías/subcategorías, con exclusiones explícitas.
- Configura tipo de beneficio, cantidades mínimas/máximas cuando correspondan y prioridad.
- Se admiten combinaciones a precio total, precio unitario especial por cantidad y aplicación de lista mayorista.
- La evaluación automática respeta prioridad y no aplica dos promociones a la misma unidad.
- Un combo permite exigir componentes diferentes: por ejemplo una unidad de un conjunto de cremas y otra de un conjunto de body splash.
- **Propuesta para revisión:** repetición configurable dentro del pedido; prioridades únicas para evitar empates. Antes de publicar una regla se validan rangos y componentes.

### SPEC-H12 — Usar y administrar cupones

**Como** comprador, **quiero** ingresar un cupón válido, **para** aplicar su beneficio al pedido.

- Se permite un cupón por pedido, fijo o porcentual, con mínimo de importe y máximo global de usos.
- Los porcentuales tienen tope; la vigencia puede ser opcional.
- El mínimo se evalúa después de promociones, sin envío; el descuento se aplica sobre esa misma base.
- Un cupón no acumulable se rechaza si hay promociones o precios mayoristas aplicados, informando el motivo.
- Al crear el pedido consume un uso; al cancelar lo devuelve una sola vez.
- Un administrador con permiso configura cupones y sus condiciones.
- Dos compras simultáneas no pueden superar el máximo global de usos.

### SPEC-H13 — Cotizar entrega y crear el pedido

**Como** comprador, **quiero** revisar el importe final y confirmar mis datos, **para** iniciar una compra.

- Se solicitan nombre, apellido, correo, teléfono y DNI; domicilio si se elige envío.
- El cliente puede comprar sin cuenta; si tiene sesión, el pedido queda vinculado a ella.
- Se ofrecen retiro y envío según distancia vial máxima y tarifas configuradas, sin exigir pertenecer al municipio de Rosario.
- Efectivo solo aparece para retiro si está habilitado; Mercado Pago es el medio online.
- El backend revalida catálogo, stock, reglas, cupón y total; si algo cambió se informa antes de cobrar.
- El pedido guarda una copia inmutable de comprador, productos, cantidades, precios, descuentos y entrega.
- Un carrito vacío, una dirección no cotizable o unidades insuficientes impiden crear esa compra en las condiciones solicitadas.

### SPEC-H14 — Pagar con Mercado Pago

**Como** comprador, **quiero** pagar online, **para** confirmar mi compra.

- Se crea un pedido pendiente de pago y se reservan unidades por 15 minutos, junto con el uso de cupón si existe.
- Cada intento de pago se relaciona con el pedido y usa el importe calculado por el servidor.
- La confirmación depende de la verificación del proveedor por backend, no de la URL de retorno ni de una declaración del navegador.
- Un pago acreditado confirma el pedido una sola vez y conserva la trazabilidad del pago y del stock.
- Un rechazo informa el resultado y permite reintentar si la reserva sigue vigente y no existe otro pago activo que pueda producir un cobro duplicado.
- Los pagos pendientes no se presentan como acreditados.
- El vencimiento sin pago aprobado conduce a cancelación y libera unidades y cupón; la política de conciliación se especifica en la sección 6.

### SPEC-H15 — Comprar en efectivo con retiro

**Como** comprador, **quiero** elegir efectivo para retirar, **para** pagar en el local.

- Si la opción está habilitada, el pedido nace confirmado con pago pendiente de cobro.
- Se reservan unidades sin vencimiento automático.
- La acción administrativa «Completar pedido» registra conjuntamente cobro y entrega, y solicita la factura.
- Cancelarlo libera unidades y devuelve el uso de cupón.

### SPEC-H16 — Consultar el pedido y recibir su confirmación

**Como** comprador, **quiero** consultar mi pedido, **para** conocer su estado y detalle.

- El invitado accede mediante enlace privado no adivinable, enviado por correo y por WhatsApp al confirmarse el pedido.
- El cliente registrado accede a sus pedidos desde su cuenta y recibe por WhatsApp el aviso de confirmación, sin el enlace privado destinado al invitado.
- La consulta muestra productos, cantidades, precios históricos, descuentos, total, entrega y estados de pedido/pago.
- No se permite consultar pedidos ajenos por modificar un identificador.
- **Propuesta para revisión:** mostrar al invitado el enlace seguro también al crear el pedido para que pueda consultar el pago pendiente. La notificación de confirmación permanece ligada a la confirmación, no al mero inicio del pago.

### SPEC-H17 — Gestionar preparación y entrega

**Como** administrador autorizado, **quiero** consultar pedidos y registrar su avance, **para** gestionar la entrega.

- Consulta listado y detalle con comprador, pago, productos y entrega; puede filtrar por fecha y estados.
- En envío, marca despacho y posteriormente entrega como dos acciones separadas.
- En retiro, completa la entrega; en efectivo esa acción registra también el cobro.
- No cambia productos, cantidades, precios ni domicilio del pedido creado.
- Cada transición se valida y se audita; no puede completarse una compra web pendiente de Mercado Pago.

### SPEC-H18 — Cancelar y registrar el reintegro externo

**Como** administrador autorizado, **quiero** cancelar un pedido a solicitud del comprador por WhatsApp, **para** registrar su resolución.

- El comprador no cancela directamente desde la tienda.
- Cancelar un pedido no pagado libera el stock comprometido y devuelve el uso de cupón.
- Cancelar uno pagado registra que el reintegro se realizó fuera del sistema, por declaración del administrador.
- La acción informa expresamente que no ejecuta una devolución en Mercado Pago.
- Se conserva responsable, fecha y motivo; repetir la operación no devuelve stock ni cupón dos veces.
- **Pendiente:** cancelación después del despacho y/o entrega; no debe reponer unidades sin constatar devolución física. Ver sección 6.

### SPEC-H19 — Emitir y consultar la factura

**Como** comprador, **quiero** consultar la factura de mi compra, **para** disponer de su documento fiscal.

- Se solicita a la API externa al acreditarse Mercado Pago o al completar el retiro en efectivo.
- Se asocia el resultado al pedido y se permite acceso solo al comprador autorizado o la administración.
- **Propuesta para revisión:** errores de emisión quedan pendientes para reintento sin revertir el pago; las repeticiones no emiten documentos duplicados.
- **Pendiente técnico a cargo del equipo:** contrato de API, datos fiscales requeridos, idempotencia y tratamiento documental de cancelaciones.

### SPEC-H20 — Consultar estadísticas online

**Como** administrador autorizado, **quiero** consultar indicadores básicos de la web, **para** conocer su desempeño comercial.

- Selecciona un período y obtiene los indicadores definidos en la sección 6.
- Incluye pedidos web pagados y no cancelados, incluso efectivo ya cobrado.
- Excluye ventas externas, pedidos sin cobrar y envíos de la base monetaria.
- Agrupa variantes en el producto para el ranking de productos; el ranking de usuarios incluye solo clientes registrados.
- Si no hay ventas, muestra cantidades e importes en cero y rankings vacíos, sin errores de división.

### SPEC-H21 — Configurar tienda y consultar auditoría

**Como** administrador autorizado, **quiero** configurar entrega y efectivo y consultar el historial de cambios, **para** controlar la operación.

- Configura ubicación del local, distancia vial máxima, tarifas por rangos y habilitación de efectivo para retiro.
- Cambios de configuración no alteran pedidos ya creados.
- Consulta la auditoría de modificaciones de pedidos, incluidos cambios automáticos.
- **Propuesta para revisión:** tarifas sin huecos ni superposiciones dentro de la cobertura; cada frontera debe pertenecer a un solo rango. Auditoría de solo lectura desde las funciones ordinarias.

## 5. Flujo principal

### Compra web con Mercado Pago

1. El comprador explora el catálogo y consulta una variante y sus alternativas.
2. Agrega cantidades al carrito. El sistema evalúa promociones por prioridad; puede aplicar un cupón compatible.
3. Continúa como invitado o inicia sesión. Completa/verifica sus datos.
4. Elige retiro o envío. Para envío se calcula el recorrido vial, se valida la distancia máxima y se aplica el rango tarifario.
5. Revisa productos, descuentos, entrega y total. Puede volver al carrito antes de crear el pedido.
6. El servidor verifica datos y disponibilidad, congela la información comercial y crea el pedido pendiente. Reserva lotes por FIFO/LIFO y consume el uso del cupón en una operación consistente.
7. Se inicia Mercado Pago. El plazo de reserva del pedido es de 15 minutos.
8. El backend verifica la aprobación, confirma el pedido y consolida la venta sin duplicar efectos. Solicita facturación y las notificaciones acordadas.
9. El invitado consulta con el enlace privado; el cliente registrado, desde su cuenta.
10. La administración prepara el pedido. Si hay envío, registra despacho y luego entrega; si hay retiro, completa la entrega.

### Alternativa: efectivo con retiro

1. Se realizan los pasos de catálogo, carrito, datos y validación.
2. Si el efectivo está habilitado y se eligió retiro, el pedido nace confirmado y sin cobrar, con unidades reservadas sin vencimiento automático.
3. Cuando el comprador paga y retira, el administrador completa el pedido con una acción única. Se registran cobro, entrega, consumo de unidades y solicitud de factura.
4. Si no retira, la reserva continúa hasta una cancelación administrativa; no aplica el plazo de Mercado Pago.

### Excepciones principales

- Falta de stock o cambio de precio: se informa antes de pagar y se solicita revisar el resumen; no se cobra un importe modificado sin conocimiento del comprador.
- Fallo al cotizar envío: no se inventa un costo ni se asume envío gratis; se puede reintentar o elegir retiro.
- Fallo al iniciar Mercado Pago: no se marca pagado; se permite reintento dentro del plazo o liberación por cancelación/vencimiento.
- Pago en proceso: se muestra como pendiente, sujeto a conciliación.
- Aprobación tardía: no se confirma sin verificar nuevamente la disponibilidad si la reserva fue liberada.
- Cancelación: la administración registra la resolución y sus consecuencias; no modifica retroactivamente el detalle del pedido.

## 6. Reglas de negocio

### Identidad y autorización

- El registro público nunca concede privilegios administrativos. La clave bootstrap se verifica únicamente en el servidor, sin exponerse al navegador ni a registros de diagnóstico.
- Solo el dueño administra cuentas administrativas y roles. Otros permisos permiten delegar catálogo, stock, promociones, pedidos, configuración y estadísticas según lo asignado.
- El dueño inicial y el rol dueño mantienen las protecciones de la sección 2.
- El enlace de invitado habilita acceso únicamente al pedido asociado. **Pendiente:** vigencia, revocación y recuperación de ese enlace.

### Catálogo y existencias

- Sin stock disponible, la variante permanece visible como «SIN STOCK», pero no se compra ni se encarga.
- Desactivar no equivale a agotar: una variante inactiva no se ofrece aunque conserve existencias.
- Las unidades son enteras; stock disponible = existencias de lotes menos reservas activas.
- FIFO asigna primero lotes ingresados antes; LIFO asigna primero los más recientes. Una compra puede consumir varios lotes.
- **Propuesta para revisión:** desempatar lotes por identificador estable y aplicar cambios de metodología solo a nuevas asignaciones, sin alterar reservas ni ventas anteriores.
- La reserva se asigna a lotes concretos; no puede venderse por otro canal una unidad reservada.
- La liberación de reserva no es una entrada física. Convertir una reserva en venta no descuenta dos veces la unidad. **Propuesta de implementación conceptual:** al reservar baja la disponibilidad; al consolidar la venta se registra la salida y se cierra la reserva asociada. Si se cancela antes de entregar, se revierte lo que efectivamente se hubiera registrado, una sola vez.
- Los costos de lotes permiten trazabilidad histórica; reportes de rentabilidad no forman parte del panel inicial.

### Motor de promociones

- No existe un umbral mayorista universal de cinco unidades: los umbrales y productos se configuran por regla.
- Se admiten tres tipos acotados: combinación con precio total, precio unitario especial por tramo y lista mayorista de cada variante al alcanzar la cantidad elegible.
- En combos deben poder expresarse componentes diferenciados y sus cantidades; «dos unidades de cualquier producto» no sustituye «una crema y un body splash».
- Los conjuntos elegibles admiten selección por producto, variante o categoría y exclusiones. **Propuesta para revisión:** inclusión por unión de selecciones y exclusión con precedencia; una categoría incluye sus descendientes.
- Las reglas se aplican de mayor a menor prioridad. Cada unidad se usa en una única promoción; cantidades restantes pueden participar en otra.
- La condición de cantidad solo cuenta unidades elegibles. El sistema informa el precio resultante al cambiar cantidades.
- **Pendiente:** definir si los tramos se evalúan sobre el total elegible original o sobre unidades todavía libres tras promociones superiores, y cómo elegir unidades cuando varias combinaciones cumplen una regla. Propuesta: evaluar sobre unidades libres, usar un orden estable y no prometer optimización del descuento.
- **Propuesta para revisión:** repetición configurable de combos; los máximos de tramo delimitan elegibilidad y no son automáticamente un límite de usos del combo. Son parámetros distintos.
- Los importes $59.900, $29.900 y $25.900 mencionados en el relevamiento son ejemplos de configuración, no precios fijos codificados en el sistema.

### Cupones y cálculo de importes

1. Subtotal base = suma de cantidad × precio minorista.
2. Se aplican promociones según prioridad, incluyendo lista mayorista cuando corresponda.
3. El subtotal resultante se usa para verificar el mínimo del cupón.
4. Se calcula descuento fijo o porcentual, limitado por el tope aplicable y por el subtotal.
5. Total = subtotal después de promociones − descuento de cupón + envío.

- El envío no participa del mínimo ni recibe descuento de cupón.
- Se admite un cupón por pedido; un cupón no acumulable se rechaza si existen beneficios incompatibles.
- El uso se consume al crear el pedido y no vuelve a consumirse en confirmaciones, reintentos ni finalización. Se restituye solo por cancelación.
- La validación del cupo es atómica para que dos pedidos no consuman el último uso simultáneamente.
- **Propuesta para revisión:** códigos únicos sin distinción de mayúsculas; reglas de redondeo decimal únicas y reparto de descuentos por detalle consistente con el total. El contrato fiscal deberá respetar ese reparto.

### Pedido, pago y cumplimiento

Los estados del pedido, del pago y de la factura se distinguen. «Confirmado» no implica siempre «Pagado»: el retiro en efectivo nace confirmado sin cobro.

**Propuesta de estados para revisión:**

| Dimensión   | Estados                                                                                                       |
| ----------- | ------------------------------------------------------------------------------------------------------------- |
| Pedido      | Pendiente de pago, Confirmado, Despachado, Completado, Cancelado.                                             |
| Pago        | Pendiente, Aprobado/cobrado, Rechazado; Reembolsado por declaración administrativa. Se conserva cada intento. |
| Facturación | Pendiente de emisión, Emitida, Error de emisión; estados adicionales según API externa.                       |

| Operación                   | Precondición                                                  | Resultado                                      |
| --------------------------- | ------------------------------------------------------------- | ---------------------------------------------- |
| Aprobación MP               | Pedido pendiente y disponibilidad comprometida válida         | Confirmado y pago aprobado.                    |
| Crear retiro en efectivo    | Opción habilitada y stock disponible                          | Confirmado y pago pendiente.                   |
| Despachar                   | Envío confirmado con pago aprobado                            | Despachado.                                    |
| Entregar envío              | Despachado                                                    | Completado.                                    |
| Completar retiro MP         | Retiro confirmado y pagado                                    | Completado.                                    |
| Completar retiro efectivo   | Retiro confirmado y no cancelado                              | Completado y pago cobrado; factura solicitada. |
| Vencer compra MP sin pago   | Plazo agotado y resultado conciliado según política pendiente | Cancelado; liberación de stock y cupón.        |
| Cancelar por administración | Estado permitido por política de cancelación                  | Cancelado; consecuencias auditadas.            |

- Los pedidos MP reservan por 15 minutos desde su creación. Reintentar el pago no reinicia indefinidamente el plazo. **Propuesta para revisión:** conservar el vencimiento original en todos los intentos del mismo pedido.
- Los pedidos en efectivo no vencen automáticamente. Deshabilitar efectivo solo afecta compras nuevas, no anula pedidos existentes.
- Productos, cantidades, precios, descuentos, comprador y domicilio del pedido se conservan como copia histórica. Correcciones operativas no alteran silenciosamente lo comprado.
- La misma notificación o acción repetida no duplica pedidos, cobros registrados, movimientos, usos de cupón ni solicitudes fiscales.
- **Pendiente crítico de integración:** medios MP admitidos, vigencia de la operación externa, consulta al vencer y notificaciones tardías. Una preferencia vencida o un temporizador local no bastan para asumir ausencia de pago.
- **Propuesta de contingencia:** si llega una aprobación después de liberar unidades, detener la confirmación automática y registrar una incidencia. Verificar disponibilidad y resolver de forma controlada; si no puede cumplirse, tramitar cancelación/reintegro externo. El pedido nunca se completa con stock negativo ni se informa reembolsado sin intervención administrativa.

### Cancelaciones y auditoría

- El comprador solicita cancelar por WhatsApp; solo la administración con permiso persiste la decisión.
- Para un pedido pagado, cancelar implica declarar reintegro efectuado fuera del sistema. La interfaz debe explicitarlo; no se llama a MP para devolver dinero.
- **Propuesta para revisión:** exigir motivo y confirmación explícita de reintegro; no permitir cancelación ordinaria después del despacho o entrega hasta definir recuperación física y tratamiento fiscal. Este límite sigue pendiente de aceptación del equipo.
- No se reponen unidades entregadas automáticamente por el mero cambio a cancelado. La restitución usa los lotes realmente afectados y es idempotente.
- Se auditan cambios con actor humano o sistema, fecha, acción y datos anteriores/nuevos. La declaración de reembolso queda distinguida de un reintegro verificado por un proveedor.
- Cancelar no elimina la factura; el tratamiento documental de la cancelación corresponde a la integración fiscal pendiente.

### Entrega

- La cobertura depende únicamente de la distancia vial máxima desde el local; no se exige pertenecer administrativamente a Rosario.
- Se aplica la tarifa del rango de distancia correspondiente. El retiro no tiene cargo de envío en esta propuesta.
- **Propuesta para revisión:** usar metros para evaluar rangos sin redondear antes de tarifar; intervalos sin solapamientos, con tratamiento inequívoco de límites. Configuración incompleta o ruta no disponible bloquea la opción de envío, no la compra con retiro.
- El servicio de mapas debe localizar el domicilio y devolver distancia de recorrido vial. Proveedor pendiente: Nominatim puede resolver direcciones, pero requiere un motor de rutas adicional; Google Routes es otra alternativa.
- Distancia, tarifa y domicilio se guardan en el pedido. Cambiar tarifas, origen o cobertura no recalcula compras anteriores.
- El sistema no asigna repartidores ni realiza seguimiento GPS; registra despacho y entrega manualmente.

### Estadísticas

Todos los indicadores admiten un período. Se consideran exclusivamente pedidos originados en la web, pagados y actualmente no cancelados. Se incluyen invitados y efectivo cobrado al retirar. Se excluyen ventas externas y pedidos en efectivo aún sin cobrar. La fecha de atribución es la acreditación o cobro.

| Indicador                       | Cálculo                                                                                                                                  |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Total vendido en la web         | Suma de subtotales de productos después de promociones y cupones, sin envío.                                                             |
| Unidades vendidas               | Suma de cantidades de esos pedidos.                                                                                                      |
| Ticket promedio                 | Total vendido / cantidad de pedidos considerados.                                                                                        |
| Cantidad promedio de productos  | Unidades vendidas / cantidad de pedidos considerados; cuenta unidades, no referencias distintas.                                         |
| Evolución de ventas             | Total vendido agrupado por día, semana o mes, en gráfico de barras o línea.                                                              |
| Diez productos más vendidos     | Orden descendente por unidades; se suman todas las variantes de cada producto.                                                           |
| Diez usuarios que más compraron | Clientes registrados ordenados por importe de productos comprados después de descuentos y sin envío. Invitados excluidos de este ranking. |

El nombre del indicador es «Total vendido», no «Total facturado», porque su base son pedidos pagados, no documentos fiscales. No se agrupan invitados como usuarios por coincidencia de correo o DNI. Cancelar una venta la excluye de estos indicadores, incluso al consultar su período original; no se presenta este panel como un libro contable de reintegros.

**Propuesta para revisión:** zona horaria comercial Argentina y desempates estables en rankings. Si no hay pedidos, promedios en cero.

### Qué pasa al borrar

- Productos, variantes y usuarios vinculados a pedidos, movimientos o facturas no se borran en cascada. Se conserva el historial.
- Productos y variantes pueden desactivarse y reactivarse; las compras históricas mantienen su copia comercial.
- El rol dueño y la cuenta inicial tienen las restricciones de protección indicadas.
- Pedidos se cancelan, no se eliminan; movimientos y auditoría no se editan o eliminan desde operaciones ordinarias. **Propuesta:** errores de movimientos se corrigen con movimientos compensatorios.
- **Propuesta para revisión:** proveedores con lotes se conservan y se desactivan; categorías/marcas con productos activos no se desactivan hasta reasignarlos o desactivar esos productos. No borrar roles asignados sin reasignar cuentas.
- Quitar un favorito o una línea de carrito sí elimina esa selección editable y no modifica historial de ventas.

### Qué se calcula y qué se guarda

| Información            | Tratamiento                                                                                                                     |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| Disponibilidad         | Se deriva de existencias de lotes y reservas activas; una caché eventual no reemplaza la validación transaccional.              |
| Totales del carrito    | Se recalculan con catálogo y reglas vigentes; no garantizan precio ni unidades.                                                 |
| Pedido creado          | Guarda copia de comprador, variantes, nombres/códigos, cantidades, precios, descuentos, reglas aplicadas, cupón, envío y total. |
| Reserva                | Guarda lotes, unidades, estado y vencimiento MP; efectivo sin vencimiento automático.                                           |
| Costos de venta        | Guarda distribución de cantidades por lote y costo unitario histórico.                                                          |
| Estados de pedido/pago | Se persisten a partir de operaciones verificadas, con auditoría y referencias externas.                                         |
| Factura                | Guarda referencia y resultado externo; la API fiscal define el documento y su tratamiento.                                      |
| Estadísticas           | Se calculan desde pedidos elegibles y sus copias históricas; no desde precios actuales del catálogo.                            |
| Permisos               | Se derivan de roles actuales; cambios requieren invalidar o actualizar cualquier caché de autorización.                         |

## 7. Requisitos no funcionales

### Usabilidad

- La tienda y el checkout funcionan en celular, tablet y computadora sin desplazamientos horizontales inesperados.
- Los errores identifican el campo y el motivo sin perder datos ya ingresados.
- Se distinguen claramente precio minorista, condición promocional, descuentos, envío y total a pagar.
- Se muestran estados de carga y se evita el envío accidental repetido de formularios. La protección visual no sustituye la idempotencia del backend.
- El comprador conoce cuándo su pago sigue pendiente y cuándo su pedido está confirmado.
- Las acciones sensibles, como cancelar un pedido pagado, explican sus consecuencias.
- **Propuesta de validación:** probar compra como invitado, compra registrada, retiro efectivo y gestión de pedido con una persona ajena al equipo antes de la entrega; registrar dificultades y corregir bloqueos. No se afirma una prueba con el dueño que aún no está acordada.

### Accesibilidad

- Todo se puede operar con teclado y el foco es visible.
- Los campos tienen etiquetas asociadas, no solo placeholder.
- Las imágenes informativas tienen texto alternativo; las decorativas, alternativo vacío.
- El contraste entre texto y fondo alcanza 4,5:1 y 3:1 para texto grande.
- Los errores no se comunican solo mediante color: incluyen texto.

### Seguridad e integridad

- Contraseñas almacenadas mediante hash seguro; secretos de bootstrap e integraciones solo en servidor.
- Autorización en backend por operación y recurso; no basta ocultar controles.
- No se exponen credenciales, datos sensibles ni trazas internas en mensajes públicos.
- La tienda no almacena números completos de tarjeta ni códigos de seguridad.
- Unidades, cupones y transiciones se validan de forma consistente bajo concurrencia.
- Los enlaces privados requieren tokens no adivinables; no se registra su contenido sensible en logs públicos.
- **Pendiente técnico:** políticas concretas de sesión, tokens, limitación de intentos y conservación de datos.

### Fiabilidad y operación

- Una falla de correo o WhatsApp no revierte la compra; se conserva el resultado comercial.
- Las integraciones tienen estados de error trazables y reintentos que no duplican efectos.
- La compra pendiente de confirmación externa no se presenta como pagada por un timeout.
- **Pendiente del equipo:** objetivos medibles de rendimiento y disponibilidad, volumen de prueba, copias de seguridad/restauración y política operativa de reintentos. No se inventan tiempos de respuesta o cargas no acordadas.

## 8. Integración externa

| Integración                                                | Para qué                                                                                             | Comportamiento ante fallos                                                                                                                                |
| ---------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Mercado Pago                                               | Cobro online, consulta y notificación de resultados                                                  | Mantener trazabilidad del intento; no confirmar desde el navegador. Reconciliar pendientes y aprobaciones tardías conforme a política pendiente.          |
| Sistema fiscal existente del equipo                        | Emitir factura al cobrar MP o completar efectivo; obtener referencia/documento                       | Propuesta: registrar emisión pendiente/error y reintentar sin duplicar ni deshacer una compra pagada. Contrato y ajustes documentales a cargo del equipo. |
| OpenWA en backend                                          | WhatsApp de confirmación: enlace privado para invitado; aviso sin ese enlace para cliente registrado | Propuesta: registrar fallo y reintentar sin cambiar el pedido ni duplicar mensajes. No enviar otras campañas o avisos por esta integración en v1.         |
| Servicio de correo, proveedor pendiente                    | Invitaciones, recuperación de contraseña, enlace privado de invitado y reposición de stock           | Propuesta: reintentar y registrar resultado; evitar tokens vencidos y duplicación de avisos.                                                              |
| Geocodificación y rutas, proveedor pendiente               | Ubicar domicilio y calcular distancia vial desde el local                                            | No cotizar con una distancia inventada. Permitir reintentar o seleccionar retiro.                                                                         |
| Almacenamiento de imágenes/documentos, mecanismo pendiente | Servir imágenes y, según contrato fiscal, documentos                                                 | Imágenes fallidas tienen alternativa visual. El error no elimina el producto ni modifica su stock.                                                        |

La integración fiscal y sus detalles técnicos quedan expresamente a cargo de la usuaria/equipo. No se presupone que la API emita ajustes, admita idempotencia o acepte determinados datos sin verificar su contrato.

Referencias técnicas de mapas consultadas para distinguir servicios: [API de Nominatim](https://nominatim.org/release-docs/latest/api/Overview/) y [Google Routes](https://developers.google.com/maps/documentation/routes/compute_route_directions). La elección de proveedor no está cerrada.

## 9. Fuera de alcance

- Reseñas de productos.
- Compra por encargo o venta de unidades sin stock disponible.
- Múltiples negocios, sucursales o depósitos con stock separado.
- Órdenes de compra a proveedores, cuentas corrientes y pagos a proveedores. Sí se incluyen proveedor básico y lotes.
- Contabilidad, rentabilidad y estadísticas avanzadas por proveedor/lote; los costos se guardan sin incorporar esos reportes al panel inicial.
- Gestión comercial, cobro, facturación y estadísticas de ventas externas: su registro en esta versión solo descuenta stock.
- Edición de productos, cantidades, precios o domicilio de un pedido ya creado.
- Cancelación directa del comprador desde la web.
- Reintegros automáticos desde el sistema.
- Gestión de repartidores, planificación logística y seguimiento GPS.
- Métodos de rotación distintos de FIFO y LIFO.
- Repetir pedidos con un botón: figuraba en las HU originales, pero no quedó confirmado en el nuevo alcance.
- App nativa y múltiples idiomas, como propuesta de límite para esta entrega.

### RFC futuros registrados

| RFC                          | Propuesta futura                                          | Aspectos por analizar                                                                                                |
| ---------------------------- | --------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| RFC-001 — Carrito abandonado  | Contactar a clientes registrados para recordar su carrito | Consentimiento/preferencias, canal, plazo, frecuencia y condiciones de no envío. No activado en v1.                  |
| RFC-002 — Transporte nacional | Cotizar y operar con un transportista de larga distancia  | Comparar cobertura, tarifas, acceso a API y operación de Andreani, Correo Argentino u OCA. Ningún proveedor elegido. |
| RFC-003 — Reintegro integrado | Ejecutar y verificar devoluciones mediante Mercado Pago   | Permisos, idempotencia, estados pendientes/fallidos, restitución de stock y documentos fiscales.                     |

### Puntos a resolver durante la revisión, sin ampliar el alcance

1. Estados definitivos, cancelación después de despacho/entrega y conciliación de pagos tardíos.
2. Repetición de promociones, desempates, selección de unidades y evaluación de tramos.
3. Proveedor de rutas, modo de recorrido y validación de direcciones/rangos.
4. Contrato fiscal y tratamiento documental de cancelaciones; política de reintentos.
5. Duración de carritos/enlaces, combinación de carritos y comportamiento de cuentas administrativas como compradores.
6. Reglas de desactivación de marcas/categorías/proveedores y ajuste de stock reservado.
7. Umbrales operativos de rendimiento, seguridad y recuperación.

Estos puntos están identificados como pendientes o propuestas en sus secciones correspondientes. No invalidan las decisiones confirmadas ni autorizan agregar funcionalidades ajenas a esta versión.
