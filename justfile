# Justfile para el proyecto Templet
ext := if os_family() == "windows" { ".exe" } else { "" }
bin := "templet" + ext
dirDist := "./dist/" + os_family()
version := shell("git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0'")
ldflags := "-s -w -X main.Version=" + version

# Realiza comprobaciones de control de calidad en el código fuente
check:
    go mod tidy
    go vet ./...
    go test ./...
    golangci-lint run ./...
    go mod verify

# Compila el programa
build:
    @rm -rf {{dirDist}}
    mkdir -p {{dirDist}}
    go build -o {{dirDist}}/{{bin}} -ldflags "{{ldflags}}" ./cmd/templet
    upx --best {{dirDist}}/{{bin}}
    ls -lh {{dirDist}}/{{bin}}

# Ejecuta el programa a traves del comando `go run`, generando en dist/sample para no ensuciar el repositorio
run:
    go run ./cmd/templet -dest {{dirDist}}/sample docs/sample.toml
