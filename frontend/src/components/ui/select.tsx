import * as React from "react"

import { NativeSelect } from "@/components/ui/native-select"
import { cn } from "@/lib/utils"

// El proyecto usa siempre el <select> nativo (AGENTS.md, «Mobile-first»):
// en el teléfono abre el selector del sistema y no depende de JavaScript.
// Los hijos son <option> comunes.
function Select({ className, ...props }: React.ComponentProps<typeof NativeSelect>) {
  return <NativeSelect className={cn("w-full", className)} {...props} />
}

export { Select }
