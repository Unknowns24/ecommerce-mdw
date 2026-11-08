import { Skeleton } from "@/components/ui/skeleton";

// Mismo contenedor que page.tsx: título y tres tarjetas con la forma de un pedido del listado.
export default function CargandoMisPedidos() {
  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6" aria-busy="true">
      <span className="sr-only">Cargando tus pedidos…</span>
      <Skeleton className="h-8 w-40" />
      <ul className="flex flex-col gap-3">
        {[0, 1, 2].map((i) => (
          <li key={i}>
            <Skeleton className="h-44 w-full" />
          </li>
        ))}
      </ul>
    </main>
  );
}
