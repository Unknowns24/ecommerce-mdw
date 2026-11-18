"use client";

import { useState } from "react";

import { cn } from "@/lib/utils";

import { ImagenVariante } from "./imagen-variante";

// Galería de la ficha: la foto principal y, si hay más, miniaturas que la
// cambian (botones con aria-pressed, operables con teclado). Sin zoom ni
// carrusel: las fotos del catálogo son de producto, no de campaña.
export function GaleriaVariante({ imagenes, nombre, rotulo }: { imagenes: string[]; nombre: string; rotulo?: string }) {
  const [actual, setActual] = useState(0);
  const principal = imagenes[actual];
  return (
    <div className="grid gap-3">
      <ImagenVariante src={principal} alt={nombre} rotulo={rotulo} principal sizes="(min-width: 768px) 55vw, 100vw" />
      {imagenes.length > 1 ? (
        <ul className="grid grid-cols-5 gap-2" aria-label="Fotos del producto">
          {imagenes.map((src, i) => (
            <li key={`${src}-${i}`}>
              <button
                type="button"
                onClick={() => setActual(i)}
                aria-pressed={i === actual}
                aria-label={`Ver la foto ${i + 1} de ${imagenes.length}`}
                className={cn("block w-full outline-offset-2", i === actual ? "ring-2 ring-primary" : "opacity-80 hover:opacity-100")}
              >
                <ImagenVariante src={src} alt="" sizes="10vw" />
              </button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}
