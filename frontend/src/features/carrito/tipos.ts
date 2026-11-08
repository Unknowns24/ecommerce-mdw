import type { EstadoFormulario } from "@/lib/acciones";

// Formas derivadas de docs/api/pedidos.http (sección 5) y de
// GET /api/catalogo/variantes/{id} (docs/api/catalogo.http).

/** Lo único que el invitado guarda en localStorage. Nunca precios ni nombres. */
export type ItemInvitado = { varianteId: string; unidades: number };

/** Línea de GET /api/carrito (LineaCarritoResponse). */
export type LineaApi = {
  id: string; // id de la línea (item), distinto del de la variante
  varianteId: string;
  codigo: string;
  nombre: string;
  unidades: number;
  precioUnitarioCentavos: number;
  subtotalCentavos: number;
  disponible: boolean;
};

/** Respuesta de GET/POST/PATCH /api/carrito (CarritoResponse). */
export type CarritoApi = { items: LineaApi[]; totalEstimadoCentavos: number };

/** Campos que el carrito usa de GET /api/catalogo/variantes/{id}. */
export type VarianteCatalogo = {
  id: string;
  codigo: string;
  nombre: string;
  precioMinoristaCentavos: number;
  disponible: boolean;
};

/**
 * Línea lista para dibujar, igual para los dos compradores.
 * `id` es el de la línea para el cliente y el `varianteId` para el invitado
 * (el invitado no tiene líneas en el servidor).
 */
export type LineaCarrito = {
  id: string;
  varianteId: string;
  codigo: string;
  nombre: string;
  unidades: number;
  precioUnitarioCentavos: number;
  subtotalCentavos: number;
  disponible: boolean;
};

export type VistaCarrito = { lineas: LineaCarrito[]; totalEstimadoCentavos: number };

/**
 * Estado de las acciones del carrito: el de toda acción más `modo`.
 * `modo: "invitado"` le dice al Client Component que escriba en localStorage
 * `unidades` de la variante (el servidor ya comprobó que existe y tiene stock).
 */
export type EstadoCarrito = EstadoFormulario & {
  modo?: "cliente" | "invitado";
  unidades?: number;
};
