import type { EstadoCarrito } from "./tipos";

// Estado inicial de las acciones del carrito en Client Components. Equivale a
// SIN_ENVIAR de lib/acciones, pero no se importa de ahí: ese módulo trae
// traducirError y con él lib/api.ts (server-only), que no puede llegar al navegador.
export const ESTADO_INICIAL: EstadoCarrito = { enviado: false };
