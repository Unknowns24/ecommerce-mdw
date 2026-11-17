import { Skeleton } from "@/components/ui/skeleton";

// Con la forma del catálogo: título, filtros y la grilla de vitrinas 4:5
// (2 columnas en el teléfono). No hace saltar la página cuando llega.
export function EsqueletoGrilla() {
  return (
    <div aria-busy="true" aria-label="Cargando productos" className="grid gap-6 lg:grid-cols-[15rem_1fr] lg:gap-10">
      <div className="grid content-start gap-3">
        <Skeleton className="h-9 w-40" />
        <Skeleton className="h-11 w-full" />
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-1">
          <Skeleton className="h-11 w-full" />
          <Skeleton className="h-11 w-full" />
        </div>
      </div>
      <ul className="grid grid-cols-2 gap-x-3 gap-y-8 sm:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4">
        {Array.from({ length: 8 }, (_, i) => (
          <li key={i} className="grid gap-2">
            <Skeleton className="aspect-[4/5] w-full rounded-none" />
            <Skeleton className="h-3 w-1/3" />
            <Skeleton className="h-5 w-4/5" />
            <Skeleton className="h-4 w-1/2" />
          </li>
        ))}
      </ul>
    </div>
  );
}
