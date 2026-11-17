"use client";

import { useReportWebVitals } from "next/web-vitals";

// Web Vitals de usuarios reales (LCP, INP, CLS, FCP, TTFB). Se mandan al
// propio servidor de Next (/metricas), que las escribe en su log: sin
// servicios externos ni cookies. La ruta se generaliza para no registrar
// tokens ni ids (/seguimiento/[token], /cuenta/pedidos/[id]…).
export function rutaGeneral(ruta: string): string {
  return ruta
    .split("/")
    .map((s) => (s.length >= 20 || /^[0-9a-f-]{8,}$/i.test(s) ? "[id]" : s))
    .join("/");
}

export function MedicionWebVitals() {
  useReportWebVitals((m) => {
    const cuerpo = JSON.stringify({
      nombre: m.name,
      valor: Math.round(m.name === "CLS" ? m.value * 1000 : m.value),
      calificacion: m.rating,
      navegacion: m.navigationType,
      ruta: rutaGeneral(window.location.pathname),
    });
    if (navigator.sendBeacon) {
      navigator.sendBeacon("/metricas", new Blob([cuerpo], { type: "application/json" }));
    } else {
      void fetch("/metricas", { method: "POST", body: cuerpo, keepalive: true, headers: { "Content-Type": "application/json" } });
    }
  });
  return null;
}
