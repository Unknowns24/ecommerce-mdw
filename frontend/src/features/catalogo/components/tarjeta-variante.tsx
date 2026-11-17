import Link from "next/link";

import { formatearPrecio } from "@/lib/formato";

import type { VarianteResumen } from "../tipos";
import { SinStock } from "./estado-stock";
import { ImagenVariante } from "./imagen-variante";

// Tarjeta del catálogo: sin caja ni sombra. La foto en su vitrina nácar, y
// debajo la marca, el nombre como en la etiqueta del frasco y el precio
// (DESIGN.md, sección 6). Todo el bloque es un solo link.
export function TarjetaVariante({ variante }: { variante: VarianteResumen }) {
  return (
    <Link href={`/variantes/${variante.id}`} className="group grid gap-2 rounded-sm">
      <ImagenVariante src={variante.imagenPrincipal} alt="" sizes="(min-width: 1024px) 22vw, (min-width: 640px) 31vw, 46vw" />
      <div className="grid gap-0.5">
        <p className="text-[0.8125rem] text-muted-foreground">{variante.marca}</p>
        <h2 className="font-heading text-lg leading-tight font-medium text-pretty group-hover:underline group-hover:underline-offset-4">
          {variante.nombre}
        </h2>
        <p className="font-semibold tabular-nums text-precio">{formatearPrecio(variante.precioMinoristaCentavos)}</p>
        {variante.disponible ? null : (
          <div className="pt-1">
            <SinStock />
          </div>
        )}
      </div>
    </Link>
  );
}
