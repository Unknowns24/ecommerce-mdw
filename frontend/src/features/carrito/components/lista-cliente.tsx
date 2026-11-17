"use client";

import { useState } from "react";
import { Alert } from "@/components/ui/alert";
import { listarDisponibles, textoDisponibles } from "../disponibles";
import type { ConflictoCarrito, LineaCarrito } from "../tipos";
import { LineaCliente } from "./linea-cliente";

// Lista del carrito de quien tiene sesión. Guarda el último 409 de cualquier
// línea y lo dibuja ARRIBA de la lista (AGENTS.md: «409 arriba del
// formulario»); el 400 sigue debajo del campo, dentro de cada línea.
export function ListaCliente({ lineas }: { lineas: LineaCarrito[] }) {
  const [conflicto, setConflicto] = useState<ConflictoCarrito | null>(null);
  const filas = conflicto ? listarDisponibles(conflicto.detalles) : [];

  return (
    <div className="flex flex-col gap-4">
      {conflicto ? (
        <Alert variant="destructive" role="alert">
          <p className="font-medium">{conflicto.mensaje}</p>
          {filas.length > 0 ? (
            <ul className="list-disc pl-5">
              {filas.map((f) => (
                <li key={f.varianteId}>
                  {lineas.find((l) => l.varianteId === f.varianteId)?.nombre ?? "Producto"}:{" "}
                  {textoDisponibles(f.disponibles)}
                </li>
              ))}
            </ul>
          ) : null}
        </Alert>
      ) : null}
      <ul className="flex flex-col gap-3">
        {lineas.map((linea) => (
          <LineaCliente key={linea.id} linea={linea} alConflicto={setConflicto} />
        ))}
      </ul>
    </div>
  );
}
