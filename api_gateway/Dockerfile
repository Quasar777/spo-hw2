FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o api-gateway .cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-gateway .

EXPOSE 8000

CMD ["./api-gateway"]
