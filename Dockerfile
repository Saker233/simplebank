# Build Stage
FROM golang:1.25.11-alpine3.23 AS builder

WORKDIR /app

COPY . .

# Build the Go application
RUN go build -o main main.go

# Download the migrate binary
RUN apk add --no-cache curl tar && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-amd64.tar.gz \
        -o migrate.tar.gz && \
    tar -xzf migrate.tar.gz && \
    chmod +x migrate && \
    rm migrate.tar.gz

# Runtime Stage
FROM alpine:3.23

WORKDIR /app

COPY --from=builder /app/main /app/main
COPY --from=builder /app/migrate /app/migrate

COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./migration

RUN chmod +x /app/start.sh /app/migrate

EXPOSE 8080

ENTRYPOINT ["/app/start.sh"]