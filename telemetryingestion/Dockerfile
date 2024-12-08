# Stage 1: Build the Go application with a statically linked binary
FROM golang:1.22-alpine as builder

# Install git to pull in dependencies if needed
RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY ../go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project directory into the container
COPY telemetryingestion ./telemetryingestion

# Set the working directory to the main command directory for building the binary
WORKDIR /app/telemetryingestion/cmd/telemetryingestion

# Build the Go app statically linked
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /telemetryingestion .

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Install CA certificates to enable secure database connections
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder container
COPY --from=builder /telemetryingestion .

# Expose the UDP port used by the application
EXPOSE 8089/udp

# Run the Go application
CMD ["./telemetryingestion"]
