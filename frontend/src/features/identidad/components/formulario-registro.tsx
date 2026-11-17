"use client";

import { useActionState } from "react";

import { Campo } from "@/components/campo";
import { Button } from "@/components/ui/button";
import { SIN_ENVIAR, type EstadoFormulario } from "@/lib/acciones";

import { AvisoFormulario } from "./aviso-formulario";

type Accion = (prev: EstadoFormulario, formData: FormData) => Promise<EstadoFormulario>;

export function FormularioRegistro({ accion }: { accion: Accion }) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  const v = estado.valores ?? {};
  const e = estado.errores ?? {};
  return (
    <form action={enviar} className="grid gap-4" noValidate>
      <AvisoFormulario mensaje={estado.mensaje} />
      <Campo etiqueta="Nombre" nombre="nombre" requerido valorInicial={v.nombre} error={e.nombre} />
      <Campo etiqueta="Apellido" nombre="apellido" requerido valorInicial={v.apellido} error={e.apellido} />
      <Campo etiqueta="Correo" nombre="correo" tipo="email" inputMode="email" requerido valorInicial={v.correo} error={e.correo} />
      <Campo etiqueta="Contraseña" nombre="contrasena" tipo="password" requerido ayuda="Al menos 8 caracteres." error={e.contrasena} />
      <Button type="submit" disabled={pendiente} className="w-full sm:w-auto">
        {pendiente ? "Creando tu cuenta…" : "Crear cuenta"}
      </Button>
    </form>
  );
}
