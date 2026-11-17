import { Skeleton } from "@/components/ui/skeleton";

// Mismo contenedor que page.tsx para que no salte el diseño.
export default function CargandoRetornoPago() {
  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-4" aria-busy="true">
      <span className="sr-only">Cargando…</span>
      <Skeleton className="h-8 w-72" />
      <Skeleton className="h-28 w-full" />
      <Skeleton className="h-12 w-full sm:w-48" />
    </div>
  );
}
