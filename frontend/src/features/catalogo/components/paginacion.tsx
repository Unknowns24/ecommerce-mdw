import Link from "next/link";

import { Button } from "@/components/ui/button";

import { hrefCatalogo } from "../filtros";
import type { FiltrosCatalogo } from "../tipos";

const claseBoton = "text-xs tracking-[0.14em] uppercase";

// La paginación conserva búsqueda, filtros y orden (H04). Centrada debajo
// de la grilla.
export function Paginacion({ filtros, total, porPagina }: { filtros: FiltrosCatalogo; total: number; porPagina: number }) {
  const paginas = Math.max(1, Math.ceil(total / porPagina));
  if (paginas <= 1) return null;
  const actual = Math.min(filtros.pagina, paginas);
  return (
    <nav aria-label="Páginas del catálogo" className="grid grid-cols-[1fr_auto_1fr] items-center gap-3 pt-6 sm:mx-auto sm:w-full sm:max-w-md">
      {actual > 1 ? (
        <Button asChild variant="outline" className={`${claseBoton} justify-self-start`}>
          <Link href={hrefCatalogo(filtros, { pagina: actual - 1 })}>Anterior</Link>
        </Button>
      ) : (
        <span />
      )}
      <p className="text-xs tracking-[0.14em] text-muted-foreground uppercase">
        Página {actual} de {paginas}
      </p>
      {actual < paginas ? (
        <Button asChild variant="outline" className={`${claseBoton} justify-self-end`}>
          <Link href={hrefCatalogo(filtros, { pagina: actual + 1 })}>Siguiente</Link>
        </Button>
      ) : (
        <span />
      )}
    </nav>
  );
}
