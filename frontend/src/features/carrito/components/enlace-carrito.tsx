import Link from "next/link";
import { Suspense } from "react";
import { ContadorCarrito } from "./contador-carrito";

// Link a /carrito con la cantidad de unidades. Con sesión la cantidad sale de
// la API (servidor); sin sesión, de localStorage (ContadorInvitado).
export function EnlaceCarrito() {
  return (
    <Link href="/carrito" className="inline-flex items-center gap-1">
      Carrito
      <Suspense fallback={null}>
        <ContadorCarrito />
      </Suspense>
    </Link>
  );
}
