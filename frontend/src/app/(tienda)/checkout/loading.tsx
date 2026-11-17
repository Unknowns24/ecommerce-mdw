import { Skeleton } from "@/components/ui/skeleton";
import { EsqueletoCheckout } from "@/features/pedidos/components/esqueleto-checkout";

// Mismo contenedor que page.tsx para que no salte el diseño cuando llegan los datos.
export default function CargandoCheckout() {
  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
      <Skeleton className="h-8 w-40" />
      <EsqueletoCheckout />
    </div>
  );
}
