# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/payment-service ./cmd/payment-service

# Runtime stage
FROM alpine:3.20

WORKDIR /app
COPY --from=builder /bin/payment-service /bin/payment-service

EXPOSE 8081

ENTRYPOINT ["/bin/payment-service"]
