import type { Metadata, Viewport } from "next";
import { Bodoni_Moda, Public_Sans } from "next/font/google";

import "./globals.css";

import { MedicionWebVitals } from "@/components/medicion-web-vitals";

// Fuentes alojadas en el mismo dominio (next/font las descarga en el build):
// sin pedidos a Google en cada visita, solo latin y solo los ejes variables.
const publicSans = Public_Sans({ subsets: ["latin"], variable: "--font-public-sans", display: "swap" });
const bodoni = Bodoni_Moda({ subsets: ["latin"], variable: "--font-bodoni", display: "swap", weight: ["500", "600"] });

export const metadata: Metadata = {
  title: { default: "BC Importados", template: "%s · BC Importados" },
  description: "Perfumes, cremas, body splash y electrónica importados, con retiro en el local o envío.",
};

export const viewport: Viewport = { width: "device-width", initialScale: 1, themeColor: "#f5f1f4" };

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="es-AR" className={`${publicSans.variable} ${bodoni.variable}`}>
      <body>
        <MedicionWebVitals />
        {children}
      </body>
    </html>
  );
}
