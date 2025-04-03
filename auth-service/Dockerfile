# Білдер для компіляції
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service

# Фінальний образ
FROM alpine:latest
WORKDIR /app
COPY --from=builder /auth-service /app/auth-service
COPY --from=builder /app/configs /app/configs  # Якщо є конфіги

# Вказуємо порт та команду для запуску
EXPOSE 8080
CMD ["/app/auth-service"]