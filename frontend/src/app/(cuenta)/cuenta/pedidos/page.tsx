import type { Metadata } from "next";
import Link from "next/link";
import { SinDatos } from "@/components/sin-datos";
import { buttonVariants } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { requerirSesion } from "@/features/identidad/sesion";
import { listarMisPedidos } from "@/features/pedidos/api";
import {
  InsigniaEstadoPago,
  InsigniaEstadoPedido,
} from "@/features/pedidos/components/estado-pedido";
import { TEXTO_ENTREGA } from "@/features/pedidos/etiquetas";
import { formatearFecha, formatearPrecio } from "@/lib/formato";

export const metadata: Metadata = { title: "Mis pedidos" };

const TAMANO = 10;

function paginaDe(valor: string | string[] | undefined): number {
  const texto = Array.isArray(valor) ? valor[0] : valor;
  return texto && /^\d+$/.test(texto) && Number(texto) >= 1 ? Number(texto) : 1;
}

// /cuenta/pedidos (H16): los pedidos del cliente, paginados. El layout de
// (cuenta) redirige sin sesión, pero eso no autoriza: la página la exige de
// nuevo y la API decide con la sesión.
//
// Los cinco estados: cargando (loading.tsx), con datos, vacío («Todavía no
// hiciste pedidos» → catálogo), página sin pedidos (hay pedidos, pero no en
// esa página: texto distinto del vacío) y error (error.tsx).
export default async function PaginaMisPedidos({
  searchParams,
}: {
  searchParams: Promise<{ [clave: string]: string | string[] | undefined }>;
}) {
  await requerirSesion();
  const pagina = paginaDe((await searchParams).pagina);
  const datos = await listarMisPedidos(pagina, TAMANO);
  const totalPaginas = Math.max(1, Math.ceil(datos.total / datos.tamano));

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
      <h1 className="text-2xl font-semibold">Mis pedidos</h1>

      {datos.total === 0 ? (
        <SinDatos
          titulo="Todavía no hiciste pedidos"
          descripcion="Cuando compres, tus pedidos van a aparecer acá."
          accion={{ texto: "Ir al catálogo", href: "/" }}
        />
      ) : datos.items.length === 0 ? (
        <SinDatos
          titulo={`No hay pedidos en la página ${pagina}`}
          descripcion={`Tenés ${datos.total} ${datos.total === 1 ? "pedido" : "pedidos"} en ${totalPaginas} ${totalPaginas === 1 ? "página" : "páginas"}.`}
          accion={{ texto: "Volver a la primera página", href: "/cuenta/pedidos" }}
        />
      ) : (
        <>
          <ul className="flex flex-col gap-3">
            {datos.items.map((p) => (
              <li key={p.id}>
                <Card>
                  <CardContent className="flex flex-col gap-3">
                    <div className="flex flex-wrap items-baseline justify-between gap-2">
                      <p className="font-medium">Pedido #{p.numero}</p>
                      <p className="text-sm text-muted-foreground">{formatearFecha(p.creadoEn)}</p>
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="text-sm">Pedido:</span>
                      <InsigniaEstadoPedido estado={p.estadoPedido} />
                      <span className="text-sm">Pago:</span>
                      <InsigniaEstadoPago estado={p.estadoPago} medio={p.medioPago} />
                    </div>
                    <p className="text-sm">{TEXTO_ENTREGA[p.modoEntrega]}</p>
                    <p className="flex items-baseline justify-between gap-2 font-semibold">
                      <span>Total</span>
                      <span>{formatearPrecio(p.totalCentavos)}</span>
                    </p>
                    <Link
                      href={`/cuenta/pedidos/${p.id}`}
                      className={buttonVariants({ variant: "outline", className: "w-full sm:w-auto sm:self-end" })}
                    >
                      Ver detalle<span className="sr-only"> del pedido #{p.numero}</span>
                    </Link>
                  </CardContent>
                </Card>
              </li>
            ))}
          </ul>

          {totalPaginas > 1 ? (
            <nav aria-label="Paginación de pedidos" className="flex flex-wrap items-center justify-between gap-3">
              {datos.pagina > 1 ? (
                <Link
                  href={`/cuenta/pedidos?pagina=${datos.pagina - 1}`}
                  className={buttonVariants({ variant: "outline" })}
                >
                  Anterior
                </Link>
              ) : (
                <span />
              )}
              <p className="text-sm">
                Página {datos.pagina} de {totalPaginas}
              </p>
              {datos.pagina < totalPaginas ? (
                <Link
                  href={`/cuenta/pedidos?pagina=${datos.pagina + 1}`}
                  className={buttonVariants({ variant: "outline" })}
                >
                  Siguiente
                </Link>
              ) : (
                <span />
              )}
            </nav>
          ) : null}
        </>
      )}
    </main>
  );
}
