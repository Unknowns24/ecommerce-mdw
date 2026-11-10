import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "BC Importados",
  description: "Tienda de BC Importados",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="es-AR">
      <body>{children}</body>
    </html>
  );
}
