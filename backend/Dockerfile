# Многоэтапная сборка для оптимизации размера образа
FROM golang:1.23-alpine AS builder

# Установка ca-certificates
RUN apk update && apk add --no-cache ca-certificates

# Создание директории приложения
WORKDIR /app

# Копирование go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./

# Скачивание зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Финальный образ
FROM alpine:latest

# Установка ca-certificates для HTTPS запросов
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копирование бинарного файла из builder stage
COPY --from=builder /app/main .

# Открытие порта
EXPOSE 8080

# Запуск приложения
CMD ["./main"]