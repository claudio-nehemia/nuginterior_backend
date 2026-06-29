FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the pre-built Linux binary and migrations
COPY api .
COPY migrations ./migrations

# Expose port
EXPOSE 8080

# Run the backend
CMD ["./api"]
