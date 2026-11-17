import type { Metadata } from "next";

import { BotonAccion, Formulario, type DefCampo } from "@/components/formulario";
import { SinDatos } from "@/components/sin-datos";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { listarCategorias } from "@/features/catalogo/admin/api";
import { EstadoCatalogo } from "@/features/catalogo/admin/estado";
import { activo, type Categoria } from "@/features/catalogo/admin/tipos";
import { requerirPermiso } from "@/features/identidad/sesion";

import { crearCategoriaAction, desactivarCategoriaAction, editarCategoriaAction } from "./acciones";

export const metadata: Metadata = { title: "Categorías" };

function campos(categorias: Categoria[], c?: Categoria): DefCampo[] {
  return [
    { nombre: "nombre", etiqueta: c ? `Nombre de ${c.nombre}` : "Nombre", requerido: true, valorInicial: c?.nombre },
    {
      nombre: "padreId",
      etiqueta: c ? `Categoría padre de ${c.nombre}` : "Categoría padre",
      tipo: "select",
      valorInicial: c?.padreId ?? "",
      ayuda: "Elegí una para crear una subcategoría.",
      opciones: [
        { valor: "", texto: "Ninguna (categoría principal)" },
        ...categorias.filter((o) => o.id !== c?.id && activo(o.estado)).map((o) => ({ valor: o.id, texto: o.nombre })),
      ],
    },
  ];
}

// Categorías y subcategorías (H08).
export default async function Categorias() {
  await requerirPermiso("catalogo.escribir");
  const categorias = (await listarCategorias()).items;
  const nombre = new Map(categorias.map((c) => [c.id, c.nombre]));
  return (
    <section className="grid gap-8">
      <h1 className="text-2xl font-semibold">Categorías</h1>
      <Card>
        <CardHeader>
          <CardTitle>Nueva categoría</CardTitle>
        </CardHeader>
        <CardContent>
          <Formulario accion={crearCategoriaAction} campos={campos(categorias)} columnas={2} textoEnviar="Crear categoría" textoEnviando="Creando la categoría…" />
        </CardContent>
      </Card>
      {categorias.length === 0 ? (
        <SinDatos titulo="Todavía no hay categorías" descripcion="Creá la primera con el formulario de arriba." />
      ) : (
        <ul className="grid gap-4 sm:grid-cols-2">
          {categorias.map((c) => (
            <li key={c.id}>
              <Card>
                <CardHeader>
                  <CardTitle className="flex flex-wrap items-center gap-2">
                    {c.nombre} <EstadoCatalogo estado={c.estado} />
                  </CardTitle>
                  <p className="text-sm text-muted-foreground">{c.padreId ? `Subcategoría de ${nombre.get(c.padreId) ?? "otra categoría"}` : "Categoría principal"}</p>
                </CardHeader>
                <CardContent className="grid gap-4">
                  <Formulario
                    accion={editarCategoriaAction.bind(null, c.id)}
                    campos={campos(categorias, c)}
                    textoEnviar="Guardar"
                    textoEnviando="Guardando…"
                    varianteBoton="outline"
                  />
                  {activo(c.estado) ? (
                    <BotonAccion
                      accion={desactivarCategoriaAction.bind(null, c.id)}
                      texto="Desactivar"
                      textoEnviando="Desactivando…"
                      variante="destructive"
                      etiqueta={`Desactivar la categoría ${c.nombre}`}
                    />
                  ) : null}
                </CardContent>
              </Card>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
