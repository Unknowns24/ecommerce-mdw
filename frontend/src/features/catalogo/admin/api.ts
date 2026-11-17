import "server-only";

import { api } from "@/lib/api";

import type { Categoria, Imagen, Marca, Pagina, Producto } from "./tipos";

export const POR_PAGINA_ADMIN = 24;

export function listarProductos(pagina: number) {
  return api<Pagina<Producto>>(`/api/admin/productos?pagina=${pagina}&porPagina=${POR_PAGINA_ADMIN}`);
}

export function obtenerProducto(id: string) {
  return api<Producto>(`/api/admin/productos/${encodeURIComponent(id)}`);
}

export function listarMarcas(pagina = 1, porPagina = 100) {
  return api<Pagina<Marca>>(`/api/admin/marcas?pagina=${pagina}&porPagina=${porPagina}`);
}

export function listarCategorias(pagina = 1, porPagina = 100) {
  return api<Pagina<Categoria>>(`/api/admin/categorias?pagina=${pagina}&porPagina=${porPagina}`);
}

export function imagenesDe(varianteId: string) {
  return api<Imagen[]>(`/api/admin/variantes/${encodeURIComponent(varianteId)}/imagenes`);
}
