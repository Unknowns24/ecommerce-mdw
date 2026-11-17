// Recibe las Web Vitals que manda MedicionWebVitals y las escribe en el log
// del servidor de Next, una línea JSON por métrica. No es una pantalla ni
// pasa por la API de Go: Nginx manda /metricas al contenedor web. Solo acepta
// los campos conocidos y descarta todo lo demás.

const NOMBRES = new Set(["LCP", "INP", "CLS", "FCP", "TTFB"]);

export async function POST(request: Request) {
  const texto = await request.text();
  if (texto.length > 1024) {
    return new Response(null, { status: 413 });
  }
  try {
    const m = JSON.parse(texto) as Record<string, unknown>;
    if (typeof m.nombre === "string" && NOMBRES.has(m.nombre) && typeof m.valor === "number") {
      console.log(
        JSON.stringify({
          metrica: m.nombre,
          valor: m.valor,
          calificacion: typeof m.calificacion === "string" ? m.calificacion.slice(0, 20) : undefined,
          navegacion: typeof m.navegacion === "string" ? m.navegacion.slice(0, 20) : undefined,
          ruta: typeof m.ruta === "string" ? m.ruta.slice(0, 200) : undefined,
        }),
      );
    }
  } catch {
    // Un cuerpo que no es JSON no se registra.
  }
  return new Response(null, { status: 204 });
}
