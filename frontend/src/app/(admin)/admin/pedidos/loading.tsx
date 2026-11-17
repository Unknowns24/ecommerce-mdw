import { Skeleton } from "@/components/ui/skeleton";

// Con la forma del listado: título, filtros y filas de la tabla.
export default function CargandoPedidos() {
  return (
    <div aria-busy="true" aria-label="Cargando pedidos" className="grid gap-6">
      <Skeleton className="h-8 w-40" />
      <div className="grid gap-3 sm:grid-cols-4">
        {Array.from({ length: 4 }, (_, i) => (
          <Skeleton key={i} className="h-14 w-full" />
        ))}
      </div>
      <div className="grid gap-2">
        {Array.from({ length: 8 }, (_, i) => (
          <Skeleton key={i} className="h-10 w-full" />
        ))}
      </div>
    </div>
  );
}
