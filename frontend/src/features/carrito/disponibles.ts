// El 409 de la API trae detalles { "<varianteId>": <unidades disponibles> }.
export function unidadesDisponibles(detalles: unknown, varianteId: string): number | null {
  if (typeof detalles !== "object" || detalles === null) return null;
  const n = (detalles as Record<string, unknown>)[varianteId];
  return typeof n === "number" ? n : null;
}

export function textoDisponibles(n: number): string {
  return `Hay ${n} ${n === 1 ? "unidad disponible" : "unidades disponibles"}.`;
}
