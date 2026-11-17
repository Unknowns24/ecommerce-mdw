"use client";

import { useActionState } from "react";

import { Campo } from "@/components/campo";
import { Button } from "@/components/ui/button";
import { SIN_ENVIAR, type EstadoFormulario } from "@/lib/acciones";

import { AvisoFormulario } from "./aviso-formulario";

type Accion = (prev: EstadoFormulario, formData: FormData) => Promise<EstadoFormulario>;

export function FormularioNuevoUsuario({ accion }: { accion: Accion }) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  const v = estado.valores ?? {};
  const e = estado.errores ?? {};
  return (
    <form action={enviar} className="grid gap-4 sm:grid-cols-2" noValidate>
      <div className="sm:col-span-2">
        <AvisoFormulario mensaje={estado.mensaje} />
      </div>
      <Campo etiqueta="Nombre" nombre="nombre" requerido valorInicial={v.nombre} error={e.nombre} />
      <Campo etiqueta="Apellido" nombre="apellido" requerido valorInicial={v.apellido} error={e.apellido} />
      <Campo etiqueta="Correo" nombre="correo" tipo="email" inputMode="email" requerido valorInicial={v.correo} error={e.correo} />
      <Campo etiqueta="Contraseña inicial" nombre="contrasena" tipo="password" requerido ayuda="Al menos 8 caracteres. Pasásela a la persona por un canal seguro." error={e.contrasena} />
      <div className="sm:col-span-2">
        <Button type="submit" disabled={pendiente} className="w-full sm:w-auto">
          {pendiente ? "Creando la cuenta…" : "Crear cuenta administrativa"}
        </Button>
      </div>
    </form>
  );
}

export function FormularioEstadoUsuario({ accion, activo, nombre }: { accion: Accion; activo: boolean; nombre: string }) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  return (
    <form action={enviar} className="grid gap-2">
      <AvisoFormulario mensaje={estado.mensaje} />
      <Button type="submit" variant="outline" disabled={pendiente} aria-label={`${activo ? "Desactivar" : "Activar"} a ${nombre}`}>
        {pendiente ? "Guardando…" : activo ? "Desactivar" : "Activar"}
      </Button>
    </form>
  );
}

export function FormularioRolesUsuario({
  accion,
  roles,
  asignados,
  nombre,
}: {
  accion: Accion;
  roles: string[];
  asignados: string[];
  nombre: string;
}) {
  const [estado, enviar, pendiente] = useActionState(accion, SIN_ENVIAR);
  const marcados = estado.valores?.roles !== undefined ? estado.valores.roles.split(",").filter(Boolean) : asignados;
  return (
    <form action={enviar} className="grid gap-2">
      <AvisoFormulario mensaje={estado.mensaje} />
      <fieldset className="grid gap-1">
        <legend className="text-sm font-medium">Roles de {nombre}</legend>
        {roles.map((r) => (
          <label key={r} className="flex items-center gap-2 text-sm">
            <input type="checkbox" name="roles" value={r} defaultChecked={marcados.includes(r)} className="size-4 accent-primary" />
            {r}
          </label>
        ))}
        {estado.errores?.roles ? <p className="text-sm text-destructive">{estado.errores.roles}</p> : null}
      </fieldset>
      <Button type="submit" variant="outline" disabled={pendiente}>
        {pendiente ? "Guardando…" : "Guardar roles"}
      </Button>
    </form>
  );
}
