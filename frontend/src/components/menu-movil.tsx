"use client";

import { MenuIcon, XIcon } from "lucide-react";
import { useEffect, useId, useState, type ReactNode } from "react";

import { Button } from "@/components/ui/button";

// Menú del teléfono: un botón que abre y cierra un panel debajo del
// encabezado. Se cierra al elegir un link, con Escape o con el mismo botón.
// El contenido (categorías y cuenta) lo arma el servidor.
export function MenuMovil({ children }: { children: ReactNode }) {
  const [abierto, setAbierto] = useState(false);
  const id = useId();

  useEffect(() => {
    if (!abierto) return;
    const alTeclear = (e: KeyboardEvent) => {
      if (e.key === "Escape") setAbierto(false);
    };
    window.addEventListener("keydown", alTeclear);
    return () => window.removeEventListener("keydown", alTeclear);
  }, [abierto]);

  return (
    <>
      <Button
        type="button"
        variant="ghost"
        size="icon"
        aria-expanded={abierto}
        aria-controls={id}
        onClick={() => setAbierto((a) => !a)}
        className="-ml-2"
      >
        {abierto ? <XIcon className="size-6" aria-hidden="true" /> : <MenuIcon className="size-6" aria-hidden="true" />}
        <span className="sr-only">{abierto ? "Cerrar el menú" : "Abrir el menú"}</span>
      </Button>
      <div
        id={id}
        hidden={!abierto}
        onClick={(e) => {
          if ((e.target as HTMLElement).closest("a")) setAbierto(false);
        }}
        className="absolute inset-x-0 top-full z-50 max-h-[calc(100dvh-8rem)] overflow-y-auto border-b bg-background px-4 pb-6 shadow-lg animate-in fade-in slide-in-from-top-1 duration-200"
      >
        {children}
      </div>
    </>
  );
}
