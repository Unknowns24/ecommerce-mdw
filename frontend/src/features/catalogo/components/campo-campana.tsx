import Image from "next/image";
import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

// Tonos de la paleta para los bloques de campaña (DESIGN.md, sección 12):
// tinta y humo, alternados. Se asignan por posición, no por nombre de
// categoría: si cambian las categorías, la portada sigue armada.
export const TONOS = [
  "bg-primary text-primary-foreground", // tinta
  "bg-muted text-foreground", // humo
] as const;

export const tono = (i: number) => TONOS[i % TONOS.length];

// Fondo de un bloque de campaña: la foto si existe (con un velo para que el
// texto se lea) o el campo de color. El contenido va siempre encima.
export function CampoCampana({
  foto,
  color,
  className,
  children,
}: {
  foto?: string;
  color: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <div className={cn("relative isolate overflow-hidden", foto ? "bg-foreground text-white" : color, className)}>
      {foto ? (
        <>
          <Image src={foto} alt="" fill sizes="(min-width: 768px) 50vw, 100vw" className="-z-10 object-cover" />
          <div aria-hidden="true" className="absolute inset-0 -z-10 bg-linear-to-t from-black/60 via-black/10 to-transparent" />
        </>
      ) : null}
      {children}
    </div>
  );
}
