FROM golang:1.23-alpine AS builder
WORKDIR /app

# Copy go mod files first (better Docker caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o accessibility-api main.go

# Final stage - minimal image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/accessibility-api .

# Copy .env.example (optional, for reference)
COPY .env.example* ./

# Port will be read from .env file
EXPOSE ${PORT:-8080}

# Run the binary
CMD ["./accessibility-api"]