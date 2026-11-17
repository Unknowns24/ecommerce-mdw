"use client";

import { useActionState } from "react";

import { Campo } from "@/components/campo";
import { Button } from "@/components/ui/button";
import { SIN_ENVIAR, type EstadoFormulario } from "@/lib/acciones";

import { NOMBRES_PERMISOS } from "../tipos";
import { AvisoFormulario } from "./aviso-formulario";

type Accion = (prev: EstadoFormulario, formData: FormData) => Promise<EstadoFormulario>;

export function FormularioRol({ accion, permisos }: { accion: Accion; permisos: string[] }) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  const marcados = (estado.valores?.permisos ?? "").split(",").filter(Boolean);
  return (
    <form action={enviar} className="grid gap-4" noValidate>
      <AvisoFormulario mensaje={estado.mensaje} />
      <Campo etiqueta="Nombre del rol" nombre="nombre" requerido valorInicial={estado.valores?.nombre} error={estado.errores?.nombre} />
      <fieldset className="grid gap-2">
        <legend className="mb-1 text-sm font-medium">Permisos</legend>
        {permisos.map((p) => (
          <label key={p} className="flex items-start gap-2 text-sm">
            <input type="checkbox" name="permisos" value={p} defaultChecked={marcados.includes(p)} className="mt-0.5 size-4 accent-primary" />
            <span>
              {NOMBRES_PERMISOS[p] ?? p} <span className="text-muted-foreground">({p})</span>
            </span>
          </label>
        ))}
        {estado.errores?.permisos ? <p className="text-sm text-destructive">{estado.errores.permisos}</p> : null}
      </fieldset>
      <Button type="submit" disabled={pendiente} className="w-full sm:w-auto">
        {pendiente ? "Creando el rol…" : "Crear rol"}
      </Button>
    </form>
  );
}
