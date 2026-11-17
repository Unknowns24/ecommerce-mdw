import { Skeleton } from "@/components/ui/skeleton";

// Esqueleto con la forma del detalle: imagen a la izquierda, datos a la derecha.
export default function CargandoVariante() {
  return (
    <div aria-busy="true" aria-label="Cargando el producto" className="grid gap-6 md:grid-cols-2">
      <Skeleton className="aspect-square w-full" />
      <div className="grid content-start gap-4">
        <Skeleton className="h-4 w-1/4" />
        <Skeleton className="h-8 w-3/4" />
        <Skeleton className="h-10 w-1/3" />
        <Skeleton className="h-10 w-full sm:w-48" />
        <Skeleton className="h-24 w-full" />
      </div>
    </div>
  );
}
