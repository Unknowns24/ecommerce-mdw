import type { Metadata } from "next";
import Link from "next/link";

import { SinDatos } from "@/components/sin-datos";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { listarRoles, listarUsuarios, POR_PAGINA_USUARIOS } from "@/features/identidad/api";
import { FormularioEstadoUsuario, FormularioNuevoUsuario, FormularioRolesUsuario } from "@/features/identidad/components/formularios-usuarios";
import { requerirPermiso } from "@/features/identidad/sesion";

import { asignarRolesAction, cambiarEstadoAction, crearUsuarioAction } from "./acciones";

export const metadata: Metadata = { title: "Cuentas" };

// Cuentas administrativas (H03): listar, crear, activar y desactivar, y
// asignar roles.
export default async function Usuarios({ searchParams }: PageProps<"/admin/usuarios">) {
  await requerirPermiso("usuarios.gestionar");
  const { pagina: p } = await searchParams;
  const pagina = Math.max(1, Number.parseInt(typeof p === "string" ? p : "1", 10) || 1);
  const [usuarios, roles] = await Promise.all([listarUsuarios(pagina), listarRoles()]);
  const nombresRoles = roles.map((r) => r.nombre).filter((r) => r !== "cliente");

  return (
    <section className="grid gap-8">
      <h1 className="text-2xl font-semibold">Cuentas</h1>

      <Card>
        <CardHeader>
          <CardTitle>Nueva cuenta administrativa</CardTitle>
        </CardHeader>
        <CardContent>
          <FormularioNuevoUsuario accion={crearUsuarioAction} />
        </CardContent>
      </Card>

      {usuarios.length === 0 ? (
        <SinDatos
          titulo={pagina > 1 ? "No hay más cuentas" : "Todavía no hay cuentas"}
          descripcion={pagina > 1 ? "Esta página quedó vacía." : "Creá la primera cuenta administrativa con el formulario de arriba."}
          accion={pagina > 1 ? { texto: "Volver a la primera página", href: "/admin/usuarios" } : undefined}
        />
      ) : (
        <ul className="grid gap-4">
          {usuarios.map((u) => {
            const nombre = `${u.nombre} ${u.apellido}`;
            const activo = u.estado === "ACTIVO";
            return (
              <li key={u.id}>
                <Card>
                  <CardHeader>
                    <CardTitle className="flex flex-wrap items-center gap-2">
                      {nombre}
                      <Badge variant={activo ? "secondary" : "destructive"}>{activo ? "Activa" : "Inactiva"}</Badge>
                      {u.es_dueno_inicial ? <Badge variant="outline">Dueño inicial</Badge> : null}
                    </CardTitle>
                    <p className="text-sm text-muted-foreground">
                      {u.correo} · {u.origen === "ADMINISTRATIVO" ? "Cuenta administrativa" : "Cliente registrado"} · Roles:{" "}
                      {u.roles && u.roles.length > 0 ? u.roles.join(", ") : "ninguno"}
                    </p>
                  </CardHeader>
                  <CardContent className="grid gap-4 sm:grid-cols-2">
                    <FormularioEstadoUsuario accion={cambiarEstadoAction.bind(null, u.id, activo ? "INACTIVO" : "ACTIVO")} activo={activo} nombre={nombre} />
                    {u.origen === "ADMINISTRATIVO" ? (
                      <FormularioRolesUsuario accion={asignarRolesAction.bind(null, u.id)} roles={nombresRoles} asignados={u.roles ?? []} nombre={nombre} />
                    ) : null}
                  </CardContent>
                </Card>
              </li>
            );
          })}
        </ul>
      )}

      <nav aria-label="Páginas de cuentas" className="flex justify-between gap-2">
        {pagina > 1 ? (
          <Button asChild variant="outline">
            <Link href={`/admin/usuarios?pagina=${pagina - 1}`}>Anterior</Link>
          </Button>
        ) : (
          <span />
        )}
        {usuarios.length === POR_PAGINA_USUARIOS ? (
          <Button asChild variant="outline">
            <Link href={`/admin/usuarios?pagina=${pagina + 1}`}>Siguiente</Link>
          </Button>
        ) : null}
      </nav>
    </section>
  );
}
