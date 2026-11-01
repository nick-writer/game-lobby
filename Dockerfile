# --- Этап 1: Сборка (Builder) ---
FROM golang:1.27-alpine AS builder
WORKDIR /app

# Копируем только файлы зависимостей, чтобы использовать кэш Docker
COPY go.mod go.sum* ./
RUN go mod download

# Копируем исходный код и собираем статический бинарник
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# --- Этап 2: Запуск (Runtime) ---
FROM alpine:latest
# Устанавливаем сертификаты
RUN apk --no-cache add ca-certificates
WORKDIR /app

# Копируем ТОЛЬКО скомпилированный бинарник из этапа сборки
COPY --from=builder /app/server .
# Копируем папку со статикой
COPY static/ ./static/ 

EXPOSE 8080
CMD ["./server"]