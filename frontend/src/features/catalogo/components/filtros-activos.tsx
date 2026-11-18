import Link from "next/link";
import { XIcon } from "lucide-react";

import { hrefCatalogo } from "../filtros";
import type { CategoriaPublica, FiltrosCatalogo, MarcaPublica } from "../tipos";

// Qué está filtrando la grilla y cómo quitarlo, uno por uno.
export function FiltrosActivos({
  filtros,
  categorias,
  marcas,
}: {
  filtros: FiltrosCatalogo;
  categorias: CategoriaPublica[];
  marcas: MarcaPublica[];
}) {
  const nombreCategoria = categorias.flatMap((c) => [c, ...(c.hijas ?? [])]).find((c) => c.id === filtros.categoria)?.nombre;
  const nombreMarca = marcas.find((m) => m.id === filtros.marca)?.nombre;
  const activos = [
    filtros.q ? { texto: `«${filtros.q}»`, href: hrefCatalogo(filtros, { q: "", pagina: 1 }), quitar: "la búsqueda" } : null,
    nombreCategoria ? { texto: nombreCategoria, href: hrefCatalogo(filtros, { categoria: "", pagina: 1 }), quitar: `la categoría ${nombreCategoria}` } : null,
    nombreMarca ? { texto: nombreMarca, href: hrefCatalogo(filtros, { marca: "", pagina: 1 }), quitar: `la marca ${nombreMarca}` } : null,
  ].filter((a): a is { texto: string; href: string; quitar: string } => a !== null);
  if (activos.length === 0) return null;
  return (
    <ul aria-label="Filtros aplicados" className="flex flex-wrap gap-2">
      {activos.map((a) => (
        <li key={a.href}>
          <Link
            href={a.href}
            aria-label={`Quitar ${a.quitar}`}
            className="inline-flex h-9 items-center gap-1.5 bg-secondary px-3 text-xs tracking-[0.1em] text-secondary-foreground uppercase hover:bg-secondary/70"
          >
            {a.texto}
            <XIcon className="size-3.5" aria-hidden="true" />
          </Link>
        </li>
      ))}
      <li>
        <Link href="/" className="inline-flex h-9 items-center px-1 text-xs tracking-[0.1em] uppercase underline underline-offset-4">
          Quitar todo
        </Link>
      </li>
    </ul>
  );
}
