"use client";

import { useEffect } from "react";

import { ErrorDeCarga } from "@/components/error-de-carga";

// El detalle técnico va a la consola, nunca a la pantalla (sección 7).
export default function ErrorCatalogo({ error, retry }: { error: Error & { digest?: string }; retry: () => void }) {
  useEffect(() => {
    console.error(error);
  }, [error]);
  return <ErrorDeCarga reintentar={retry} />;
}
