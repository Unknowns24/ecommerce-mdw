import type { Metadata } from "next";
import Link from "next/link";

import { Formulario } from "@/components/formulario";
import { SinDatos } from "@/components/sin-datos";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { listarCategorias, listarMarcas, listarProductos, POR_PAGINA_ADMIN } from "@/features/catalogo/admin/api";
import { EstadoCatalogo } from "@/features/catalogo/admin/estado";
import { camposProducto } from "@/features/catalogo/admin/formularios";
import { activo } from "@/features/catalogo/admin/tipos";
import { requerirPermiso } from "@/features/identidad/sesion";

import { crearProductoAction } from "./acciones";

export const metadata: Metadata = { title: "Productos" };

// Productos del panel (H08): listado paginado y alta.
export default async function Productos({ searchParams }: PageProps<"/admin/productos">) {
  await requerirPermiso("catalogo.escribir");
  const { pagina: p } = await searchParams;
  const pagina = Math.max(1, Number.parseInt(typeof p === "string" ? p : "1", 10) || 1);
  const [productos, marcas, categorias] = await Promise.all([listarProductos(pagina), listarMarcas(), listarCategorias()]);
  const nombreMarca = new Map(marcas.items.map((m) => [m.id, m.nombre]));
  const paginas = Math.max(1, Math.ceil(productos.total / POR_PAGINA_ADMIN));

  return (
    <section className="grid gap-8">
      <h1 className="text-2xl font-semibold">Productos</h1>

      {productos.total === 0 ? (
        <SinDatos titulo="Todavía no hay productos" descripcion="Cargá el primero con el formulario de abajo." />
      ) : productos.items.length === 0 ? (
        <SinDatos titulo="No hay más productos" descripcion="Esta página quedó vacía." accion={{ texto: "Volver a la primera página", href: "/admin/productos" }} />
      ) : (
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Producto</TableHead>
                <TableHead>Marca</TableHead>
                <TableHead>Rotación</TableHead>
                <TableHead>Estado</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {productos.items.map((prod) => (
                <TableRow key={prod.id}>
                  <TableCell>
                    <Link href={`/admin/productos/${prod.id}`} className="font-medium underline underline-offset-4">
                      {prod.nombre}
                    </Link>
                  </TableCell>
                  <TableCell>{nombreMarca.get(prod.marcaId) ?? "—"}</TableCell>
                  <TableCell>{prod.metodologiaRotacion}</TableCell>
                  <TableCell>
                    <EstadoCatalogo estado={prod.estado} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {paginas > 1 ? (
        <nav aria-label="Páginas de productos" className="flex items-center justify-between gap-2">
          {pagina > 1 ? (
            <Button asChild variant="outline">
              <Link href={`/admin/productos?pagina=${pagina - 1}`}>Anterior</Link>
            </Button>
          ) : (
            <span />
          )}
          <p className="text-sm text-muted-foreground">
            Página {pagina} de {paginas}
          </p>
          {pagina < paginas ? (
            <Button asChild variant="outline">
              <Link href={`/admin/productos?pagina=${pagina + 1}`}>Siguiente</Link>
            </Button>
          ) : (
            <span />
          )}
        </nav>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>Nuevo producto</CardTitle>
        </CardHeader>
        <CardContent>
          <Formulario
            accion={crearProductoAction}
            campos={camposProducto(
              marcas.items.filter((m) => activo(m.estado)),
              categorias.items.filter((c) => activo(c.estado)),
            )}
            columnas={2}
            textoEnviar="Crear producto"
            textoEnviando="Creando el producto…"
          />
        </CardContent>
      </Card>
    </section>
  );
}
