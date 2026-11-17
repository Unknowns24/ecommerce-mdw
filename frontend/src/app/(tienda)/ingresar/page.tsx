import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";

import { FormularioIngreso } from "@/features/identidad/components/formulario-ingreso";
import { obtenerCuenta } from "@/features/identidad/sesion";

import { ingresarAction } from "./acciones";

export const metadata: Metadata = { title: "Ingresar" };

export default async function Ingresar({ searchParams }: PageProps<"/ingresar">) {
  if (await obtenerCuenta()) redirect("/cuenta");
  const { siguiente } = await searchParams;
  const destino = typeof siguiente === "string" ? siguiente : undefined;
  const sufijo = destino ? `?siguiente=${encodeURIComponent(destino)}` : "";
  return (
    <section className="mx-auto grid max-w-md gap-6">
      <div className="grid gap-1">
        <h1 className="font-heading text-3xl leading-tight font-semibold">Ingresar</h1>
        <p className="text-muted-foreground">Para comprar no hace falta una cuenta: podés seguir como invitado.</p>
      </div>
      <FormularioIngreso accion={ingresarAction.bind(null, destino)} />
      <p className="text-sm">
        ¿No tenés cuenta?{" "}
        <Link href={`/registro${sufijo}`} className="underline underline-offset-4">
          Registrate
        </Link>
      </p>
    </section>
  );
}
