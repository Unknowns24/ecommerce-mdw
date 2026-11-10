import Link from "next/link";
import type { ReactElement } from "react";
import { buttonVariants } from "@/components/ui/button";

export async function MenuSesion(): Promise<ReactElement> {
  return <Link className={buttonVariants({ variant: "ghost" })} href="/ingresar">Ingresar</Link>;
}
