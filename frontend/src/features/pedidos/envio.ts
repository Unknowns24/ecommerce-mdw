import type { ConfiguracionTienda } from "./tipos";

// ESTIMACIÓN del envío para mostrar antes de confirmar. El costo que vale es
// el que calcula POST /api/pedidos con la misma configuración (spec, sección 6,
// «Entrega»): rangos [desde, hasta) en metros, sin redondear, y la distancia
// igual al máximo configurado sí está cubierta y paga el último rango.
export type EstimacionEnvio =
  | { tipo: "costo"; centavos: number }
  | { tipo: "fuera-de-cobertura" }
  | { tipo: "sin-tarifa" };

export function estimarEnvio(config: ConfiguracionTienda, metros: number): EstimacionEnvio {
  if (metros > config.distanciaMaximaMetros) return { tipo: "fuera-de-cobertura" };
  const ordenadas = [...config.tarifas].sort((a, b) => a.desdeMetros - b.desdeMetros);
  const tarifa = ordenadas.find((t) => metros >= t.desdeMetros && metros < t.hastaMetros);
  if (tarifa) return { tipo: "costo", centavos: tarifa.costoCentavos };
  const ultima = ordenadas[ordenadas.length - 1];
  if (ultima && metros === config.distanciaMaximaMetros && metros === ultima.hastaMetros) {
    return { tipo: "costo", centavos: ultima.costoCentavos };
  }
  return { tipo: "sin-tarifa" };
}
