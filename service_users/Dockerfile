# Этап сборки
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Сначала модули (оптимизация кеша)
COPY go.mod go.sum ./
RUN go mod download

# Потом остальной код
COPY . .

RUN go build -o users-service ./main.go

# Этап рантайма
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/users-service .

EXPOSE 8000

CMD ["./users-service"]
