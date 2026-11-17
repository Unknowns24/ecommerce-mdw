import "server-only";

import { api, ErrorApi } from "@/lib/api";

import type { CategoriaPublica, DetalleVariante, FiltrosCatalogo, MarcaPublica, PaginaVariantes } from "./tipos";

export const POR_PAGINA = 12;

export function listarVariantes(f: FiltrosCatalogo): Promise<PaginaVariantes> {
  const q = new URLSearchParams({ orden: f.orden, pagina: String(f.pagina), porPagina: String(POR_PAGINA) });
  if (f.q) q.set("q", f.q);
  if (f.categoria) q.set("categoria", f.categoria);
  if (f.marca) q.set("marca", f.marca);
  return api<PaginaVariantes>(`/api/catalogo/variantes?${q}`);
}

export async function obtenerVariante(id: string): Promise<DetalleVariante> {
  const v = await api<DetalleVariante>(`/api/catalogo/variantes/${encodeURIComponent(id)}`);
  // Una lista ausente se lee como vacía.
  return { ...v, imagenes: v.imagenes ?? [], categorias: v.categorias ?? [], otrasVariantes: v.otrasVariantes ?? [] };
}

export function listarCategorias(): Promise<CategoriaPublica[]> {
  return api<CategoriaPublica[]>("/api/catalogo/categorias");
}

export function listarMarcas(): Promise<MarcaPublica[]> {
  return api<MarcaPublica[]>("/api/catalogo/marcas");
}

// Condiciones de entrega y pago de la tienda (GET /api/tienda/configuracion,
// público). null si la tienda todavía no cargó configuración (404): la ficha
// no muestra nada inventado.
export type CondicionesEntrega = {
  efectivoHabilitado: boolean;
  distanciaMaximaMetros: number;
  tarifas: { desdeMetros: number; hastaMetros: number; costoCentavos: number }[];
};

export async function obtenerCondicionesEntrega(): Promise<CondicionesEntrega | null> {
  try {
    return await api<CondicionesEntrega>("/api/tienda/configuracion");
  } catch (e) {
    if (e instanceof ErrorApi && e.status === 404) return null;
    throw e;
  }
}
