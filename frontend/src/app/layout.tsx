import type { Metadata, Viewport } from "next";
import { Jost } from "next/font/google";

import "./globals.css";

import { MedicionWebVitals } from "@/components/medicion-web-vitals";

// Una sola familia (DESIGN.md, sección 4): Jost, geométrica y variable.
// next/font la descarga en el build y la sirve desde el mismo dominio: sin
// pedidos a Google en cada visita, solo latin.
const jost = Jost({ subsets: ["latin"], variable: "--font-jost", display: "swap" });

export const metadata: Metadata = {
  title: { default: "BC Importados", template: "%s · BC Importados" },
  description: "Perfumes, cremas, body splash y electrónica importados, con retiro en el local o envío.",
};

export const viewport: Viewport = { width: "device-width", initialScale: 1, themeColor: "#ffffff" };

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="es-AR" className={jost.variable}>
      <body>
        <MedicionWebVitals />
        {children}
      </body>
    </html>
  );
}
