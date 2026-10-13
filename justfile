# Justfile para el proyecto Templet
ext := if os_family() == "windows" { ".exe" } else { "" }
bin := "templet" + ext
dirDist := "./dist/" + os_family()
version := shell("git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0'")

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
    go build -o {{dirDist}}/{{bin}} -ldflags "-s -w -X main.Version={{version}}" cmd/templet/main.go
    upx --best {{dirDist}}/{{bin}}
    ls -lh {{dirDist}}/{{bin}}

# Ejecuta el programa en modo live, se recompila y lo recarga cada vez que se modifica el código fuente
live:
    air --build.cmd "go build -o {{dirDist}}/{{bin}} -ldflags '-X main.Version=live' cmd/api/main.go" --build.entrypoint "{{dirDist}}/{{bin}}" -- config/config.json

# Ejecuta el programa a traves del comando `go run`)
run:
    go run cmd/templet/main.go sample.toml
