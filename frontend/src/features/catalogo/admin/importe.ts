// Los precios se escriben en pesos («16500» o «16.500,50») y viajan en
// centavos enteros (ADR-004). La conversión es de texto a entero, sin pasar
// por coma flotante. Decidir si un precio es válido es del backend; acá solo
// se traduce el formato.

export function aCentavos(texto: string | undefined): number | null {
  const limpio = (texto ?? "").trim().replace(/\s|\$/g, "");
  if (!limpio) return null;
  // «16.500,50» → «16500,50»; «16500.50» se acepta igual.
  const conComa = /,\d{0,2}$/.test(limpio) || /^-?\d{1,3}(\.\d{3})+$/.test(limpio) ? limpio.replace(/\./g, "").replace(",", ".") : limpio.replace(/,/g, "");
  const m = /^(-?)(\d+)(?:[.,](\d{1,2}))?$/.exec(conComa);
  if (!m) return null;
  const centavos = Number.parseInt(m[2], 10) * 100 + Number.parseInt((m[3] ?? "0").padEnd(2, "0"), 10);
  return m[1] ? -centavos : centavos;
}

// 1650050 → «16500,50» (para el valor inicial de un campo).
export function aPesos(centavos: number): string {
  const signo = centavos < 0 ? "-" : "";
  const abs = Math.abs(centavos);
  return `${signo}${Math.trunc(abs / 100)},${String(abs % 100).padStart(2, "0")}`;
}
