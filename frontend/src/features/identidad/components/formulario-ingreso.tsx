"use client";

import { useActionState } from "react";

import { Campo } from "@/components/campo";
import { Button } from "@/components/ui/button";
import { SIN_ENVIAR, type EstadoFormulario } from "@/lib/acciones";

import { AvisoFormulario } from "./aviso-formulario";

type Accion = (prev: EstadoFormulario, formData: FormData) => Promise<EstadoFormulario>;

export function FormularioIngreso({ accion }: { accion: Accion }) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  return (
    <form action={enviar} className="grid gap-4" noValidate>
      <AvisoFormulario mensaje={estado.mensaje} />
      <Campo etiqueta="Correo" nombre="correo" tipo="email" inputMode="email" requerido valorInicial={estado.valores?.correo} error={estado.errores?.correo} />
      <Campo etiqueta="Contraseña" nombre="contrasena" tipo="password" requerido error={estado.errores?.contrasena} />
      <Button type="submit" disabled={pendiente} className="w-full sm:w-auto">
        {pendiente ? "Ingresando…" : "Ingresar"}
      </Button>
    </form>
  );
}
