"use client";

import { useEffect } from "react";
import { ErrorDeCarga } from "@/components/error-de-carga";

export default function ErrorCheckout({
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
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-4 px-4 py-6">
      <ErrorDeCarga reintentar={retry} />
    </main>
  );
}
