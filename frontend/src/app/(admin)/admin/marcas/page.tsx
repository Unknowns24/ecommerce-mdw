import type { Metadata } from "next";

import { BotonAccion, Formulario } from "@/components/formulario";
import { SinDatos } from "@/components/sin-datos";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { listarMarcas } from "@/features/catalogo/admin/api";
import { EstadoCatalogo } from "@/features/catalogo/admin/estado";
import { activo } from "@/features/catalogo/admin/tipos";
import { requerirPermiso } from "@/features/identidad/sesion";

import { crearMarcaAction, desactivarMarcaAction, editarMarcaAction } from "./acciones";

export const metadata: Metadata = { title: "Marcas" };

// Marcas (H08): alta, edición y desactivación.
export default async function Marcas() {
  await requerirPermiso("catalogo.escribir");
  const marcas = (await listarMarcas()).items;
  return (
    <section className="grid gap-8">
      <h1 className="text-2xl font-semibold">Marcas</h1>
      <Card>
        <CardHeader>
          <CardTitle>Nueva marca</CardTitle>
        </CardHeader>
        <CardContent>
          <Formulario
            accion={crearMarcaAction}
            campos={[{ nombre: "nombre", etiqueta: "Nombre", requerido: true }]}
            textoEnviar="Crear marca"
            textoEnviando="Creando la marca…"
          />
        </CardContent>
      </Card>
      {marcas.length === 0 ? (
        <SinDatos titulo="Todavía no hay marcas" descripcion="Creá la primera con el formulario de arriba." />
      ) : (
        <ul className="grid gap-4 sm:grid-cols-2">
          {marcas.map((m) => (
            <li key={m.id}>
              <Card>
                <CardHeader>
                  <CardTitle className="flex flex-wrap items-center gap-2">
                    {m.nombre} <EstadoCatalogo estado={m.estado} />
                  </CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4">
                  <Formulario
                    accion={editarMarcaAction.bind(null, m.id)}
                    campos={[{ nombre: "nombre", etiqueta: `Nombre de ${m.nombre}`, requerido: true, valorInicial: m.nombre }]}
                    textoEnviar="Guardar"
                    textoEnviando="Guardando…"
                    varianteBoton="outline"
                  />
                  {activo(m.estado) ? (
                    <BotonAccion
                      accion={desactivarMarcaAction.bind(null, m.id)}
                      texto="Desactivar"
                      textoEnviando="Desactivando…"
                      variante="destructive"
                      etiqueta={`Desactivar la marca ${m.nombre}`}
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
