import { Skeleton } from "@/components/ui/skeleton";
import { EsqueletoCarrito } from "@/features/carrito/components/esqueleto-carrito";

// Mismo contenedor que page.tsx para que no salte el diseño cuando llegan los datos.
export default function CargandoCarrito() {
  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
      <Skeleton className="h-8 w-40" />
      <EsqueletoCarrito />
    </main>
  );
}
