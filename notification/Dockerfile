# Build stage
FROM --platform=linux/arm64 golang:1.23-alpine AS builder
WORKDIR /app

# RUN chmod +x notification

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

# Explicitly set target OS/arch
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o notification .
RUN go build -o notification .

# Final stage
FROM --platform=linux/arm64 alpine:latest
WORKDIR /app

COPY --from=builder /app/notification .

EXPOSE 8080

# RUN chmod +x notification

CMD ["./notification"]
