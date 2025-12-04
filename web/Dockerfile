# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY main.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o miniflux-reader .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/miniflux-reader .

# Copy templates and static files
COPY templates ./templates
COPY static ./static

# Default port (can be overridden with PORT environment variable at runtime)
ENV PORT=8080

# Expose port (can be overridden with -p flag)
EXPOSE 8080

# Run the application
CMD ["./miniflux-reader"]
