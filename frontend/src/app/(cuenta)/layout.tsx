import Link from "next/link";
import type { ReactNode } from "react";

import { EncabezadoTienda } from "@/components/encabezado-tienda";
import { PieTienda } from "@/components/pie-tienda";
import { buttonVariants } from "@/components/ui/button";
import { requerirSesion } from "@/features/identidad/sesion";

// Grupo (cuenta): el encabezado de la tienda más el menú de la cuenta. Sin
// sesión, /ingresar. Es experiencia de uso, no autorización: cada página igual
// llama a la API con la sesión, y la API decide.
export default async function LayoutCuenta({ children }: { children: ReactNode }) {
  await requerirSesion();
  return (
    <>
      <EncabezadoTienda />
      <div className="mx-auto grid w-full max-w-6xl gap-4 px-4 py-6">
        <nav aria-label="Tu cuenta" className="flex flex-wrap gap-1 border-b pb-2">
          <Link href="/cuenta" className={buttonVariants({ variant: "ghost" })}>
            Mis datos
          </Link>
          <Link href="/cuenta/pedidos" className={buttonVariants({ variant: "ghost" })}>
            Mis pedidos
          </Link>
        </nav>
        <main>{children}</main>
      </div>
      <PieTienda />
    </>
  );
}
