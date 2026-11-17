import { ORDENES, type FiltrosCatalogo, type Orden } from "./tipos";

type Busqueda = Record<string, string | string[] | undefined>;

const uno = (v: string | string[] | undefined) => (Array.isArray(v) ? v[0] : v)?.trim() ?? "";

// La selección vive en la query string: se puede compartir y la paginación
// la conserva.
export function filtrosDe(sp: Busqueda): FiltrosCatalogo {
  const orden = uno(sp.orden);
  const pagina = Number.parseInt(uno(sp.pagina), 10);
  return {
    q: uno(sp.q),
    categoria: uno(sp.categoria),
    marca: uno(sp.marca),
    orden: orden in ORDENES ? (orden as Orden) : "nombre_asc",
    pagina: Number.isFinite(pagina) && pagina > 0 ? pagina : 1,
  };
}

export function hayFiltros(f: FiltrosCatalogo): boolean {
  return Boolean(f.q || f.categoria || f.marca);
}

export function hrefCatalogo(f: FiltrosCatalogo, cambios: Partial<FiltrosCatalogo>): string {
  const n = { ...f, ...cambios };
  const q = new URLSearchParams();
  if (n.q) q.set("q", n.q);
  if (n.categoria) q.set("categoria", n.categoria);
  if (n.marca) q.set("marca", n.marca);
  if (n.orden !== "nombre_asc") q.set("orden", n.orden);
  if (n.pagina > 1) q.set("pagina", String(n.pagina));
  const s = q.toString();
  return s ? `/?${s}` : "/";
}
