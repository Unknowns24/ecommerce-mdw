import type { CuerpoCrearPedido, ItemPedido, MedioPago, ModoEntrega } from "./tipos";

// Funciones puras del formulario de checkout (sin server-only: las prueba y las
// usa cualquiera). No hay reglas de negocio acá: solo se arma el cuerpo que
// espera POST /api/pedidos y se ubican los errores del 400.

/** Nombres de los controles del formulario (coinciden con las claves de la API). */
export const CAMPOS = [
  "nombre",
  "apellido",
  "correo",
  "telefono",
  "dni",
  "modoEntrega",
  "domicilio",
  "distanciaMetros",
  "medioPago",
] as const;

/**
 * Lo que llega del navegador no se confía: de cada ítem solo pasa
 * {varianteId, unidades entera >= 1}. Precios, ids de usuario y cualquier otro
 * campo se descartan (la API tampoco los lee).
 */
export function sanearItems(items: unknown): ItemPedido[] {
  const limpios: ItemPedido[] = [];
  if (!Array.isArray(items)) return limpios;
  for (const it of items) {
    if (
      it &&
      typeof it.varianteId === "string" &&
      Number.isInteger(it.unidades) &&
      it.unidades >= 1
    ) {
      limpios.push({ varianteId: it.varianteId, unidades: it.unidades });
    }
  }
  return limpios;
}

export function aCuerpoPedido(valores: Record<string, string>, items: ItemPedido[]): CuerpoCrearPedido {
  const modoEntrega = (valores.modoEntrega === "ENVIO" ? "ENVIO" : "RETIRO") satisfies ModoEntrega;
  const medioPago = (valores.medioPago === "EFECTIVO" ? "EFECTIVO" : "MERCADO_PAGO") satisfies MedioPago;
  const cuerpo: CuerpoCrearPedido = {
    items,
    comprador: {
      nombre: valores.nombre ?? "",
      apellido: valores.apellido ?? "",
      correo: valores.correo ?? "",
      telefono: valores.telefono ?? "",
      dni: valores.dni ?? "",
    },
    modoEntrega,
    medioPago,
  };
  if (modoEntrega === "ENVIO") {
    cuerpo.domicilio = valores.domicilio ?? "";
    // Si no es un entero no se manda: el 400 de la API lo explica debajo del campo.
    const metros = (valores.distanciaMetros ?? "").trim();
    if (/^\d+$/.test(metros)) cuerpo.distanciaMetros = Number(metros);
  }
  return cuerpo;
}

/**
 * Reparte `detalles.campos` del 400: "comprador.nombre" → campo "nombre",
 * los demás campos del formulario tal cual, y lo que no tiene control propio
 * ("items", "body"…) pasa a `generales` para mostrarlo arriba del formulario.
 */
export function ubicarErrores(campos: Record<string, string> | undefined): {
  porCampo: Record<string, string>;
  generales: string[];
} {
  const porCampo: Record<string, string> = {};
  const generales: string[] = [];
  for (const [clave, texto] of Object.entries(campos ?? {})) {
    const nombre = clave.startsWith("comprador.") ? clave.slice("comprador.".length) : clave;
    if ((CAMPOS as readonly string[]).includes(nombre)) porCampo[nombre] = texto;
    else generales.push(texto);
  }
  return { porCampo, generales };
}
