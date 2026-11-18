import Link from "next/link";

import { cn } from "@/lib/utils";

import { hrefCatalogo } from "../filtros";
import type { CategoriaPublica, FiltrosCatalogo } from "../tipos";

// Las categorías como accesos rápidos sobre la grilla (las de la API, no
// inventadas). Conservan búsqueda, marca y orden; vuelven a la página 1. La
// activa se marca con aria-current y con texto subrayado, no solo con color.
// Centradas; si no entran (teléfono), se desplazan de costado. En escritorio
// no se muestran: las categorías ya están en la fila del encabezado.
export function AccesosCategorias({ filtros, categorias }: { filtros: FiltrosCatalogo; categorias: CategoriaPublica[] }) {
  const todas = categorias.flatMap((c) => [c, ...(c.hijas ?? [])]);
  if (todas.length === 0) return null;
  const enlace = (id: string, texto: string) => {
    const activo = filtros.categoria === id;
    return (
      <li key={id || "todas"}>
        <Link
          href={hrefCatalogo(filtros, { categoria: id, pagina: 1 })}
          aria-current={activo ? "page" : undefined}
          className={cn(
            "inline-flex h-11 items-center border-b whitespace-nowrap text-xs tracking-[0.16em] uppercase",
            activo ? "border-foreground font-medium text-foreground" : "border-transparent text-muted-foreground hover:text-foreground",
          )}
        >
          {texto}
        </Link>
      </li>
    );
  };
  return (
    <nav aria-label="Categorías" className="-mx-4 overflow-x-auto px-4 lg:hidden">
      <ul className="mx-auto flex w-max gap-7">
        {enlace("", "Todo")}
        {todas.map((c) => enlace(c.id, c.nombre))}
      </ul>
    </nav>
  );
}
