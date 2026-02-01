# Updated Dockerfile

# Base image
FROM golang:1.16 as builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Set up the gRPC gateway version
ENV GATEWAY=$(go env GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@$(go list -m -f '{{.Version}}' github.com/grpc-ecosystem/grpc-gateway/v2)/third_party/googleapis

# Copy the source code
COPY . .

# Run protoc with the new include path
RUN protoc -I. -I"$GATEWAY" --go_out=plugins=grpc:. *.proto

# Build the final image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app .

# Command to run the application
CMD ["./your-app-binary"]