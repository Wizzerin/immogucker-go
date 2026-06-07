# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Собираем с флагами для статической линковки
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o immogucker-api ./cmd/api

# Final stage
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/immogucker-api .
COPY --from=builder /app/migrations ./migrations
# Копируем шаблоны туда, где их ожидает Go (в подпапку web)
COPY --from=builder /app/web ./web

EXPOSE 8080
ENTRYPOINT ["./immogucker-api"]
