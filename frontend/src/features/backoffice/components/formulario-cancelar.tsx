"use client";

import { useActionState } from "react";

import { Campo } from "@/components/campo";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { SIN_ENVIAR, type EstadoFormulario } from "@/lib/acciones";

type Accion = (prev: EstadoFormulario, formData: FormData) => Promise<EstadoFormulario>;

// Cancelar un pedido (H18). Explica las consecuencias antes de confirmar.
export function FormularioCancelar({ accion, pagado, cancelado }: { accion: Accion; pagado: boolean; cancelado: boolean }) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  const resultado = estado.detalles as { aviso?: string | null; cancelado?: boolean } | undefined;

  if (resultado?.cancelado) {
    return (
      <Alert role="status">
        <AlertTitle>Pedido cancelado</AlertTitle>
        <AlertDescription>{resultado.aviso ?? "Se liberaron las unidades reservadas."}</AlertDescription>
      </Alert>
    );
  }

  if (cancelado) {
    return <p className="text-sm">Este pedido está cancelado. El motivo y la fecha figuran en «Entrega y pago».</p>;
  }

  return (
    <form action={enviar} className="grid gap-4" noValidate>
      {estado.mensaje ? (
        <Alert variant="destructive" role="alert">
          <AlertDescription>{estado.mensaje}</AlertDescription>
        </Alert>
      ) : null}
      <p className="text-sm">
        {pagado
          ? "Este pedido está pagado. Cancelarlo no devuelve el dinero: el reintegro se hace a mano, fuera del sistema, y acá solo se registra."
          : "Cancelarlo libera las unidades reservadas. No se puede deshacer."}
      </p>
      <Campo etiqueta="Motivo de la cancelación" nombre="motivo" requerido valorInicial={estado.valores?.motivo} error={estado.errores?.motivo}>
        {(control) => <Textarea {...control} rows={2} />}
      </Campo>
      {pagado ? (
        <div className="grid gap-1">
          <label className="flex items-start gap-2 text-sm">
            <input
              type="checkbox"
              name="reintegro"
              value="si"
              defaultChecked={estado.valores?.reintegro === "si"}
              aria-describedby={estado.errores?.reintegro ? "error-reintegro" : undefined}
              aria-invalid={estado.errores?.reintegro ? true : undefined}
              className="mt-0.5 size-4 accent-primary"
            />
            Confirmo que el reintegro ya se hizo fuera del sistema.
          </label>
          {estado.errores?.reintegro ? (
            <p id="error-reintegro" className="text-sm text-destructive">
              {estado.errores.reintegro}
            </p>
          ) : null}
        </div>
      ) : null}
      <Button type="submit" variant="destructive" disabled={pendiente} className="w-full sm:w-auto">
        {pendiente ? "Cancelando…" : "Cancelar pedido"}
      </Button>
    </form>
  );
}
