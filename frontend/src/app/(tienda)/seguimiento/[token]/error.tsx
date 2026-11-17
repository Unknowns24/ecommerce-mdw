"use client";

import { useEffect } from "react";
import { ErrorDeCarga } from "@/components/error-de-carga";

export default function ErrorSeguimiento({
  error,
  retry,
}: {
  error: Error & { digest?: string };
  retry: () => void;
}) {
  useEffect(() => {
    // El detalle va al log; la pantalla nunca muestra error.message.
    console.error(error);
  }, [error]);

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-4">
      <ErrorDeCarga reintentar={retry} />
    </div>
  );
}
