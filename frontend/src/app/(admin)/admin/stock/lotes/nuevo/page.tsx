import type { Metadata } from "next";

import { Formulario } from "@/components/formulario";
import { SinDatos } from "@/components/sin-datos";
import { requerirPermiso } from "@/features/identidad/sesion";
import { listarProveedores, variantesParaElegir } from "@/features/stock/api";

import { ingresarLoteAction } from "./acciones";

export const metadata: Metadata = { title: "Ingreso de lote" };

const hoy = () => new Intl.DateTimeFormat("en-CA", { timeZone: "America/Argentina/Buenos_Aires" }).format(new Date());

// Ingreso de lote (H09). La variante puede venir elegida desde su stock.
export default async function NuevoLote({ searchParams }: PageProps<"/admin/stock/lotes/nuevo">) {
  await requerirPermiso("stock.gestionar");
  const [{ varianteId }, variantes, proveedores] = await Promise.all([searchParams, variantesParaElegir(), listarProveedores()]);

  if (proveedores.length === 0) {
    return (
      <SinDatos
        titulo="Primero cargá un proveedor"
        descripcion="Cada lote se asocia a un proveedor."
        accion={{ texto: "Ir a proveedores", href: "/admin/proveedores" }}
      />
    );
  }

  return (
    <section className="grid max-w-2xl gap-6">
      <h1 className="text-2xl font-semibold">Ingreso de lote</h1>
      <Formulario
        accion={ingresarLoteAction}
        columnas={2}
        campos={[
          {
            nombre: "varianteId",
            etiqueta: "Variante",
            tipo: "select",
            requerido: true,
            valorInicial: typeof varianteId === "string" ? varianteId : "",
            opciones: [{ valor: "", texto: "Elegí una variante" }, ...variantes.map((v) => ({ valor: v.id, texto: `${v.nombre} (${v.codigo})` }))],
            ancho: "completo",
          },
          {
            nombre: "proveedorId",
            etiqueta: "Proveedor",
            tipo: "select",
            requerido: true,
            opciones: [{ valor: "", texto: "Elegí un proveedor" }, ...proveedores.map((p) => ({ valor: p.id, texto: p.nombre }))],
          },
          { nombre: "fecha", etiqueta: "Fecha de ingreso", tipo: "date", requerido: true, valorInicial: hoy() },
          { nombre: "unidades", etiqueta: "Unidades", inputMode: "numeric", requerido: true },
          { nombre: "costoUnitarioCentavos", etiqueta: "Costo unitario (en pesos)", inputMode: "numeric", requerido: true, ayuda: "Por ejemplo 1500,00" },
        ]}
        textoEnviar="Ingresar lote"
        textoEnviando="Ingresando el lote…"
      />
    </section>
  );
}
