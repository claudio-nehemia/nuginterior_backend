# Step 1: Build the Go binary
FROM golang:1.25-alpine AS builder

# Install git and certificates
RUN apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the application using Go cache mounts to speed up compilation
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/api ./cmd/api/main.go

# Step 2: Run the binary
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary and migrations from builder
COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the backend
CMD ["./api"]
