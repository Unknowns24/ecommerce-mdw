// El 409 de la API trae detalles { "<varianteId>": <unidades disponibles> }.
// Único lugar que sabe leer esa forma: lo usan el botón de agregar, el
// carrito del cliente y el checkout.
export function unidadesDisponibles(detalles: unknown, varianteId: string): number | null {
  if (typeof detalles !== "object" || detalles === null) return null;
  const n = (detalles as Record<string, unknown>)[varianteId];
  return typeof n === "number" ? n : null;
}

/** Todas las parejas variante → unidades disponibles de un 409, para enumerarlas. */
export function listarDisponibles(detalles: unknown): Array<{ varianteId: string; disponibles: number }> {
  if (typeof detalles !== "object" || detalles === null) return [];
  const filas: Array<{ varianteId: string; disponibles: number }> = [];
  for (const [varianteId, n] of Object.entries(detalles as Record<string, unknown>)) {
    if (typeof n === "number") filas.push({ varianteId, disponibles: n });
  }
  return filas;
}

export function textoDisponibles(n: number): string {
  return `Hay ${n} ${n === 1 ? "unidad disponible" : "unidades disponibles"}.`;
}
