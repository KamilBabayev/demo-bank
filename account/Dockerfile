# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and source files
COPY go.mod ./
COPY *.go ./
RUN go mod tidy && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o account .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/account .

# Expose port
EXPOSE 8080

CMD ["./account"]
