import Link from "next/link";
import { Alert } from "@/components/ui/alert";
import { buttonVariants } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { formatearPrecio } from "@/lib/formato";

// Aviso fijo de SPEC-H07: el carrito no reserva ni congela precios.
export function AvisoCarrito({ invitado }: { invitado: boolean }) {
  return (
    <Alert role="note">
      <p className="font-medium">El carrito no reserva unidades ni congela precios.</p>
      <p>
        Las unidades y los precios se confirman al hacer el pedido.
        {invitado ? " La disponibilidad se verifica en el checkout." : null}
      </p>
    </Alert>
  );
}

// Total estimado y continuar. Solo se dibuja con el carrito no vacío.
export function ResumenCarrito({ totalCentavos }: { totalCentavos: number }) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-4">
        <p className="flex items-baseline justify-between gap-2 text-lg font-semibold">
          <span>Total estimado</span>
          <span>{formatearPrecio(totalCentavos)}</span>
        </p>
        <Link href="/checkout" className={buttonVariants({ className: "w-full sm:w-auto sm:self-end" })}>
          Continuar al checkout
        </Link>
      </CardContent>
    </Card>
  );
}
