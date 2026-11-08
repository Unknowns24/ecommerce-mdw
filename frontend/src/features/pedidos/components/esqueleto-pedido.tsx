import { Skeleton } from "@/components/ui/skeleton";

// Esqueleto con la forma del detalle de un pedido: título, estado, productos y entrega.
export function EsqueletoPedido() {
  return (
    <div className="flex flex-col gap-4" aria-busy="true" aria-live="polite">
      <span className="sr-only">Cargando el pedido…</span>
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-36 w-full" />
      <Skeleton className="h-64 w-full" />
      <Skeleton className="h-28 w-full" />
    </div>
  );
}
