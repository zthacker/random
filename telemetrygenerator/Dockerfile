# Stage 1: Build the Go application
FROM golang:1.20-alpine as builder

# Install git for dependency management if required
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application as a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o telemetry-generator .

# Stage 2: Create the runtime container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Install necessary packages (e.g., for networking)
RUN apk --no-cache add ca-certificates

# Copy the built binary from the builder stage
COPY --from=builder /app/telemetry-generator .

# Expose the UDP port if needed (optional for this use case)
EXPOSE 8089/udp

# Specify the default command (entrypoint) as a shell
ENTRYPOINT ["./telemetry-generator"]
