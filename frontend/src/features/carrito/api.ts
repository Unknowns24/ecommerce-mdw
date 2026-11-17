import "server-only";
import { api, ErrorApi } from "@/lib/api";
import type {
  CarritoApi,
  ItemInvitado,
  VarianteCatalogo,
  VistaCarrito,
} from "./tipos";

// Llamadas a la API de Go del carrito. Solo se usan desde el servidor
// (Server Components y Server Actions de este módulo).

// Tope técnico de líneas que se valorizan por pedido, para que una lista
// enorme no dispare cientos de llamadas a la API. No es una regla de negocio.
export const MAX_LINEAS_INVITADO = 50;

function aVista(c: CarritoApi): VistaCarrito {
  return { lineas: c.items, totalEstimadoCentavos: c.totalEstimadoCentavos };
}

// --- Cliente (con sesión) ---------------------------------------------------

export async function verCarrito(): Promise<VistaCarrito> {
  return aVista(await api<CarritoApi>("/api/carrito"));
}

export async function agregarItem(varianteId: string, unidades: number): Promise<VistaCarrito> {
  return aVista(
    await api<CarritoApi>("/api/carrito/items", { metodo: "POST", cuerpo: { varianteId, unidades } }),
  );
}

export async function cambiarUnidades(itemId: string, unidades: number): Promise<VistaCarrito> {
  return aVista(
    await api<CarritoApi>(`/api/carrito/items/${encodeURIComponent(itemId)}`, {
      metodo: "PATCH",
      cuerpo: { unidades },
    }),
  );
}

export async function quitarItem(itemId: string): Promise<void> {
  await api<void>(`/api/carrito/items/${encodeURIComponent(itemId)}`, { metodo: "DELETE" });
}

export async function vaciarCarrito(): Promise<void> {
  await api<void>("/api/carrito", { metodo: "DELETE" });
}

/** Cantidad de unidades del carrito del cliente (para el encabezado). */
export async function unidadesDelCliente(): Promise<number> {
  const { lineas } = await verCarrito();
  return lineas.reduce((suma, l) => suma + l.unidades, 0);
}

// --- Invitado (sin sesión) --------------------------------------------------

/** GET /api/catalogo/variantes/{id}: público. Lanza ErrorApi (404 si no existe). */
export function leerVariante(varianteId: string): Promise<VarianteCatalogo> {
  return api<VarianteCatalogo>(`/api/catalogo/variantes/${encodeURIComponent(varianteId)}`);
}

/**
 * Valoriza la lista del invitado con el catálogo vigente: nombre, precio y
 * disponibilidad salen siempre de la API. Una variante que ya no existe queda
 * como línea no disponible y sin precio, igual que en el carrito del cliente.
 */
export async function valorizarInvitado(items: ItemInvitado[]): Promise<VistaCarrito> {
  const lineas = await Promise.all(
    items.slice(0, MAX_LINEAS_INVITADO).map(async ({ varianteId, unidades }) => {
      try {
        const v = await leerVariante(varianteId);
        const precio = v.disponible ? v.precioMinoristaCentavos : 0;
        return {
          id: varianteId,
          varianteId,
          codigo: v.codigo,
          nombre: v.nombre,
          unidades,
          precioUnitarioCentavos: precio,
          subtotalCentavos: precio * unidades,
          disponible: v.disponible,
        };
      } catch (e) {
        if (!(e instanceof ErrorApi) || (e.status !== 404 && e.status !== 400)) throw e;
        return {
          id: varianteId,
          varianteId,
          codigo: "",
          nombre: "Producto que ya no está disponible",
          unidades,
          precioUnitarioCentavos: 0,
          subtotalCentavos: 0,
          disponible: false,
        };
      }
    }),
  );
  return {
    lineas,
    totalEstimadoCentavos: lineas.reduce((suma, l) => suma + l.subtotalCentavos, 0),
  };
}
