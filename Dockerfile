# Stage 1: Build the Go binary
FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

# Copy dependency manifests and download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/main.go

# Stage 2: Final lightweight execution container
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the pre-built Linux binary and migrations from the builder stage
COPY --from=builder /app/api .
RUN chmod +x api
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the backend
CMD ["./api"]
