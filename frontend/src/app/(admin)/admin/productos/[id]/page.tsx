import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { BotonAccion, Formulario } from "@/components/formulario";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { buttonVariants } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { imagenesDe, listarCategorias, listarMarcas, obtenerProducto } from "@/features/catalogo/admin/api";
import { EstadoCatalogo } from "@/features/catalogo/admin/estado";
import { camposProducto, camposVariante } from "@/features/catalogo/admin/formularios";
import { activo } from "@/features/catalogo/admin/tipos";
import { ImagenVariante } from "@/features/catalogo/components/imagen-variante";
import { requerirPermiso } from "@/features/identidad/sesion";
import { ErrorApi } from "@/lib/api";
import { formatearPrecio } from "@/lib/formato";

import {
  agregarImagenAction,
  crearVarianteAction,
  desactivarProductoAction,
  desactivarVarianteAction,
  editarProductoAction,
  editarVarianteAction,
  imagenPrincipalAction,
  quitarImagenAction,
} from "./acciones";

export const metadata: Metadata = { title: "Editar producto" };

const MENSAJES: Record<string, string> = { producto: "Guardamos el producto.", variante: "Guardamos la variante." };

// Producto (H08): edición, variantes con precios minorista y mayorista,
// imágenes y el link «Ver stock» de cada variante.
export default async function EditarProducto({ params, searchParams }: PageProps<"/admin/productos/[id]">) {
  await requerirPermiso("catalogo.escribir");
  const [{ id }, { guardado }] = await Promise.all([params, searchParams]);

  const [producto, marcas, categorias] = await Promise.all([
    obtenerProducto(id).catch((e: unknown) => {
      if (e instanceof ErrorApi && (e.status === 400 || e.status === 404)) notFound();
      throw e;
    }),
    listarMarcas(),
    listarCategorias(),
  ]);
  const variantes = producto.variantes ?? [];
  const imagenes = await Promise.all(variantes.map((v) => imagenesDe(v.id)));
  const aviso = typeof guardado === "string" ? MENSAJES[guardado] : undefined;

  return (
    <section className="grid gap-8">
      <div className="grid gap-2">
        <Link href="/admin/productos" className="text-sm underline underline-offset-4">
          Volver a productos
        </Link>
        <h1 className="flex flex-wrap items-center gap-2 text-2xl font-semibold">
          {producto.nombre} <EstadoCatalogo estado={producto.estado} />
        </h1>
      </div>

      {aviso ? (
        <Alert role="status">
          <AlertDescription>{aviso}</AlertDescription>
        </Alert>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>Datos del producto</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-6">
          <Formulario
            accion={editarProductoAction.bind(null, producto.id)}
            campos={camposProducto(
              marcas.items.filter((m) => activo(m.estado) || m.id === producto.marcaId),
              categorias.items.filter((c) => activo(c.estado) || producto.categorias?.includes(c.id)),
              producto,
            )}
            columnas={2}
            textoEnviar="Guardar producto"
            textoEnviando="Guardando…"
          />
          {activo(producto.estado) ? (
            <div className="max-w-xs">
              <BotonAccion
                accion={desactivarProductoAction.bind(null, producto.id)}
                texto="Desactivar producto"
                textoEnviando="Desactivando…"
                variante="destructive"
              />
              <p className="mt-2 text-sm text-muted-foreground">Deja de ofrecerse en la tienda; su historial se conserva.</p>
            </div>
          ) : null}
        </CardContent>
      </Card>

      <section className="grid gap-4">
        <h2 className="text-xl font-semibold">Variantes</h2>
        {variantes.length === 0 ? (
          <p className="text-muted-foreground">Este producto todavía no tiene variantes: cargá la primera abajo.</p>
        ) : null}
        {variantes.map((v, i) => (
          <Card key={v.id}>
            <CardHeader>
              <CardTitle className="flex flex-wrap items-center gap-2">
                {v.nombre} <EstadoCatalogo estado={v.estado} />
              </CardTitle>
              <p className="text-sm text-muted-foreground">
                {v.codigo} · {formatearPrecio(v.precioMinoristaCentavos)} minorista · {formatearPrecio(v.precioMayoristaCentavos)} mayorista
              </p>
              <div className="flex flex-wrap gap-2">
                <Link href={`/admin/stock/${v.id}`} className={buttonVariants({ variant: "outline", size: "sm" })}>
                  Ver stock
                </Link>
                {activo(v.estado) ? (
                  <Link href={`/variantes/${v.id}`} className={buttonVariants({ variant: "ghost", size: "sm" })}>
                    Ver en la tienda
                  </Link>
                ) : null}
              </div>
            </CardHeader>
            <CardContent className="grid gap-6">
              <Formulario
                accion={editarVarianteAction.bind(null, producto.id, v.id)}
                campos={camposVariante(v)}
                columnas={2}
                textoEnviar="Guardar variante"
                textoEnviando="Guardando…"
              />

              <Separator />
              <section className="grid gap-3">
                <h3 className="font-semibold">Imágenes</h3>
                {imagenes[i].length === 0 ? <p className="text-sm text-muted-foreground">Sin imágenes todavía.</p> : null}
                <ul className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                  {imagenes[i].map((img) => (
                    <li key={img.id} className="grid gap-2">
                      <ImagenVariante src={img.referenciaImagen} alt={`${v.nombre}, imagen ${img.orden}`} sizes="(min-width: 640px) 25vw, 50vw" />
                      {img.esPrincipal ? (
                        <Badge>Principal</Badge>
                      ) : (
                        <BotonAccion
                          accion={imagenPrincipalAction.bind(null, producto.id, v.id, img.id)}
                          texto="Elegir como principal"
                          textoEnviando="Guardando…"
                        />
                      )}
                      <BotonAccion
                        accion={quitarImagenAction.bind(null, producto.id, v.id, img.id)}
                        texto="Quitar"
                        textoEnviando="Quitando…"
                        variante="ghost"
                        etiqueta={`Quitar la imagen ${img.orden} de ${v.nombre}`}
                      />
                    </li>
                  ))}
                </ul>
                <Formulario
                  accion={agregarImagenAction.bind(null, producto.id, v.id)}
                  campos={[
                    { nombre: "referenciaImagen", etiqueta: "Dirección de la imagen", tipo: "url", requerido: true, ayuda: "Una URL https de la foto." },
                    { nombre: "orden", etiqueta: "Orden", inputMode: "numeric", valorInicial: String(imagenes[i].length + 1) },
                  ]}
                  columnas={2}
                  textoEnviar="Agregar imagen"
                  textoEnviando="Agregando…"
                  varianteBoton="outline"
                />
              </section>

              {activo(v.estado) ? (
                <div className="max-w-xs">
                  <BotonAccion
                    accion={desactivarVarianteAction.bind(null, producto.id, v.id)}
                    texto="Desactivar variante"
                    textoEnviando="Desactivando…"
                    variante="destructive"
                  />
                </div>
              ) : null}
            </CardContent>
          </Card>
        ))}

        <Card>
          <CardHeader>
            <CardTitle>Nueva variante</CardTitle>
          </CardHeader>
          <CardContent>
            <Formulario
              accion={crearVarianteAction.bind(null, producto.id)}
              campos={camposVariante()}
              columnas={2}
              textoEnviar="Crear variante"
              textoEnviando="Creando la variante…"
            />
          </CardContent>
        </Card>
      </section>
    </section>
  );
}
