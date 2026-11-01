---
name: git
description: Convenciones de Git, formato de mensajes de commit prefijados con la rama actual y confirmación explícita previa a cualquier commit.
---

# Git Workflow & Commit Conventions Skill

Usa esta skill al gestionar ramas, preparar cambios, redactar mensajes de commit y realizar operaciones de Git en este repositorio.

## 1. Formato Obligatorio del Mensaje de Commit

Todo commit **DEBE** comenzar estrictamente con el nombre exacto de la rama en la que se está trabajando, seguido de `: ` y la descripción concisa del cambio en presente o imperativo.

### Estructura:

```text
"<nombre-de-rama>: <descripción del cambio>"
```

### Ejemplos según la rama:

* **Rama de tarea (`TASK-...`)**:
  - Rama: `TASK-001`
  - Commit: `"TASK-001: implement health checks for database"`

* **Rama de característica (`feat/...`)**:
  - Rama: `feat/integrate-go-utils`
  - Commit: `"feat/integrate-go-utils: add git convention skill and register in agents instructions"`

* **Rama de corrección (`fix/...`)**:
  - Rama: `fix/read-timeout`
  - Commit: `"fix/read-timeout: adjust server timeout to prevent early disconnect"`

---

## 2. Confirmación Explícita del Usuario (Regla de Oro)

1. **Nunca hacer commit automáticamente**:
   - Los agentes **NUNCA** deben ejecutar `git commit` por iniciativa propia ni como parte de una cadena de tareas sin solicitud previa.
2. **Dejar cambios listos en el Working Tree**:
   - Tras realizar las modificaciones de código y validar los tests, los cambios deben permanecer preparados y visibles en el área de trabajo.
3. **Esperar la orden del usuario**:
   - Solo se procederá a commitear y hacer push cuando el usuario lo pida expresamente (ej: *"haz commit"*, *"commitea y sube"*).

---

## 3. Flujo de Ramas

- **Ramas de trabajo**: Todo desarrollo o corrección debe realizarse en una rama dedicada (`feat/*`, `fix/*`, `TASK-*`).
- **Ramas protegidas**: No trabajar ni commitear directamente sobre `main` o `master`.
- **Detección de rama**: Antes de redactar el mensaje de commit, verificar la rama activa ejecutando `git branch --show-current`.
