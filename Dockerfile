FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o lineserve-api .

# Use a minimal alpine image for the final image
FROM alpine:3.18

# Set working directory
WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/lineserve-api .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./lineserve-api"] 