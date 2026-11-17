import type { Metadata } from "next";
import Link from "next/link";

import { SinDatos } from "@/components/sin-datos";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { listarPedidos, TAMANO_PAGINA, type FiltrosPedidos } from "@/features/backoffice/api";
import { EstadoDelPago, EstadoDelPedido } from "@/features/backoffice/components/estados";
import { ESTADOS_PEDIDO, TEXTO_ENTREGA } from "@/features/backoffice/tipos";
import { requerirPermiso } from "@/features/identidad/sesion";
import { ErrorApi } from "@/lib/api";
import { formatearFecha, formatearPrecio } from "@/lib/formato";

export const metadata: Metadata = { title: "Pedidos" };

const uno = (v: string | string[] | undefined) => (Array.isArray(v) ? v[0] : v)?.trim() ?? "";

function href(f: FiltrosPedidos, pagina: number) {
  const q = new URLSearchParams();
  if (f.estado) q.set("estado", f.estado);
  if (f.desde) q.set("desde", f.desde);
  if (f.hasta) q.set("hasta", f.hasta);
  if (pagina > 1) q.set("pagina", String(pagina));
  const s = q.toString();
  return s ? `/admin/pedidos?${s}` : "/admin/pedidos";
}

// Pedidos del panel (H17): listado paginado con filtros por estado y fecha.
export default async function PedidosPanel({ searchParams }: PageProps<"/admin/pedidos">) {
  await requerirPermiso("pedidos.gestionar");
  const sp = await searchParams;
  const filtros: FiltrosPedidos = {
    estado: uno(sp.estado),
    desde: uno(sp.desde),
    hasta: uno(sp.hasta),
    pagina: Math.max(1, Number.parseInt(uno(sp.pagina), 10) || 1),
  };
  const hayFiltros = Boolean(filtros.estado || filtros.desde || filtros.hasta);

  // Un filtro inválido (400) se muestra debajo de su campo, no como error del sistema.
  const resultado = await listarPedidos(filtros).then(
    (p) => ({ pagina: p, errores: {} as Record<string, string> }),
    (e: unknown) => {
      if (e instanceof ErrorApi && e.status === 400) {
        const campos = (e.cuerpo.detalles as { campos?: Record<string, string> } | undefined)?.campos ?? {};
        return { pagina: null, errores: campos };
      }
      throw e;
    },
  );
  const { pagina, errores } = resultado;
  const paginas = pagina ? Math.max(1, Math.ceil(pagina.total / TAMANO_PAGINA)) : 1;

  return (
    <section className="grid gap-6">
      <h1 className="text-2xl font-semibold">Pedidos</h1>

      <form action="/admin/pedidos" className="grid gap-3 sm:grid-cols-4 sm:items-end" aria-label="Filtrar pedidos">
        <div className="grid gap-1.5">
          <Label htmlFor="filtro-estado">Estado</Label>
          <Select id="filtro-estado" name="estado" defaultValue={filtros.estado} aria-invalid={errores.estado ? true : undefined}>
            <option value="">Todos</option>
            {Object.entries(ESTADOS_PEDIDO).map(([valor, texto]) => (
              <option key={valor} value={valor}>
                {texto}
              </option>
            ))}
          </Select>
          {errores.estado ? <p className="text-sm text-destructive">{errores.estado}</p> : null}
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="filtro-desde">Desde</Label>
          <Input id="filtro-desde" name="desde" type="date" defaultValue={filtros.desde} className="h-10" aria-invalid={errores.desde ? true : undefined} />
          {errores.desde ? <p className="text-sm text-destructive">{errores.desde}</p> : null}
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="filtro-hasta">Hasta</Label>
          <Input id="filtro-hasta" name="hasta" type="date" defaultValue={filtros.hasta} className="h-10" aria-invalid={errores.hasta ? true : undefined} />
          {errores.hasta ? <p className="text-sm text-destructive">{errores.hasta}</p> : null}
        </div>
        <div className="flex flex-col gap-2 sm:flex-row">
          <Button type="submit" className="w-full sm:w-auto">
            Filtrar
          </Button>
          {hayFiltros ? (
            <Button asChild variant="outline" className="w-full sm:w-auto">
              <Link href="/admin/pedidos">Limpiar</Link>
            </Button>
          ) : null}
        </div>
      </form>

      {pagina === null ? null : pagina.total === 0 && !hayFiltros ? (
        <SinDatos titulo="Todavía no hay pedidos" descripcion="Cuando alguien compre en la tienda, el pedido aparece acá." />
      ) : pagina.items.length === 0 ? (
        <SinDatos
          titulo="Ningún pedido coincide"
          descripcion="No hay pedidos con ese estado o en esas fechas."
          accion={{ texto: "Limpiar filtros", href: "/admin/pedidos" }}
        />
      ) : (
        <>
          <p className="text-sm text-muted-foreground" role="status">
            {pagina.total === 1 ? "1 pedido" : `${pagina.total} pedidos`}
          </p>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Pedido</TableHead>
                  <TableHead>Fecha</TableHead>
                  <TableHead>Comprador</TableHead>
                  <TableHead>Entrega</TableHead>
                  <TableHead>Estado</TableHead>
                  <TableHead>Pago</TableHead>
                  <TableHead className="text-right">Total</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {pagina.items.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell>
                      <Link href={`/admin/pedidos/${p.id}`} className="font-medium underline underline-offset-4">
                        Pedido #{p.numero}
                      </Link>
                    </TableCell>
                    <TableCell>{formatearFecha(p.creadoEn)}</TableCell>
                    <TableCell>
                      {p.compradorNombre} {p.compradorApellido}
                    </TableCell>
                    <TableCell>{TEXTO_ENTREGA[p.modoEntrega]}</TableCell>
                    <TableCell>
                      <EstadoDelPedido estado={p.estadoPedido} />
                    </TableCell>
                    <TableCell>
                      <EstadoDelPago estado={p.estadoPago} medio={p.medioPago} />
                    </TableCell>
                    <TableCell className="text-right">{formatearPrecio(p.totalCentavos)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
          {paginas > 1 ? (
            <nav aria-label="Páginas de pedidos" className="flex items-center justify-between gap-2">
              {filtros.pagina > 1 ? (
                <Button asChild variant="outline">
                  <Link href={href(filtros, filtros.pagina - 1)}>Anterior</Link>
                </Button>
              ) : (
                <span />
              )}
              <p className="text-sm text-muted-foreground">
                Página {filtros.pagina} de {paginas}
              </p>
              {filtros.pagina < paginas ? (
                <Button asChild variant="outline">
                  <Link href={href(filtros, filtros.pagina + 1)}>Siguiente</Link>
                </Button>
              ) : (
                <span />
              )}
            </nav>
          ) : null}
        </>
      )}
    </section>
  );
}
