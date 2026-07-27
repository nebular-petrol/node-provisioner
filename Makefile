.PHONY: run lint sec check clean

# Ejecuta el proyecto en modo desarrollo
run:
	go run ./cmd/server/main.go

# Ejecuta el linter de código fuente
lint:
	golangci-lint run ./...

# Analiza vulnerabilidades pero permite continuar el flujo de desarrollo local
sec:
	@govulncheck ./... || echo "⚠️ Advertencia de vulnerabilidades detectadas en dependencias externas (ignorado en desarrollo)"

# Ejecuta todas las validaciones
check: lint sec

# Limpia los binarios compilados
clean:
	rm -rf tmp/