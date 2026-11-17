import type { Metadata } from "next";

import { Formulario } from "@/components/formulario";
import { SinDatos } from "@/components/sin-datos";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { requerirPermiso } from "@/features/identidad/sesion";
import { listarProveedores } from "@/features/stock/api";
import { formatearFecha } from "@/lib/formato";

import { crearProveedorAction } from "./acciones";

export const metadata: Metadata = { title: "Proveedores" };

// Proveedores (H08, H09): listado y alta.
export default async function Proveedores() {
  await requerirPermiso("stock.gestionar");
  const proveedores = await listarProveedores();
  return (
    <section className="grid gap-8">
      <h1 className="text-2xl font-semibold">Proveedores</h1>
      {proveedores.length === 0 ? (
        <SinDatos titulo="Todavía no hay proveedores" descripcion="Cargá el primero con el formulario de abajo para poder ingresar lotes." />
      ) : (
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead>Correo</TableHead>
                <TableHead>Teléfono</TableHead>
                <TableHead>Web</TableHead>
                <TableHead>Alta</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {proveedores.map((p) => (
                <TableRow key={p.id}>
                  <TableCell className="font-medium">{p.nombre}</TableCell>
                  <TableCell>{p.correo ?? "—"}</TableCell>
                  <TableCell>{p.telefono ?? "—"}</TableCell>
                  <TableCell>{p.web ?? "—"}</TableCell>
                  <TableCell>{formatearFecha(p.creadoEn)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
      <Card>
        <CardHeader>
          <CardTitle>Nuevo proveedor</CardTitle>
        </CardHeader>
        <CardContent>
          <Formulario
            accion={crearProveedorAction}
            columnas={2}
            campos={[
              { nombre: "nombre", etiqueta: "Nombre", requerido: true },
              { nombre: "correo", etiqueta: "Correo", tipo: "email", inputMode: "email" },
              { nombre: "telefono", etiqueta: "Teléfono", tipo: "tel", inputMode: "numeric" },
              { nombre: "web", etiqueta: "Sitio web", tipo: "url" },
            ]}
            textoEnviar="Crear proveedor"
            textoEnviando="Creando el proveedor…"
          />
        </CardContent>
      </Card>
    </section>
  );
}
