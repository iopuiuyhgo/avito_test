FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o appl .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/appl .

COPY internal/storage/postgres/migrations ./migrations
EXPOSE 8080

CMD ["./appl"]