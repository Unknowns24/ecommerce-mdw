import Link from "next/link";

// Migas de pan: dónde estoy dentro del catálogo y cómo vuelvo.
export function Migas({ pasos }: { pasos: { texto: string; href?: string }[] }) {
  return (
    <nav aria-label="Migas de pan" className="text-[0.8125rem] text-muted-foreground">
      <ol className="flex flex-wrap items-center gap-1.5">
        {pasos.map((p, i) => (
          <li key={`${p.texto}-${i}`} className="flex items-center gap-1.5">
            {i > 0 ? <span aria-hidden="true">/</span> : null}
            {p.href ? (
              <Link href={p.href} className="underline-offset-4 hover:text-foreground hover:underline">
                {p.texto}
              </Link>
            ) : (
              <span aria-current="page" className="text-foreground">
                {p.texto}
              </span>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}
