import { Skeleton } from "@/components/ui/skeleton";

// Esqueleto genérico del panel: título y una lista de tarjetas.
export default function CargandoPanel() {
  return (
    <div aria-busy="true" aria-label="Cargando" className="grid gap-4">
      <Skeleton className="h-8 w-48" />
      {Array.from({ length: 4 }, (_, i) => (
        <Skeleton key={i} className="h-24 w-full" />
      ))}
    </div>
  );
}
