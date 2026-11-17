import "server-only";

import { api, ErrorApi } from "@/lib/api";

import type { Movimiento, Proveedor, StockVariante } from "./tipos";

export async function listarProveedores(): Promise<Proveedor[]> {
  return (await api<{ items: Proveedor[] }>("/api/admin/proveedores?pagina=1&porPagina=100")).items ?? [];
}

export function stockDe(varianteId: string) {
  return api<StockVariante>(`/api/admin/variantes/${encodeURIComponent(varianteId)}/stock`);
}

export async function movimientosDe(varianteId: string): Promise<Movimiento[]> {
  return (await api<{ items: Movimiento[] }>(`/api/admin/variantes/${encodeURIComponent(varianteId)}/movimientos?pagina=1&porPagina=100`)).items ?? [];
}

// El nombre de la variante sale del catálogo público (stock.gestionar no
// incluye leer el catálogo administrativo). Una variante inactiva no está
// ahí: se muestra su código de id.
export async function nombreDeVariante(varianteId: string): Promise<string | null> {
  try {
    const v = await api<{ nombre: string }>(`/api/catalogo/variantes/${encodeURIComponent(varianteId)}`);
    return v.nombre;
  } catch (e) {
    if (e instanceof ErrorApi && (e.status === 404 || e.status === 400)) return null;
    throw e;
  }
}

export async function variantesParaElegir(): Promise<{ id: string; nombre: string; codigo: string }[]> {
  const r = await api<{ items: { id: string; nombre: string; codigo: string }[] }>("/api/catalogo/variantes?pagina=1&porPagina=100&orden=nombre_asc");
  return r.items ?? [];
}

// <input type="date"> → instante en la zona comercial (ART, -03:00).
export function fechaDeIngreso(valor: string | undefined): string {
  return valor && /^\d{4}-\d{2}-\d{2}$/.test(valor) ? `${valor}T12:00:00-03:00` : "";
}
