import { Skeleton } from "@/components/ui/skeleton";

// Con la forma del catálogo: banner del título, categorías, barra de filtros
// y la grilla de vitrinas 4:5 (2 columnas en el teléfono, 3 desde sm). No
// hace saltar la página cuando llega.
export function EsqueletoGrilla() {
  return (
    <div aria-busy="true" aria-label="Cargando productos" className="grid gap-6">
      <Skeleton className="h-36 w-full sm:h-56" />
      <div className="flex justify-center gap-7">
        {Array.from({ length: 4 }, (_, i) => (
          <Skeleton key={i} className="h-4 w-16" />
        ))}
      </div>
      <div className="flex items-end justify-between gap-3">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="hidden h-11 w-96 sm:block" />
      </div>
      <ul className="grid grid-cols-2 gap-x-2 gap-y-10 sm:grid-cols-3 sm:gap-x-3">
        {Array.from({ length: 6 }, (_, i) => (
          <li key={i} className="grid justify-items-center gap-2">
            <Skeleton className="aspect-[4/5] w-full" />
            <Skeleton className="h-3 w-1/3" />
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-4 w-1/3" />
          </li>
        ))}
      </ul>
    </div>
  );
}
