"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { api, ErrorApi } from "@/lib/api";

import { borrarSesion } from "./cookie";

// «Salir»: invalida la sesión en la API y borra la cookie. Si la sesión ya no
// valía (401), igual se borra la cookie: el resultado es el mismo.
export async function salirAction(): Promise<void> {
  try {
    await api("/api/auth/logout", { metodo: "POST" });
  } catch (e) {
    if (!(e instanceof ErrorApi && e.status === 401)) throw e;
  }
  await borrarSesion();
  revalidatePath("/", "layout");
  redirect("/");
}
