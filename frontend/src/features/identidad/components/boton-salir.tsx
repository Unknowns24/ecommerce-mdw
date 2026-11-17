"use client";

import { useFormStatus } from "react-dom";

import { Button } from "@/components/ui/button";

import { salirAction } from "../acciones";

function Enviar() {
  const { pending } = useFormStatus();
  return (
    <Button type="submit" variant="ghost" disabled={pending}>
      {pending ? "Saliendo…" : "Salir"}
    </Button>
  );
}

// «Salir», siempre en el mismo lugar (sección 7, «Navegación»).
export function BotonSalir() {
  return (
    <form action={salirAction}>
      <Enviar />
    </form>
  );
}
