# ADR — Architecture Decision Records

Los ADR registran decisiones técnicas aceptadas para la versión inicial. Cada uno explica el contexto, la decisión, sus consecuencias y qué queda fuera de la decisión. No reemplazan las reglas de negocio de [specs.md](../specs.md).

| ADR | Estado | Decisión |
| --- | --- | --- |
| [ADR-001](ADR-001-configuracion.md) | Aceptado | Configuración centralizada con Viper |
| [ADR-002](ADR-002-persistencia-y-migraciones.md) | Aceptado | GORM y migraciones versionadas con gormigrate |
| [ADR-003](ADR-003-api-http.md) | Aceptado | API HTTP con Chi |

Una decisión nueva o una modificación incompatible requiere un ADR nuevo que reemplace o deje obsoleto al anterior. Los temas aún sujetos a análisis se documentan como RFC, no como ADR.
