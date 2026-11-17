"use client";

import Link from "next/link";
import { useEffect } from "react";

import "./globals.css";

import { ErrorDeCarga } from "@/components/error-de-carga";

// Red de seguridad para los errores que ocurren en un layout (por encima de
// los error.tsx de cada sección). Reemplaza al layout raíz: lleva su propio
// <html> y <body>. Nunca muestra el mensaje técnico.
export default function ErrorGlobal({ error, retry }: { error: Error & { digest?: string }; retry: () => void }) {
  useEffect(() => {
    console.error(error);
  }, [error]);
  return (
    <html lang="es-AR">
      <body>
        <main className="mx-auto grid max-w-xl gap-4 px-4 py-10">
          <Link href="/" className="font-semibold text-primary underline-offset-4 hover:underline">
            BC Importados
          </Link>
          <ErrorDeCarga reintentar={retry} />
        </main>
      </body>
    </html>
  );
}
