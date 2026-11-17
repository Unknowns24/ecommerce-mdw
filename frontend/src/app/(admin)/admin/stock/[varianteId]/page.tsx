import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { Formulario } from "@/components/formulario";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { buttonVariants } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { requerirPermiso } from "@/features/identidad/sesion";
import { listarProveedores, movimientosDe, nombreDeVariante, stockDe } from "@/features/stock/api";
import { TIPOS_MOVIMIENTO } from "@/features/stock/tipos";
import { ErrorApi } from "@/lib/api";
import { formatearFecha, formatearPrecio } from "@/lib/formato";

import { ajustarStockAction } from "./acciones";

export const metadata: Metadata = { title: "Stock de la variante" };

// Stock de una variante (H09): disponible, lotes, movimientos y ajuste.
export default async function StockVariante({ params, searchParams }: PageProps<"/admin/stock/[varianteId]">) {
  await requerirPermiso("stock.gestionar");
  const [{ varianteId }, { ajustado }] = await Promise.all([params, searchParams]);

  const [stock, movimientos, proveedores, nombre] = await Promise.all([
    stockDe(varianteId).catch((e: unknown) => {
      if (e instanceof ErrorApi && (e.status === 400 || e.status === 404)) notFound();
      throw e;
    }),
    movimientosDe(varianteId).catch(() => null),
    listarProveedores(),
    nombreDeVariante(varianteId),
  ]);
  const proveedor = new Map(proveedores.map((p) => [p.id, p.nombre]));

  return (
    <section className="grid gap-6">
      <div className="grid gap-1">
        <h1 className="text-2xl font-semibold">Stock: {nombre ?? "variante inactiva"}</h1>
        <p className="text-3xl font-semibold" role="status">
          {stock.disponible === 1 ? "1 unidad disponible" : `${stock.disponible} unidades disponibles`}
        </p>
        <p className="text-sm text-muted-foreground">Disponible = unidades de los lotes menos las reservadas para pedidos.</p>
        <div>
          <Link href={`/admin/stock/lotes/nuevo?varianteId=${varianteId}`} className={buttonVariants({ variant: "outline" })}>
            Ingresar un lote
          </Link>
        </div>
      </div>

      {ajustado ? (
        <Alert role="status">
          <AlertDescription>Registramos el ajuste.</AlertDescription>
        </Alert>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>Lotes</CardTitle>
        </CardHeader>
        <CardContent>
          {stock.lotes.length === 0 ? (
            <p className="text-muted-foreground">Esta variante todavía no tiene lotes: ingresá el primero.</p>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Ingreso</TableHead>
                    <TableHead>Proveedor</TableHead>
                    <TableHead className="text-right">Unidades ingresadas</TableHead>
                    <TableHead className="text-right">Costo unitario</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {stock.lotes.map((l) => (
                    <TableRow key={l.id}>
                      <TableCell>{formatearFecha(l.fechaIngreso)}</TableCell>
                      <TableCell>{proveedor.get(l.proveedorId) ?? "—"}</TableCell>
                      <TableCell className="text-right">{l.unidadesIngresadas}</TableCell>
                      <TableCell className="text-right">{formatearPrecio(l.costoUnitarioCentavos)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {stock.lotes.length > 0 ? (
        <Card>
          <CardHeader>
            <CardTitle>Ajustar el stock</CardTitle>
          </CardHeader>
          <CardContent>
            <Formulario
              accion={ajustarStockAction.bind(null, varianteId)}
              columnas={2}
              campos={[
                {
                  nombre: "loteId",
                  etiqueta: "Lote",
                  tipo: "select",
                  requerido: true,
                  opciones: stock.lotes.map((l) => ({
                    valor: l.id,
                    texto: `${formatearFecha(l.fechaIngreso)} · ${proveedor.get(l.proveedorId) ?? "proveedor"} · ${l.unidadesIngresadas} u.`,
                  })),
                },
                {
                  nombre: "sentido",
                  etiqueta: "Tipo de ajuste",
                  tipo: "select",
                  opciones: [
                    { valor: "restar", texto: "Restar unidades (rotura, pérdida…)" },
                    { valor: "sumar", texto: "Sumar unidades (recuento)" },
                  ],
                },
                { nombre: "unidades", etiqueta: "Unidades", inputMode: "numeric", requerido: true },
                { nombre: "motivo", etiqueta: "Motivo", requerido: true, ayuda: "Queda registrado en los movimientos." },
              ]}
              textoEnviar="Registrar ajuste"
              textoEnviando="Registrando el ajuste…"
            />
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>Movimientos</CardTitle>
        </CardHeader>
        <CardContent>
          {movimientos === null ? (
            <p className="text-muted-foreground">No pudimos cargar los movimientos. Recargá la página para reintentar.</p>
          ) : movimientos.length === 0 ? (
            <p className="text-muted-foreground">Todavía no hay movimientos.</p>
          ) : (
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Fecha</TableHead>
                    <TableHead>Tipo</TableHead>
                    <TableHead className="text-right">Unidades</TableHead>
                    <TableHead>Motivo</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {movimientos.map((m) => (
                    <TableRow key={m.id}>
                      <TableCell>{formatearFecha(m.creadoEn)}</TableCell>
                      <TableCell>{TIPOS_MOVIMIENTO[m.tipo] ?? m.tipo}</TableCell>
                      <TableCell className="text-right">{m.unidades > 0 ? `+${m.unidades}` : m.unidades}</TableCell>
                      <TableCell className="whitespace-normal">{m.motivo}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </section>
  );
}
