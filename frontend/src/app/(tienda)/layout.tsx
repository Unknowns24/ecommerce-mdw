import Link from "next/link";
import type { ReactNode } from "react";
import { EnlaceCarrito } from "@/features/carrito/components/enlace-carrito";
import { MenuSesion } from "@/features/identidad/components/menu-sesion";

export default function LayoutTienda({ children }: { children: ReactNode }) {
  return (
    <>
      <header className="bg-muted">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-x-4 gap-y-2 px-4 py-3">
          <Link href="/" className="font-semibold text-primary underline-offset-4 hover:underline">
            BC Importados
          </Link>
          <nav aria-label="Tienda" className="flex flex-wrap items-center gap-1">
            <EnlaceCarrito />
            <MenuSesion />
          </nav>
        </div>
      </header>
      <main className="mx-auto w-full max-w-6xl px-4 py-6">{children}</main>
    </>
  );
}
