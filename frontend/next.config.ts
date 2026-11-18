import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  reactCompiler: true,
  // Fotos del propio sitio: AVIF primero y WebP si el navegador no lo
  // soporta. No hay remotePatterns: el optimizador no procesa fotos de otros
  // dominios hasta que se decida dónde se guardan (spec, sección 8).
  //
  // En Vercel el proyecto corre como servicio (vercel.json) con output
  // standalone, y ahí /_next/image no está conectado: responde 404 para
  // toda foto. Las fotos de producto ya se guardan en su formato final
  // (1200×1500 WebP), así que en Vercel se sirven tal cual. En la VPS
  // (Docker) el optimizador sigue activo.
  images: { formats: ["image/avif", "image/webp"], unoptimized: process.env.VERCEL === "1" },
};

export default nextConfig;
