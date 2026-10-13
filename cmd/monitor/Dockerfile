# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/monitor-service ./cmd/monitor

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates curl

WORKDIR /app
COPY --from=builder /bin/monitor-service /bin/monitor-service

EXPOSE 8080

ENTRYPOINT ["/bin/monitor-service"]
