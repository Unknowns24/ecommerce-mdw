import type { DefCampo } from "@/components/formulario";

import { aCentavos, aPesos } from "./importe";
import type { Categoria, Marca, Producto, VarianteAdmin } from "./tipos";

// Definición de los campos de los formularios del catálogo. Los nombres son
// los mismos que usa la API, así los 400 caen debajo del campo correcto.

export function camposProducto(marcas: Marca[], categorias: Categoria[], p?: Producto): DefCampo[] {
  return [
    { nombre: "nombre", etiqueta: "Nombre", requerido: true, valorInicial: p?.nombre },
    {
      nombre: "marcaId",
      etiqueta: "Marca",
      tipo: "select",
      requerido: true,
      valorInicial: p?.marcaId ?? "",
      opciones: [{ valor: "", texto: "Elegí una marca" }, ...marcas.map((m) => ({ valor: m.id, texto: m.nombre }))],
    },
    {
      nombre: "metodologiaRotacion",
      etiqueta: "Rotación del stock",
      tipo: "select",
      valorInicial: p?.metodologiaRotacion ?? "FIFO",
      ayuda: "Se aplica a todas las variantes del producto.",
      opciones: [
        { valor: "FIFO", texto: "FIFO: primero los lotes más viejos" },
        { valor: "LIFO", texto: "LIFO: primero los lotes más nuevos" },
      ],
    },
    { nombre: "descripcion", etiqueta: "Descripción", tipo: "textarea", valorInicial: p?.descripcion, ancho: "completo" },
    {
      nombre: "categorias",
      etiqueta: "Categorías",
      tipo: "casillas",
      valorInicial: (p?.categorias ?? []).join(","),
      opciones: categorias.map((c) => ({ valor: c.id, texto: c.nombre })),
      ancho: "completo",
    },
  ];
}

export function camposVariante(v?: VarianteAdmin): DefCampo[] {
  return [
    { nombre: "codigo", etiqueta: "Código", requerido: true, valorInicial: v?.codigo },
    { nombre: "nombre", etiqueta: "Nombre", requerido: true, valorInicial: v?.nombre },
    {
      nombre: "precioMinoristaCentavos",
      etiqueta: "Precio minorista (en pesos)",
      inputMode: "numeric",
      requerido: true,
      ayuda: "Por ejemplo 16500,00",
      valorInicial: v ? aPesos(v.precioMinoristaCentavos) : undefined,
    },
    {
      nombre: "precioMayoristaCentavos",
      etiqueta: "Precio mayorista (en pesos)",
      inputMode: "numeric",
      requerido: true,
      valorInicial: v ? aPesos(v.precioMayoristaCentavos) : undefined,
    },
    { nombre: "descripcion", etiqueta: "Descripción", tipo: "textarea", valorInicial: v?.descripcion, ancho: "completo" },
  ];
}

// Traduce los dos precios del formulario. Un formato que no se entiende vuelve
// como error del campo, sin llamar a la API.
export function preciosDe(valores: Record<string, string>) {
  const minorista = aCentavos(valores.precioMinoristaCentavos);
  const mayorista = aCentavos(valores.precioMayoristaCentavos);
  const errores: Record<string, string> = {};
  if (minorista === null) errores.precioMinoristaCentavos = "Escribí el precio en pesos, por ejemplo 16500,00";
  if (mayorista === null) errores.precioMayoristaCentavos = "Escribí el precio en pesos, por ejemplo 13200,00";
  return { minorista, mayorista, errores };
}
