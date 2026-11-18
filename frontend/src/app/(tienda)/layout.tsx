import type { ReactNode } from "react";

import { BarraPromocion } from "@/components/barra-promocion";
import { EncabezadoTienda } from "@/components/encabezado-tienda";
import { PieTienda } from "@/components/pie-tienda";

// Grupo (tienda): sesión opcional (spec, sección 5).
//
// El contenido va en una columna central de 70rem (la del encabezado). Un bloque con
// data-ancho="completo" (el hero y los bloques de campaña de la portada) ocupa
// todo el ancho de la pantalla, sin barra de desplazamiento horizontal.
export default function LayoutTienda({ children }: { children: ReactNode }) {
  return (
    <>
      <BarraPromocion />
      <EncabezadoTienda />
      <main className="grid w-full grid-cols-[minmax(1rem,1fr)_minmax(0,70rem)_minmax(1rem,1fr)] content-start py-6 [&>*]:col-start-2 [&>[data-ancho=completo]]:col-span-full [&>[data-ancho=completo]]:col-start-1">
        {children}
      </main>
      <PieTienda />
    </>
  );
}
