# templet

[![Built with Gentle-AI](https://raw.githubusercontent.com/Gentleman-Programming/gentle-ai/main/docs/assets/brand/built-with-gentle-ai.png)](https://github.com/Gentleman-Programming/gentle-ai)

Generador de proyectos a partir de una plantilla declarada en TOML.

`templet` es **agnóstico del lenguaje**: una plantilla no es más que una lista de
archivos de texto con su ruta y su contenido. No asume Go, ni Node, ni Python;
sirve para cualquier proyecto hecho de archivos de texto.

```sh
templet -dest ./salida -var Name=demo -var ModulePath=example.com/demo docs/sample-go.toml
```

## Instalación

Desde el repositorio:

```sh
go install ./cmd/templet
```

O compilando el binario optimizado:

```sh
just build      # deja el binario en dist/<os>/templet
templet -version
```

## Uso

```
templet [opciones] <plantilla.toml>
```

| Opción | Descripción |
| --- | --- |
| `-dest <dir>` | Directorio donde se genera el proyecto. Por defecto `.` |
| `-var clave=valor` | Define una variable. Se puede repetir. El valor puede contener `=` |
| `-version` | Imprime la versión y termina |

Las variables declaradas en la plantilla que no lleguen por `-var` se
preguntan por consola. Los valores se usan **tal cual se escriben**: no se
normalizan ni se convierten a slug. Si necesitás un slug, pedilo en la
plantilla con `{{ slug .Name }}`.

Un `-var` sin `=`, o con la clave vacía, es un error: no se ignora en silencio.

## Formato de la plantilla

```toml
variables = ["Name", "ModulePath"]

[[files]]
path = "{{ .Name }}/README.md"
content = '''
# {{ .Name }}
'''
```

### `variables`

Lista de variables que la plantilla necesita de quien la usa. Es solo la lista
de preguntas: dentro de `path`, `content` y `when` se accede por nombre
(`{{ .Name }}`).

### `[[files]]`

Cada entrada describe un archivo destino:

| Campo | Obligatorio | Descripción |
| --- | --- | --- |
| `path` | sí | Ruta destino, relativa a `-dest`. Es una plantilla |
| `content` | sí | Contenido del archivo. Es una plantilla |
| `when` | no | Condición: si no es verdadera, el archivo no se genera |
| `delims` | no | Par de delimitadores propio de este archivo |
| `mode` | no | Permisos del archivo en octal, por ejemplo `"0755"`. Por defecto `0644` |
| `[[files.replaces]]` | no | Reemplazos literales aplicados al contenido ya renderizado |

### `[[files.replaces]]`

```toml
[[files.replaces]]
text = "ProjectName"
replace = "{{ .Name }}"
```

Se aplican **después** del render, con `strings.ReplaceAll` (texto literal, no
expresiones regulares). Sirven para literales que no tienen forma de
delimitador; para llaves literales usá `delims`.

## Delimitadores

Por defecto se usa la sintaxis de `text/template`: `{{ ... }}`.

Cuando el archivo generado tiene que contener llaves propias de otro sistema
de plantillas — expresiones `${{ ... }}` de GitHub Actions, `{{ }}` de Jinja,
`[[ ]]` de Helm — declarás otro par:

```toml
delims = ["[[", "]]"]        # a nivel plantilla: default para todos los archivos

[[files]]
path = "[[ .Name ]]/.github/workflows/ci.yml"
when = "[[ .WithCI ]]"
delims = ["[[", "]]"]        # a nivel archivo: gana sobre el de la plantilla
content = '''
name: [[ .Name ]] CI

steps:
  - run: echo ${{ github.ref }}
'''
```

**Regla importante**: el par de un archivo se aplica a su `path`, su `when` y su
`content`. Si un archivo declara `delims`, su ruta y su condición también se
escriben con esos delimitadores. Si te equivocás, el error te dice cuáles
estaban en efecto.

## Permisos de archivo

Cada archivo se escribe con permisos `0644` salvo que declares `mode`, un
literal **octal** de permisos, entre `0000` y `0777`:

```toml
[[files]]
path = "{{ .Name }}/scripts/build.sh"
mode = "0755"                 # ejecutable
content = '''
#!/bin/sh
set -eu
go build -o bin/{{ .Name }} .
'''
```

Un `mode` inválido (`"0999"`, `"755x"`, `"8888"`, `"4755"`) es un error que
nombra el archivo, y no se escribe nada: la validación ocurre antes de tocar el
disco. Los bits de setuid, setgid y sticky no se aceptan: una plantilla de
proyecto no debería emitir archivos con privilegios elevados.

## Inclusión condicional

`when` es una plantilla que se evalúa como booleano:

- vacío o ausente → el archivo siempre se genera;
- `true`, `t`, `1`, `yes`, `y`, `on` → se genera;
- `false`, `f`, `0`, `no`, `n`, `off` → se omite;
- cualquier otra cosa → error, y no se escribe nada.

Un archivo omitido se reporta como `Skip file` y **la generación continúa** con
el siguiente. Un error, en cambio, corta la generación.

## Funciones disponibles

Además de todo lo de `text/template`, en `path`, `content` y `when`:

| Función | Ejemplo | Resultado |
| --- | --- | --- |
| `upper` | `{{ upper .Name }}` | `DEMO` |
| `lower` | `{{ lower .Name }}` | `demo` |
| `title` | `{{ title "mi proyecto" }}` | `Mi Proyecto` |
| `slug` | `{{ slug "Mi Proyecto" }}` | `mi-proyecto` |
| `starts` / `ends` | `{{ if starts .Name "de" }}` | booleano |
| `contains` / `containsAny` | `{{ if contains .Name "-" }}` | booleano |
| `repeat` | `{{ repeat "=" 10 }}` | `==========` |
| `replace` | `{{ replace .Name "_" "-" }}` | `mi-proyecto` |

`slug` pasa a minúsculas, convierte espacios y `-` en un separador `-`,
pliega los acentos latinos (`Café` → `cafe`, `Straße` → `strasse`) y
translitera griego y cirílico (`Αθήνα` → `athina`, `Щука` → `shchuka`). Conserva
`-`, `/` y `.`, y descarta todo lo demás (incluido `_`). Después colapsa las
corridas de `-` y recorta las de los extremos. Si no sobrevive nada ASCII, el
resultado es un string vacío: con nombres solo en CJK (`日本語`) no cuenta con
`slug`, usá otro criterio para esos nombres.

**Ojo con las pipelines**: pasan el valor como **último** argumento. Para pasar
de un nombre a un identificador de paquete hay que anidar la llamada, no
encadenarla:

```
{{ replace (slug .Name) "-" "_" }}   # bien: hola-mundo -> hola_mundo
{{ slug .Name | replace "-" "_" }}   # mal: el patrón queda en el lugar del valor
```

## Cómo escribe los archivos

- Crea los directorios intermedios que falten (`0755`).
- Cada archivo se escribe con `mode` si lo declarás; si no, con `0644`.
- Escribe cada archivo a través de un temporal y lo renombra: un fallo de
  escritura **nunca** deja un archivo vacío ni truncado, y no pisa el contenido
  anterior si el render falla. Los permisos se aplican al temporal, así que
  sobreviven al renombre.
- Los reportes (`Generate file`, `Skip file`) van a **stdout**; los errores, a
  **stderr**.

## Códigos de salida

| Código | Situación |
| --- | --- |
| `0` | Generación correcta, o prompt cancelado por el usuario |
| `1` | Error de argumentos, de plantilla o de escritura |
| `2` | Opción desconocida para el CLI |

## Ejemplos completos

Hay dos plantillas de ejemplo, una por lenguaje, para mostrar que el generador
no asume nada del proyecto que emite: la misma herramienta produce un servicio
Go y un paquete Python a partir de dos listas de archivos.

### Servicio Go

[`docs/sample-go.toml`](docs/sample-go.toml) genera un servicio Go mínimo y
compilable, con un workflow de CI opcional:

```sh
templet -dest /tmp -var Name=demo -var ModulePath=example.com/demo -var WithCI=yes docs/sample-go.toml
```

Produce `README.md`, `go.mod`, `main.go`, `scripts/build.sh` (ejecutable) y
`.github/workflows/ci.yml`, este último solo si `WithCI` es verdadero y
conservando sus expresiones `${{ ... }}` de GitHub Actions literales.

### Paquete Python

[`docs/sample-python.toml`](docs/sample-python.toml) genera un paquete Python
en `src/`, con `pyproject.toml`, tests de pytest, script de tests ejecutable y
CI y `Dockerfile` opcionales:

```sh
templet -dest /tmp -var Name=demo -var Description="CLI de ejemplo" -var WithCI=yes -var WithDocker=no docs/sample-python.toml
```

Produce `demo/pyproject.toml` (PEP 621, con entry point de consola),
`demo/src/demo/{__init__,__main__}.py`, `demo/tests/test_demo.py`,
`demo/README.md`, `demo/.gitignore` y `demo/scripts/test.sh` (ejecutable). El
workflow `demo/.github/workflows/ci.yml` aparece solo con `WithCI` verdadero y
el `demo/Dockerfile` solo con `WithDocker` verdadero.

## Desarrollo

```sh
just check     # go mod tidy, vet, tests, golangci-lint y go mod verify
just build     # binario comprimido con upx en dist/<os>/
just run       # ejecuta el sample Go generando en dist/<os>/sample
just run template="docs/sample-python.toml"   # lo mismo con otra plantilla
```
