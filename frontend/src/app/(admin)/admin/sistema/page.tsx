import type { Metadata } from "next";
import { notFound } from "next/navigation";
import type { ReactNode } from "react";

import { SinDatos } from "@/components/sin-datos";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { SinStock } from "@/features/catalogo/components/estado-stock";
import { EsqueletoGrilla } from "@/features/catalogo/components/esqueleto-grilla";
import { requerirSesion } from "@/features/identidad/sesion";
import { InsigniaEstadoPago, InsigniaEstadoPedido } from "@/features/pedidos/components/estado-pedido";
import type { EstadoPago, EstadoPedido, MedioPago } from "@/features/pedidos/tipos";
import { formatearPrecio } from "@/lib/formato";

import { DemoCampos, DemoErrorDeCarga } from "./demo-cliente";

export const metadata: Metadata = { title: "Sistema visual" };

const TEXTO_LARGO =
  "Perfume eau de parfum edición limitada coleccionable con notas de vainilla, ámbar, sándalo y pimienta rosa en presentación de cien mililitros con atomizador de vidrio y caja de regalo, importado directamente desde el fabricante";
const PALABRA_LARGA = "Supercalifragilisticoespialidosoperfumeimportadoedicionlimitada";

const ESTADOS_PEDIDO: EstadoPedido[] = ["PENDIENTE_DE_PAGO", "CONFIRMADO", "DESPACHADO", "COMPLETADO", "CANCELADO"];
const ESTADOS_PAGO: [EstadoPago, MedioPago][] = [
  ["PENDIENTE", "MERCADO_PAGO"],
  ["PENDIENTE", "EFECTIVO"],
  ["APROBADO", "MERCADO_PAGO"],
  ["APROBADO", "EFECTIVO"],
  ["RECHAZADO", "MERCADO_PAGO"],
];

function Seccion({ titulo, children }: { titulo: string; children: ReactNode }) {
  return (
    <section className="grid gap-3">
      <h2 className="text-xl font-semibold">{titulo}</h2>
      {children}
    </section>
  );
}

// Galería del sistema visual (spec, sección 7): cada componente con sus
// estados en una sola pantalla, para revisar contraste y foco con Tab.
export default async function Sistema() {
  const cuenta = await requerirSesion();
  if (cuenta.permisos.length === 0) notFound();

  return (
    <div className="grid gap-10">
      <div className="grid gap-1">
        <h1 className="text-2xl font-semibold">Sistema visual</h1>
        <p className="text-muted-foreground">Recorré esta pantalla con Tab: el foco tiene que verse siempre.</p>
      </div>

      <Seccion titulo="Botones">
        <div className="flex flex-wrap gap-2">
          {(["default", "secondary", "outline", "ghost", "destructive", "link"] as const).map((v) => (
            <Button key={v} variant={v}>
              {v}
            </Button>
          ))}
          <Button disabled>Deshabilitado</Button>
          <Button disabled>Guardando…</Button>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {(["xs", "sm", "default", "lg"] as const).map((s) => (
            <Button key={s} size={s} variant="outline">
              Tamaño {s}
            </Button>
          ))}
        </div>
        <Button className="w-full sm:w-auto">Botón principal a lo ancho en el teléfono</Button>
      </Seccion>

      <Seccion titulo="Insignias y estados">
        <div className="flex flex-wrap gap-2">
          {(["default", "secondary", "outline", "destructive"] as const).map((v) => (
            <Badge key={v} variant={v}>
              {v}
            </Badge>
          ))}
          <SinStock />
        </div>
        <p className="text-sm font-medium">Un pedido en cada estado</p>
        <div className="flex flex-wrap gap-2">
          {ESTADOS_PEDIDO.map((e) => (
            <InsigniaEstadoPedido key={e} estado={e} />
          ))}
        </div>
        <p className="text-sm font-medium">El pago en cada estado</p>
        <div className="flex flex-wrap gap-2">
          {ESTADOS_PAGO.map(([e, m]) => (
            <InsigniaEstadoPago key={`${e}-${m}`} estado={e} medio={m} />
          ))}
        </div>
      </Seccion>

      <Seccion titulo="Avisos">
        <Alert>
          <AlertTitle>Aviso informativo</AlertTitle>
          <AlertDescription>El carrito no reserva unidades ni congela precios.</AlertDescription>
        </Alert>
        <Alert variant="destructive">
          <AlertTitle>Rechazo de una regla (409)</AlertTitle>
          <AlertDescription>No hay unidades suficientes: Love Spell crema 236 ml, quedan 2.</AlertDescription>
        </Alert>
      </Seccion>

      <Seccion titulo="Campos">
        <DemoCampos />
        <div className="grid max-w-sm gap-1.5">
          <Label htmlFor="demo-input-suelto">Input y Label sueltos</Label>
          <Input id="demo-input-suelto" placeholder="Placeholder (no reemplaza a la etiqueta)" />
        </div>
      </Seccion>

      <Seccion titulo="Tarjetas y texto larguísimo">
        <div className="grid gap-4 sm:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>{TEXTO_LARGO}</CardTitle>
              <CardDescription>{formatearPrecio(123456789)}</CardDescription>
            </CardHeader>
            <CardContent>
              <p className="break-words">{PALABRA_LARGA}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Tarjeta corta</CardTitle>
              <CardDescription>{formatearPrecio(99)}</CardDescription>
            </CardHeader>
          </Card>
        </div>
        <Separator />
      </Seccion>

      <Seccion titulo="Tabla (con su propio desplazamiento)">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Pedido</TableHead>
                <TableHead>Comprador</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead>Total</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell>#1024</TableCell>
                <TableCell>{TEXTO_LARGO}</TableCell>
                <TableCell>
                  <InsigniaEstadoPedido estado="CONFIRMADO" />
                </TableCell>
                <TableCell>{formatearPrecio(1650000)}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </Seccion>

      <Seccion titulo="Vacío, sin resultados y error">
        <SinDatos titulo="Todavía no hiciste pedidos" descripcion="Cuando compres algo, lo vas a ver acá." accion={{ texto: "Ir al catálogo", href: "/" }} />
        <SinDatos titulo="Ningún producto coincide" descripcion="No encontramos productos con «zzz»." accion={{ texto: "Limpiar filtros", href: "/" }} />
        <DemoErrorDeCarga />
      </Seccion>

      <Seccion titulo="Esqueletos">
        <div className="grid gap-2">
          <Skeleton className="h-6 w-1/3" />
          <Skeleton className="h-4 w-2/3" />
        </div>
        <EsqueletoGrilla />
      </Seccion>
    </div>
  );
}
