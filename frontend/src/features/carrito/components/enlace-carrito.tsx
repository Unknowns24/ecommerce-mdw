import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";

export function EnlaceCarrito() {
  return <Link className={buttonVariants({ variant: "ghost" })} href="/carrito">Carrito</Link>;
}
