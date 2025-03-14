# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Install necessary tools and dependencies
RUN apk add --no-cache git curl unzip

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed.
RUN go mod download

# Download and install protoc
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v28.3/protoc-28.3-linux-x86_64.zip && \
    unzip protoc-28.3-linux-x86_64 -d /usr/local && \
    rm protoc-28.3-linux-x86_64.zip

# Install protocol buffer for go and grpc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Generate the gRPC controller
RUN protoc --go_out=. --go-grpc_out=. "api/token.proto"

# Build the Go app
RUN go build -o main .

# Stage 2: Run the application
FROM alpine:3.20.3

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the builder stage
COPY --from=builder /app/main .

# Expose the port the app runs on
EXPOSE 50051

# Command to run the executable
CMD ["./main"]
