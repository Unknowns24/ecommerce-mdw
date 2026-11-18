import Link from "next/link";

import { formatearPrecio } from "@/lib/formato";

import type { VarianteResumen } from "../tipos";
import { SinStock } from "./estado-stock";
import { ImagenVariante } from "./imagen-variante";

// Tarjeta del catálogo: sin caja ni sombra. La foto ocupa todo el ancho de
// la columna, con «SIN STOCK» encima si corresponde; debajo, centrados, la
// marca, el nombre y el precio (DESIGN.md, sección 6). Todo el bloque es un
// solo link.
// `titulo` es el nivel del encabezado: h2 en la grilla del catálogo, h3
// dentro de una sección de la portada.
export function TarjetaVariante({ variante, titulo: Titulo = "h2" }: { variante: VarianteResumen; titulo?: "h2" | "h3" }) {
  return (
    <Link href={`/variantes/${variante.id}`} className="group grid gap-3">
      <div className="relative">
        <ImagenVariante
          src={variante.imagenPrincipal}
          alt=""
          rotulo={variante.marca}
          sizes="(min-width: 640px) 31vw, 48vw"
        />
        {variante.disponible ? null : <SinStock className="absolute top-2 left-2" />}
      </div>
      <div className="grid justify-items-center gap-1 px-1 text-center">
        <p className="text-[0.6875rem] tracking-[0.15em] text-muted-foreground uppercase">{variante.marca}</p>
        <Titulo className="text-sm leading-snug font-normal text-pretty decoration-1 underline-offset-4 group-hover:underline">
          {variante.nombre}
        </Titulo>
        <p className="text-sm font-medium tabular-nums text-precio">{formatearPrecio(variante.precioMinoristaCentavos)}</p>
      </div>
    </Link>
  );
}
