import { defineConfig, devices } from "@playwright/test";

import { API_URL, BINARIOS, entornoApi, env, exigirBaseLocal, WEB_URL } from "./e2e/entorno";

exigirBaseLocal(env.DATABASE_DSN);


// Orden real de Playwright: primero levanta los webServer y espera sus URLs;
// después corre globalSetup (e2e/preparar.ts), que resetea la base local,
// migra, siembra y crea las cuentas de prueba por la API.
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: 0,
  reporter: process.env.CI ? [["list"], ["html", { open: "never" }]] : "list",
  globalSetup: "./e2e/preparar.ts",
  outputDir: "test-results",
  use: { trace: "retain-on-failure", screenshot: "only-on-failure" },
  projects: [
    {
      // Sin navegador: le pega directo a la API (docs/api.md).
      name: "permisos",
      testMatch: /permisos\.spec\.ts|sesiones\.spec\.ts/,
      use: { baseURL: API_URL },
    },
    {
      name: "flujo",
      testMatch: /flujo-principal\.spec\.ts/,
      use: { ...devices["Desktop Chrome"], baseURL: WEB_URL },
    },
  ],
  webServer: [
    {
      command: `cd ../backend && go build -o ${BINARIOS}/ ./cmd/... && ${BINARIOS}/api`,
      url: `${API_URL}/api/salud`,
      env: entornoApi(),
      timeout: 180_000,
      reuseExistingServer: false,
      stdout: "ignore",
      stderr: "pipe",
    },
    {
      // Como en producción (frontend/Dockerfile): el server.js del build standalone.
      command:
        "npx next build && cp -R public .next/standalone/ && cp -R .next/static .next/standalone/.next/ && node .next/standalone/server.js",
      // /ingresar no toca la base sin sesión: el chequeo de «listo» no se
      // cruza con el reseteo del esquema que hace preparar.ts.
      url: `${WEB_URL}/ingresar`,
      env: { API_BASE_URL: API_URL, NODE_ENV: "production", PORT: env.WEB_PORT, HOSTNAME: "localhost" },
      timeout: 300_000,
      reuseExistingServer: false,
      stdout: "ignore",
      stderr: "pipe",
    },
  ],
});
