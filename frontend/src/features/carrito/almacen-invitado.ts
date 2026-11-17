import type { ItemInvitado } from "./tipos";

// Carrito del invitado en localStorage: SOLO {varianteId, unidades}. Nada de
// precios ni nombres; eso se lee siempre de la API en el servidor.
//
// Decisión del equipo (decisión abierta 3): al iniciar sesión el carrito del
// invitado se CONSERVA SIN MEZCLAR. Con sesión se usa el carrito del cliente
// (API) y esta lista no se toca: ni se une ni se borra. Al cerrar sesión
// vuelve a estar. Por eso ningún código de este módulo escribe acá cuando hay
// sesión (el servidor responde modo "cliente").
//
// Este archivo corre solo en el navegador (lo importan Client Components).

const CLAVE = "carrito-invitado";
const EVENTO = "carrito-invitado:cambio";

function leerCrudo(): string {
  try {
    return window.localStorage.getItem(CLAVE) ?? "[]";
  } catch {
    return "[]"; // almacenamiento bloqueado: carrito vacío
  }
}

function interpretar(crudo: string): ItemInvitado[] {
  try {
    const datos: unknown = JSON.parse(crudo);
    if (!Array.isArray(datos)) return [];
    const items: ItemInvitado[] = [];
    for (const d of datos) {
      if (
        d &&
        typeof d.varianteId === "string" &&
        Number.isInteger(d.unidades) &&
        d.unidades >= 1
      ) {
        items.push({ varianteId: d.varianteId, unidades: d.unidades });
      }
    }
    return items;
  } catch {
    return [];
  }
}

function guardar(items: ItemInvitado[]) {
  try {
    window.localStorage.setItem(CLAVE, JSON.stringify(items));
  } catch {
    return; // sin almacenamiento no se puede conservar el carrito
  }
  window.dispatchEvent(new Event(EVENTO));
}

export function leerCarritoInvitado(): ItemInvitado[] {
  return interpretar(leerCrudo());
}

/** Agregar una variante ya presente suma unidades (SPEC-H07, propuesta). */
export function agregarAlCarritoInvitado(varianteId: string, unidades: number) {
  const items = leerCarritoInvitado();
  const existente = items.find((i) => i.varianteId === varianteId);
  if (existente) existente.unidades += unidades;
  else items.push({ varianteId, unidades });
  guardar(items);
}

export function cambiarUnidadesInvitado(varianteId: string, unidades: number) {
  guardar(leerCarritoInvitado().map((i) => (i.varianteId === varianteId ? { ...i, unidades } : i)));
}

export function quitarDelCarritoInvitado(varianteId: string) {
  guardar(leerCarritoInvitado().filter((i) => i.varianteId !== varianteId));
}

export function vaciarCarritoInvitado() {
  guardar([]);
}

/** Para useSyncExternalStore: el valor es el texto crudo (estable entre lecturas). */
export function suscribirseAlCarritoInvitado(avisar: () => void): () => void {
  window.addEventListener(EVENTO, avisar);
  window.addEventListener("storage", avisar); // cambios desde otra pestaña
  return () => {
    window.removeEventListener(EVENTO, avisar);
    window.removeEventListener("storage", avisar);
  };
}

export function instantaneaCarritoInvitado(): string {
  return leerCrudo();
}

export function instantaneaServidor(): string {
  return "[]";
}

export function unidadesDeInstantanea(crudo: string): number {
  return interpretar(crudo).reduce((suma, i) => suma + i.unidades, 0);
}
