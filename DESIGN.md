# DESIGN.md — dirección visual de BC Importados

**Estado:** aprobada el 10/11/2026 y **aplicada** en la Etapa E (commits `b45a28b` y `c6fa39a`). Los tokens viven en `frontend/src/app/globals.css`. Medición del costo de las fuentes en `PERFORMANCE_REPORT.md`, Etapa E.

**Límites que respeta:** la spec (sección 7: contraste 4,5:1, foco visible, nada comunicado solo con color, mobile-first a 375px), AGENTS.md («Estilos»: solo Tailwind, tokens en `globals.css`, componentes de `components/ui/`, ningún borde/radio/sombra a mano en `src/app/`) y ADR-007. No se cambia el nombre, no se inventa un logo y se conserva el azul de marca que eligió el equipo (`--primary: #1e3a5f`).

## 1. De qué se trata (investigación)

| Pregunta | Respuesta (de la spec y el catálogo real) |
| --- | --- |
| Qué es | Un negocio de Rosario con local físico que vende importados: perfumes, cremas y body splash (Victoria's Secret, Bath & Body Works) y algo de electrónica (JBL). Hoy vende por Instagram con un catálogo en Drive (spec, sección 1). |
| Quién compra | Clientes minoristas y mayoristas que llegan desde Instagram, casi siempre desde el teléfono, muchos sin cuenta (la compra como invitado es el caso principal). |
| Qué tiene que lograr la pantalla | Encontrar el producto, ver precio y si hay stock, y comprar en pocos pasos con retiro en el local o envío cercano, confiando en que el pago es serio. |
| Diseño actual | Sistema shadcn «nova» casi sin tocar: tipografía del sistema (`system-ui`), grises neutros, azul `#1e3a5f` como único color de marca, tarjetas con borde y sombra suave iguales para todo. Funciona y es accesible en lo esencial, pero es intercambiable con cualquier otra tienda (es el «kit de tarjetas SaaS» que la guía de Frontend Design marca como genérico). |
| Competencia | No hay información del negocio sobre competidores; no se inventa. El vocabulario visual del rubro (etiquetas de perfumería, frascos, mostrador de vidrio) es la referencia. |

## 2. Dirección: «el mostrador del local»

La tienda tiene que sentirse como el mostrador de vidrio de una perfumería de barrio bien atendida: el producto en primer plano sobre una superficie clara, con etiquetas sobrias y precios que se leen de un vistazo. **La audacia se gasta en un solo lugar: el nombre del producto, escrito como en la etiqueta de un perfume** (tipografía Didone). Todo lo demás —navegación, formularios, panel— queda callado y disciplinado.

## 3. Paleta (contrastes medidos con la fórmula WCAG 2.1)

| Nombre | Hex | Rol | Contraste |
| --- | --- | --- | --- |
| Tinta | `#1E3A5F` | Marca, botón principal, enlaces (es el `--primary` actual: se conserva) | 11,50:1 sobre papel; 10,28:1 sobre nácar |
| Papel | `#FFFFFF` | Fondo de página | — |
| Nácar | `#F5F1F4` | Superficies: encabezado, «vitrina» detrás de cada foto, resumen del checkout | — |
| Grafito | `#2B2F36` | Texto principal | 13,44:1 sobre papel |
| Niebla | `#5F6670` | Texto secundario (reemplaza al gris actual, que da 4,35:1 sobre `muted` y no llega al mínimo) | 5,80:1 sobre papel; 5,18:1 sobre nácar |
| Frasco | `#7A2E4A` | Único acento: el precio y el foco del teclado | 9,06:1 sobre papel; 8,10:1 sobre nácar |
| Alerta | `#9F1239` | Errores y 409 (es el `--destructive` actual) | 8,02:1 sobre papel |
| Borde | `#D9D4DA` | Separadores decorativos solamente. Los campos usan Niebla, porque un borde de control necesita 3:1 | 1,46:1 (no apto para controles) |

Sin degradados ni violetas. Sin modo oscuro en esta etapa (no está en el alcance y duplicaría la revisión de contraste).

## 4. Tipografía

| Rol | Familia | Por qué |
| --- | --- | --- |
| Nombre del producto y títulos de la tienda | **Bodoni Moda** (Didone, variable con eje óptico), 500–600 | Es el lenguaje de las etiquetas de perfumería del catálogo. Es la «una cosa memorable». |
| Todo lo demás: interfaz, precios, formularios, panel | **Public Sans** (variable), 400 y 600, con cifras tabulares en precios y tablas | Neutra, muy legible en pantallas chicas, con buen español (tildes, ñ). Elegida en lugar de Inter o la del sistema para no sonar a plantilla. |

- Se cargan con `next/font/google`: se descargan en el build y se sirven desde el mismo dominio (sin pedido a Google en cada visita), con `display: "swap"`, solo `latin` y solo los ejes usados. Reemplazan al `system-ui` actual.
- Escala (base 16px, razón 1,25): 13 · 16 · 20 · 25 · 31 · 39. Interlineado 1,5 en texto y 1,15 en títulos Didone. Líneas de menos de 75 caracteres.
- El panel administrativo usa solo Public Sans: es una herramienta, no una vitrina.
- Nada de mayúsculas sostenidas para etiquetas, salvo «SIN STOCK», que es un texto que la spec exige así (H04).

## 5. Composición y grid

- **Mobile (375px), que es el diseño principal.** Grilla de 4 columnas con 16px de margen y 12px de separación. El catálogo pasa de 1 a **2 productos por fila**: es el patrón de compra desde Instagram y duplica lo visible sin achicar el texto debajo de 16px.
- **Tablet:** 3 columnas. **Escritorio:** 4 columnas y como máximo 1200px de ancho; los filtros pasan a una columna lateral fija a la izquierda.
- Alineación a la izquierda en todo; el centrado queda solo para los estados vacíos.

```
Catálogo, 375px                         Detalle, 375px
┌───────────────────────────┐           ┌───────────────────────────┐
│ BC Importados    🛒 2  ☰  │ nácar      │ ← Catálogo                │
│ [ Buscar productos     ]  │           │ ┌───────────────────────┐ │
├───────────────────────────┤           │ │                       │ │ vitrina nácar
│ Filtros ▾   Orden ▾       │           │ │        (foto)         │ │ 4:5
│ ┌──────────┐ ┌──────────┐ │           │ └───────────────────────┘ │
│ │  (foto)  │ │  (foto)  │ │ 4:5       │ Victoria's Secret         │ niebla
│ │          │ │          │ │           │ Love Spell                │ Bodoni 31
│ └──────────┘ └──────────┘ │           │ crema 236 ml              │
│ Love Spell   Bare Vanilla │ Bodoni    │ $ 15.900,00               │ frasco
│ $ 15.900     $ 8.900      │ frasco    │ Disponible                │
│              SIN STOCK    │           │ Otras presentaciones ...  │
└───────────────────────────┘           ├───────────────────────────┤
                                        │ [−] 1 [+] [Agregar al carrito] │ barra fija abajo
                                        └───────────────────────────┘
```

- **Uso con una sola mano:** en el detalle y en el checkout, la acción principal queda en una barra fija abajo (botón de 48px de alto, a lo ancho). El carrito queda arriba a la derecha con la cantidad en texto.
- **Checkout:** una sola columna con tres bloques numerados porque sí son una secuencia (1 Tus datos, 2 Entrega, 3 Pago) y el total fijo abajo en el teléfono; en escritorio el resumen pasa a la derecha.

## 6. Componentes

| Pieza | Tratamiento |
| --- | --- |
| Tarjeta de producto | Sin borde ni sombra: la foto en su «vitrina» nácar 4:5, debajo la marca (niebla, 13px), el nombre (Bodoni) y el precio (Public Sans 600, color frasco). Las tarjetas se separan por el espacio, no por cajas. |
| Botones | Principal: tinta con texto papel, 44–48px de alto en el teléfono, radio 6px. Secundario: contorno niebla. Destructivo: alerta. El estado pendiente cambia el texto («Agregando…»), como ya pide la spec. |
| Radios | 6px para controles, 0 para las fotos (el producto manda) y 10px solo para los avisos. |
| Sombras | Ninguna, salvo la barra fija inferior (una línea de 1px arriba, color borde). |
| Insignias de estado | Texto siempre visible y fondo tenue; nunca solo color. Se conservan los textos actuales de pedidos y pagos. |
| Navegación del panel | Barra superior simple con los módulos que habilitan los permisos (ya existe); no cambia de estructura. |
| Iconos | Solo `lucide-react`, ya instalado, y solo los funcionales (carrito, buscar, cerrar, menos/más), siempre con texto o `aria-label`. Sin íconos decorativos. |

## 7. Movimiento

- Sin animaciones de entrada por sección ni efectos al pasar el mouse por cada tarjeta.
- Solo movimiento que responde a una acción: el aviso «Agregado al carrito» entra desde abajo en 150ms, y la cantidad del carrito en el encabezado se actualiza con un cambio de opacidad.
- Todo respeta `prefers-reduced-motion`.

## 8. Fotografía

- El catálogo hoy no tiene fotos cargadas (el seed no trae imágenes y el almacenamiento de imágenes está «pendiente» en la spec, sección 8). La vitrina nácar con el texto «Sin imagen» es el estado previsto, no un error.
- Cuando haya fotos: fondo blanco o neutro, producto centrado, proporción 4:5, sin recortes que corten el frasco. Ver la mejora 2 de `AUDIT_REPORT.md` para formatos y tamaños.

## 9. Qué evitamos (y por qué no aplica acá)

- Fondo crema con serif y acento terracota: el nácar es rosado-perla y frío, y el acento es bordó. No usamos `#F4F1EA` ni `#D97757`.
- Kit de tarjetas SaaS (todo en tarjetas iguales con la misma sombra): lo reemplazamos por la vitrina sin borde.
- Rótulos en mayúsculas encima de cada título, numeraciones decorativas «01/02/03» y flechas «→» en los botones.
- Degradados, violetas, brillos y animaciones gratuitas.

## 10. Cómo se verifica

- Las capturas con Playwright a 375, 768 y 1280px de: catálogo, detalle, carrito, checkout, seguimiento, cuenta y panel, comparadas con esta guía.
- El contraste de cada par de tokens, recalculado al aplicarlos, y revisado en `/admin/sistema`.
- El recorrido con teclado de punta a punta, con el foco visible (color frasco, 2px).


## 11. Ampliación del 11/11 (brief de rediseño, adaptado al catálogo real)

El brief pedía una tienda de moda (talles, colores, campañas). BC Importados vende perfumería, cosmética y electrónica: se mantuvo esta dirección (decisión del 11/11) y se sumó solo lo aplicable, **sin inventar datos**:

| Pieza | Dónde | Criterio |
| --- | --- | --- |
| Accesos por categoría | Sobre la grilla del catálogo | Categorías reales de la API; la activa con `aria-current` y subrayado (no solo color); desplazamiento horizontal propio en el teléfono |
| Filtros aplicados | Encima de la grilla | Un chip por filtro, cada uno con su «Quitar …» accesible, más «Quitar todo» |
| Migas de pan | Ficha | Catálogo / categoría / producto |
| Galería | Ficha | Foto principal más miniaturas (botones con `aria-pressed`). Sin zoom ni carrusel: no hay fotos de campaña |
| Secciones plegables | Ficha | `<details>` nativo con `<h2>` en el `summary` (navegable por encabezados) |
| Entrega y pago | Ficha | Solo con `GET /api/tienda/configuracion`: retiro sin cargo, distancia máxima, tarifa mínima, efectivo si está habilitado |
| Footer | Tienda y cuenta | Marca, navegación y «Cómo comprar». **Sin** dirección, teléfono, redes ni políticas: no hay datos cargados |

**No se implementó, por falta de datos (no se inventa):** hero de campaña, fotos de modelos o de categorías, selectores de talle o color, precios anteriores, descuentos, «Nuevo» o «Más vendido», newsletter y páginas de políticas o de contacto.

**Recursos que faltan para el próximo salto visual:**
1. Fotos de producto sobre fondo neutro en 4:5 (mínimo 1200×1500).
2. Un logotipo, si existe.
3. Datos de contacto y horarios del local.
4. Textos de cambios y devoluciones aprobados por el negocio.
