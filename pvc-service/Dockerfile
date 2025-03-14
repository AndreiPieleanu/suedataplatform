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

# Clone the manifests repository
RUN git clone https://github.com/kubeflow/manifests.git && \
    cd manifests && \
    git checkout 8634c24 && \
    cd ..

# Generate the gRPC controller
RUN protoc --go_out=. --go-grpc_out=. "api/pvc.proto"

# Build the Go app
RUN go build -o main .

# Stage 3: Run the application
FROM alpine:3.20.3

# Install the AWS CLI
RUN apk add --no-cache aws-cli

# Set the Current Working Directory inside the container
WORKDIR /app

# Accept AWS credentials as build arguments
ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY
ARG AWS_DEFAULT_REGION

# Set environment variables with the provided arguments
ENV AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
ENV AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
ENV AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION}

# # Connect the container to kubeflow cluster
RUN aws eks update-kubeconfig --name cluster-1 --region eu-north-1

# Copy the Pre-built binary file from the builder stage
COPY --from=builder /app/main .

# Expose the port the app runs on
EXPOSE 50052

# Command to run the executable
CMD ["./main"]
