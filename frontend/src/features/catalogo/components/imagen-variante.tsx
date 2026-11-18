"use client";

import Image from "next/image";
import { useState } from "react";

import { cn } from "@/lib/utils";

// Foto de una variante. Si no existe o falla al cargar, una alternativa
// visual que no rompe la página (H05). El contenedor tiene la proporción
// fija, así la foto no hace saltar la página al llegar (CLS).
//
// Las fotos del propio sitio («/…») pasan por el optimizador de Next: AVIF o
// WebP según el navegador y el ancho justo para cada pantalla (srcset y
// sizes). Las de otro dominio se muestran tal cual: todavía no hay un origen
// de fotos decidido (spec, sección 8) y el optimizador no se abre a
// cualquier sitio.
//
// `principal` es para la foto que se ve sin desplazar (el detalle): se pide
// enseguida y con prioridad. Las demás se cargan diferidas.
export function ImagenVariante({
  src,
  alt,
  className,
  sizes = "(min-width: 1024px) 25vw, (min-width: 640px) 50vw, 100vw",
  principal = false,
  rotulo,
}: {
  src?: string;
  alt: string;
  className?: string;
  sizes?: string;
  principal?: boolean;
  // Texto de la etiqueta cuando no hay foto (la marca): la vitrina se lee
  // como el rótulo del producto y no como un error.
  rotulo?: string;
}) {
  const [fallo, setFallo] = useState(false);
  // Proporción 4:5, la de una foto de frasco (DESIGN.md, «vitrina»).
  const contenedor = cn("relative aspect-[4/5] w-full overflow-hidden bg-muted", className);

  if (!src || fallo) {
    return (
      <div
        {...(alt ? { role: "img", "aria-label": `${alt}: sin imagen` } : { "aria-hidden": true })}
        className={cn(contenedor, "grid place-items-center p-[10%] text-center")}
      >
        {rotulo ? (
          <span className="grid gap-2 justify-items-center">
            <span className="text-[clamp(0.8125rem,3.5vw,1.125rem)] leading-tight font-light tracking-[0.2em] text-foreground/80 uppercase">{rotulo}</span>
            <span className="h-px w-8 bg-foreground/30" />
            <span className="text-xs text-muted-foreground">Sin foto por ahora</span>
          </span>
        ) : (
          <span className="text-sm text-muted-foreground">Sin imagen</span>
        )}
      </div>
    );
  }

  return (
    <div className={contenedor}>
      <Image
        src={src}
        alt={alt}
        fill
        sizes={sizes}
        unoptimized={!src.startsWith("/")}
        loading={principal ? "eager" : "lazy"}
        fetchPriority={principal ? "high" : "auto"}
        onError={() => setFallo(true)}
        className="object-contain p-[6%] transition-transform duration-500 ease-out group-hover:scale-[1.03]"
      />
    </div>
  );
}
