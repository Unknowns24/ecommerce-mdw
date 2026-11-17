import type { Metadata } from "next";

import { SinDatos } from "@/components/sin-datos";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { listarRoles } from "@/features/identidad/api";
import { FormularioRol } from "@/features/identidad/components/formulario-rol";
import { requerirPermiso } from "@/features/identidad/sesion";
import { NOMBRES_PERMISOS } from "@/features/identidad/tipos";

import { crearRolAction } from "./acciones";

export const metadata: Metadata = { title: "Roles" };

// Roles (H03): listar y crear con permisos. El catálogo de permisos es el de
// las acciones implementadas.
export default async function Roles() {
  await requerirPermiso("usuarios.gestionar");
  const roles = await listarRoles();
  return (
    <section className="grid gap-8">
      <h1 className="text-2xl font-semibold">Roles</h1>
      {roles.length === 0 ? (
        <SinDatos titulo="Todavía no hay roles" descripcion="Creá el primero con el formulario de abajo." />
      ) : (
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Rol</TableHead>
                <TableHead>Permisos</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {roles.map((r) => (
                <TableRow key={r.id}>
                  <TableCell className="font-medium">{r.nombre}</TableCell>
                  <TableCell className="whitespace-normal">
                    {r.permisos.length > 0 ? r.permisos.map((p) => NOMBRES_PERMISOS[p] ?? p).join(" · ") : "Sin permisos"}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
      <Card>
        <CardHeader>
          <CardTitle>Nuevo rol</CardTitle>
        </CardHeader>
        <CardContent>
          <FormularioRol accion={crearRolAction} permisos={Object.keys(NOMBRES_PERMISOS)} />
        </CardContent>
      </Card>
    </section>
  );
}
