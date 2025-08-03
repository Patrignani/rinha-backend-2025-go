# Stage 1: Build
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main main.go

# Stage 2: Runtime (Alpine)
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/main .
RUN chmod +x main

EXPOSE 8080

CMD ["./main"]
