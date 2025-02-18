# Используем официальный образ Go
FROM golang:1.23.1-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o appl .

# Создаем финальный образ
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем скомпилированное приложение из первого этапа
COPY --from=builder /app/appl .

COPY internal/storage/postgres/migrations ./migrations
# Порт, на котором будет работать приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./appl"]