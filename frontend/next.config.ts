import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  reactCompiler: true,
  // Fotos del propio sitio: AVIF primero y WebP si el navegador no lo
  // soporta. No hay remotePatterns: el optimizador no procesa fotos de otros
  // dominios hasta que se decida dónde se guardan (spec, sección 8).
  images: { formats: ["image/avif", "image/webp"] },
};

export default nextConfig;
