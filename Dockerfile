# Stage 1: Build stage
FROM golang:1.23 AS builder

WORKDIR /app

# Disable CGO and set target OS/arch
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
COPY .env .env
COPY wait-for-it.sh wait-for-it.sh
RUN go mod download

# Copy the full source code including all folders and Swagger docs
COPY . .

# Install swag and generate Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init --generalInfo cmd/main.go --output docs

# Build the Go binary statically
RUN go build -o app-server ./cmd/main.go

# Make the wait-for-it.sh script executable in the builder stage
RUN chmod +x /app/wait-for-it.sh

# Stage 2: Run stage
# Temporarily use Alpine for debugging (with shell support)
FROM alpine:latest

WORKDIR /app

# Install bash (if needed)
RUN apk add --no-cache bash

# Copy necessary files from the builder stage
COPY --from=builder /app/app-server /app/server
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/.env .env
COPY --from=builder /app/database/migrations ./database/migrations

# Copy the wait-for-it.sh script from the builder stage
COPY --from=builder /app/wait-for-it.sh /app/wait.sh

EXPOSE 8080

# Set entrypoint to wait for DB and start the server
ENTRYPOINT ["/app/wait.sh", "db:5432", "--timeout=60", "--", "/app/server"]
