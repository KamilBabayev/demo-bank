# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and all source files
COPY go.mod ./
COPY go.sum* ./
COPY . .

# Download dependencies and build
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o transfer .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary and migrations from builder
COPY --from=builder /app/transfer .
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

CMD ["./transfer"]
