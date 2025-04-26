# Этап сборки
FROM golang:1.20-alpine AS builder

# Установка рабочей директории
WORKDIR /app

# Копирование файлов модуля и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN go build -o cert-fetcher cert-fetcher.go

# Финальный образ
FROM alpine:latest

# Установка корневых сертификатов
RUN apk add --no-cache ca-certificates

# Установка рабочей директории
WORKDIR /root/

# Копирование скомпилированного бинарного файла из этапа сборки
COPY --from=builder /app/cert-fetcher .

# Открытие порта
EXPOSE 8080

# Команда запуска
CMD ["./cert-fetcher"]
