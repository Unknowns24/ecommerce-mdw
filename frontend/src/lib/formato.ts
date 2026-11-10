const precio = new Intl.NumberFormat("es-AR", {
  style: "currency",
  currency: "ARS",
});

const fecha = new Intl.DateTimeFormat("es-AR", {
  dateStyle: "short",
  timeStyle: "short",
  hourCycle: "h23",
  timeZone: "America/Argentina/Buenos_Aires",
});

export function formatearPrecio(centavos: number): string {
  return precio.format(centavos / 100).replace("\u00a0", " ");
}

export function formatearFecha(iso: string): string {
  return fecha.format(new Date(iso));
}
