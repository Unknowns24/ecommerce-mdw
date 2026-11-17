import type { Metadata } from "next";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { BotonSalir } from "@/features/identidad/components/boton-salir";
import { FormularioCuenta } from "@/features/identidad/components/formulario-cuenta";
import type { Usuario } from "@/features/identidad/tipos";
import { api } from "@/lib/api";

import { actualizarCuentaAction } from "./acciones";

export const metadata: Metadata = { title: "Mi cuenta" };

// Mi cuenta (H02): ver y editar los datos permitidos, y cerrar sesión.
export default async function MiCuenta({ searchParams }: PageProps<"/cuenta">) {
  const [{ guardado }, usuario] = await Promise.all([searchParams, api<Usuario>("/api/cuenta")]);
  return (
    <section className="grid max-w-md gap-6">
      <div className="grid gap-1">
        <h1 className="font-heading text-3xl leading-tight font-semibold">Mi cuenta</h1>
        <p className="text-muted-foreground">{usuario.correo}</p>
      </div>
      {guardado ? (
        <Alert role="status">
          <AlertDescription>Guardamos tus datos.</AlertDescription>
        </Alert>
      ) : null}
      <FormularioCuenta
        accion={actualizarCuentaAction}
        inicial={{ nombre: usuario.nombre, apellido: usuario.apellido, telefono: usuario.telefono ?? "" }}
      />
      <div>
        <BotonSalir />
      </div>
    </section>
  );
}
