# Templet

**Templet** es una herramienta CLI escrita en Go para generar proyectos a partir de plantillas (templates). Permite clonar repositorios git o copiar directorios locales, renderizar archivos con Go templates, y generar proyectos configurables mediante un sistema de variables interactivo.

---

## Características

- **Múltiples fuentes de plantillas**: Soporta repositorios Git (GitLab, GitHub, etc.) y directorios locales
- **Sistema de variables interactivo**: Define variables en un archivo `meta.hjson` y obtén los valores del usuario mediante prompts interactivos
- **Renderizado con Go templates**: Archivos `.tmpl` se procesan con `text/template` de Go, usando la sintaxis de
  [Go Templates](https://pkg.go.dev/text/template), pero los delimitadores son `{%` y `%}` en vez de `{{` y `}}`
- **Reemplazo de placeholders**: Nombres de archivos y directorios con prefijo `@` se reemplazan con valores de variables
- **Configuración flexible**: Define tus propios proveedores de templates en el archivo de configuración

##  Uso

### Sintaxis básica

```bash
tpl -n <nombre-proyecto> [-r repo_dir] <tipo>:<path>
```

- `-n,--name`: Nombre del proyecto - obligatorio
- `-r,--repo`: Ruta del repositorio git donde descargar la plantilla (host[:puerto], github.com, gitlab.com, ...) - opcional,
  obligatorio si `<tipo>` es `git`

- `<tipo>`: Tipo de plantilla (`file`/`git`)
- `<path>`: Si `typo` es `git` entonces `<path>` es `usuario/repositorio`, si `typo` es `file` entonces `<path>` es la ruta
  absoluta del directorio de la plantilla


### Ejemplos

#### Desde un repositorio Git (GitLab)

```bash
tpl -n mi-api gitlab:grupo/repository-name
```

#### Desde un directorio local

```bash
tpl -n mi-proyecto file:/ruta/absoluta/a/plantilla
```

##  Crear una plantilla

### 1. Estructura básica

```
mi-plantilla/
├── meta.hjson              # Configuración de variables
├── archivo.txt.tmpl        # Archivo con template
├── @BinName/              # Directorio que usa variable
│   └── main.go.tmpl
└── static.md              # Archivo copiado sin cambios
```

### 2. Archivo `meta.hjson`

Define las variables que se solicitarán al usuario:

```hjson
{
  nombre: "Mi plantilla"
  descripcion: "Una plantilla de ejemplo"
  variables: [
    {
      nombre: "BinName"
      descripcion: "Nombre del binario"
    }
    {
      nombre: "driver"
      descripcion: "Driver de base de datos"
      opciones: [
        "pgsql"
        "mysql"
        "sqlite"
      ]
    }
  ]
}
```

**Tipos de variables:**

- **Texto libre**: Prompt simple (sin campo `opciones`)
- **Selección múltiple**: Menú de opciones (con campo `opciones`)

### 3. Archivos `.tmpl`

Los archivos con extensión `.tmpl` se renderizan usando Go templates:

```go
// main.go.tmpl
package main

const BinName = "{% .BinName %}"
const Driver = "{% .driver %}"

func main() {
    println("Proyecto:", BinName)
}
```

### 4. Placeholders en nombres

Usa `@` como prefijo para reemplazar nombres de archivos/directorios:

```
@BinName/cmd/main.go  → mi-api/cmd/main.go
config-@env.yaml      → config-production.yaml
```
