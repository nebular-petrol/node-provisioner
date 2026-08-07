# Etapa 1: Compilacion
FROM golang:1.27rc2-alpine3.23 AS builder

WORKDIR /app

# Descargar dependencias aprovechando la cache
COPY go.mod go.sum ./
RUN go mod download

# Copiar todo el codigo fuente
COPY . .

# Compilar el servidor
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Etapa 2: Imagen final
FROM alpine:latest

WORKDIR /app

# Copiar ejecutable
COPY --from=builder /app/server .

# Copiar certificados
COPY certs ./certs

EXPOSE 8080

CMD ["./server"]
