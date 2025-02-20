FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bot cmd/bot/main.go

FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /app/bot ./
CMD ["/app/bot"]