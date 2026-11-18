import { ShoppingBagIcon } from "lucide-react";
import Link from "next/link";
import { Suspense } from "react";

import { buttonVariants } from "@/components/ui/button";

import { ContadorCarrito } from "./contador-carrito";

// Link a /carrito con la cantidad de unidades. Con sesión la cantidad sale de
// la API (servidor); sin sesión, de localStorage (ContadorInvitado). En el
// teléfono se ve el ícono con la cantidad; la palabra queda para el lector
// de pantalla.
export function EnlaceCarrito() {
  return (
    <Link href="/carrito" className={buttonVariants({ variant: "ghost", className: "-mr-2 gap-1.5 px-2 lg:mr-0" })}>
      <ShoppingBagIcon className="size-5" aria-hidden="true" />
      <span className="sr-only sm:not-sr-only">Carrito</span>
      <Suspense fallback={null}>
        <ContadorCarrito />
      </Suspense>
    </Link>
  );
}
