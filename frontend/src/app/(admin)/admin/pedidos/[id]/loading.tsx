import { Skeleton } from "@/components/ui/skeleton";

export default function CargandoPedido() {
  return (
    <div aria-busy="true" aria-label="Cargando el pedido" className="grid gap-6">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-6 w-72" />
      <div className="grid gap-6 md:grid-cols-2">
        <Skeleton className="h-40 w-full" />
        <Skeleton className="h-40 w-full" />
      </div>
      <Skeleton className="h-48 w-full" />
    </div>
  );
}
