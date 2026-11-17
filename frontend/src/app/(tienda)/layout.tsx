import type { ReactNode } from "react";

import { EncabezadoTienda } from "@/components/encabezado-tienda";
import { PieTienda } from "@/components/pie-tienda";

// Grupo (tienda): sesión opcional (spec, sección 5).
export default function LayoutTienda({ children }: { children: ReactNode }) {
  return (
    <>
      <EncabezadoTienda />
      <main className="mx-auto w-full max-w-6xl px-4 py-6">{children}</main>
      <PieTienda />
    </>
  );
}
