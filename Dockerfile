# Build stage
# Build stage
FROM docker.io/library/golang:1.24-alpine AS builder

# Instalar dependencias necesarias
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copiar go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar la aplicación
# Compilar la aplicación (garble eliminado por incompatibilidad con Go 1.24)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o server cmd/server/main.go

# Runtime stage
FROM docker.io/library/alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar el binario compilado
COPY --from=builder /app/server .

# Crear directorio para datos
RUN mkdir -p /root/data

# Exponer puerto
EXPOSE 8080

# Variables de entorno por defecto
ENV APP_PORT=8080
ENV APP_ENV=production
ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV DB_NAME=kerokero
ENV DB_USER=kerokero
ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

# Comando de inicio
CMD ["./server"]
