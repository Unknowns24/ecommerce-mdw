"use client";

import Link from "next/link";
import { useActionState, useEffect, useRef } from "react";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Campo } from "@/components/campo";
import { ESTADO_INICIAL } from "../estado-inicial";
import { agregarAction } from "../acciones";
import { agregarAlCarritoInvitado } from "../almacen-invitado";
import type { EstadoCarrito } from "../tipos";

export function BotonAgregarAlCarrito(props: { varianteId: string; disponible: boolean }) {
  if (!props.disponible) {
    return (
      <Button type="button" disabled className="w-full sm:w-auto">
        SIN STOCK
      </Button>
    );
  }
  return <FormularioAgregar varianteId={props.varianteId} />;
}

function FormularioAgregar({ varianteId }: { varianteId: string }) {
  const [estado, accion, pendiente] = useActionState<EstadoCarrito, FormData>(
    agregarAction.bind(null, varianteId),
    ESTADO_INICIAL,
  );

  // Invitado: el servidor validó la variante; la lista de {varianteId, unidades}
  // se escribe acá, en el navegador. Con sesión (modo "cliente") no se toca.
  const procesado = useRef<EstadoCarrito | null>(null);
  useEffect(() => {
    if (procesado.current === estado) return;
    procesado.current = estado;
    if (estado.modo === "invitado" && estado.unidades) {
      agregarAlCarritoInvitado(varianteId, estado.unidades);
    }
  }, [estado, varianteId]);

  const agregado = estado.modo !== undefined && !estado.mensaje && !estado.errores;
  const disponibles = unidadesDisponibles(estado.detalles, varianteId);

  return (
    <form action={accion} className="flex flex-col gap-3">
      {estado.mensaje ? (
        <Alert variant="destructive" role="alert">
          {estado.mensaje}
          {disponibles !== null
            ? ` Hay ${disponibles} ${disponibles === 1 ? "unidad disponible" : "unidades disponibles"}.`
            : null}
        </Alert>
      ) : null}
      {agregado ? (
        <Alert role="status">
          Agregado al carrito.{" "}
          <Link href="/carrito" className="underline">
            Ver carrito
          </Link>
        </Alert>
      ) : null}
      <Campo
        etiqueta="Unidades"
        nombre="unidades"
        tipo="number"
        inputMode="numeric"
        requerido
        valorInicial={estado.valores?.unidades ?? "1"}
        error={estado.errores?.unidades}
      />
      <Button type="submit" disabled={pendiente} className="w-full sm:w-auto">
        {pendiente ? "Agregando…" : "Agregar al carrito"}
      </Button>
    </form>
  );
}

// El 409 trae detalles { "<varianteId>": <disponible> }.
function unidadesDisponibles(detalles: unknown, varianteId: string): number | null {
  if (typeof detalles !== "object" || detalles === null) return null;
  const n = (detalles as Record<string, unknown>)[varianteId];
  return typeof n === "number" ? n : null;
}
