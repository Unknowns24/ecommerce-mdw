import { Skeleton } from "@/components/ui/skeleton";

export default function CargandoCuenta() {
  return (
    <div aria-busy="true" aria-label="Cargando tu cuenta" className="grid max-w-md gap-4">
      <Skeleton className="h-8 w-40" />
      <Skeleton className="h-4 w-56" />
      {Array.from({ length: 3 }, (_, i) => (
        <Skeleton key={i} className="h-16 w-full" />
      ))}
      <Skeleton className="h-10 w-full sm:w-40" />
    </div>
  );
}
