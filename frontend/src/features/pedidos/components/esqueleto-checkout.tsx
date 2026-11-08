import { Skeleton } from "@/components/ui/skeleton";

// Esqueleto con la forma del checkout: tres bloques de campos, el resumen y el botón.
export function EsqueletoCheckout() {
  return (
    <div className="flex flex-col gap-6" aria-busy="true" aria-live="polite">
      <span className="sr-only">Cargando el checkout…</span>
      <Skeleton className="h-72 w-full" />
      <Skeleton className="h-40 w-full" />
      <Skeleton className="h-24 w-full" />
      <Skeleton className="h-48 w-full" />
      <Skeleton className="h-12 w-full sm:w-48 sm:self-end" />
    </div>
  );
}
