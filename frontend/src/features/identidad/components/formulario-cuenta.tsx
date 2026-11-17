"use client";

import { useActionState } from "react";

import { Campo } from "@/components/campo";
import { Button } from "@/components/ui/button";
import { SIN_ENVIAR, type EstadoFormulario } from "@/lib/acciones";

import { AvisoFormulario } from "./aviso-formulario";

type Accion = (prev: EstadoFormulario, formData: FormData) => Promise<EstadoFormulario>;

export function FormularioCuenta({ accion, inicial }: { accion: Accion; inicial: { nombre: string; apellido: string; telefono: string } }) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  const v = estado.valores ?? inicial;
  const e = estado.errores ?? {};
  return (
    <form action={enviar} className="grid gap-4" noValidate>
      <AvisoFormulario mensaje={estado.mensaje} />
      <Campo etiqueta="Nombre" nombre="nombre" requerido valorInicial={v.nombre} error={e.nombre} />
      <Campo etiqueta="Apellido" nombre="apellido" requerido valorInicial={v.apellido} error={e.apellido} />
      <Campo etiqueta="Teléfono" nombre="telefono" tipo="tel" inputMode="numeric" valorInicial={v.telefono} error={e.telefono} />
      <Button type="submit" disabled={pendiente} className="w-full sm:w-auto">
        {pendiente ? "Guardando…" : "Guardar cambios"}
      </Button>
    </form>
  );
}
