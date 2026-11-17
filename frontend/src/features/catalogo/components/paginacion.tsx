import Link from "next/link";

import { Button } from "@/components/ui/button";

import { hrefCatalogo } from "../filtros";
import type { FiltrosCatalogo } from "../tipos";

// La paginación conserva búsqueda, filtros y orden (H04).
export function Paginacion({ filtros, total, porPagina }: { filtros: FiltrosCatalogo; total: number; porPagina: number }) {
  const paginas = Math.max(1, Math.ceil(total / porPagina));
  if (paginas <= 1) return null;
  const actual = Math.min(filtros.pagina, paginas);
  return (
    <nav aria-label="Páginas del catálogo" className="flex items-center justify-between gap-2">
      {actual > 1 ? (
        <Button asChild variant="outline">
          <Link href={hrefCatalogo(filtros, { pagina: actual - 1 })}>Anterior</Link>
        </Button>
      ) : (
        <span />
      )}
      <p className="text-sm text-muted-foreground">
        Página {actual} de {paginas}
      </p>
      {actual < paginas ? (
        <Button asChild variant="outline">
          <Link href={hrefCatalogo(filtros, { pagina: actual + 1 })}>Siguiente</Link>
        </Button>
      ) : (
        <span />
      )}
    </nav>
  );
}
