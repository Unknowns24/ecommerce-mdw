import { Skeleton } from "@/components/ui/skeleton";

// Esqueleto con la forma del carrito: aviso, tres líneas y el resumen.
// Lo usan loading.tsx y el carrito del invitado mientras valoriza.
export function EsqueletoCarrito() {
  return (
    <div className="flex flex-col gap-4" aria-busy="true" aria-live="polite">
      <span className="sr-only">Cargando el carrito…</span>
      <Skeleton className="h-16 w-full" />
      <ul className="flex flex-col gap-3">
        {[0, 1, 2].map((i) => (
          <li key={i}>
            <Skeleton className="h-40 w-full" />
          </li>
        ))}
      </ul>
      <Skeleton className="h-28 w-full" />
    </div>
  );
}
