# Build stage (target linux/amd64 for GKE)
FROM --platform=linux/amd64 golang:1.24-alpine AS builder

# Install protoc and required tools
RUN apk add --no-cache protobuf protobuf-dev curl unzip

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Install protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Copy source code
COPY . .

# Generate protobuf code
RUN set -e; \
    GATEWAY=$(go env GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v*/third_party/googleapis; \
    protoc -I. -I="$GATEWAY" \
      --go_out=gen --go_opt=paths=source_relative \
      --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
      --grpc-gateway_out=gen --grpc-gateway_opt=paths=source_relative \
      --grpc-gateway_opt=generate_unbound_methods=true \
      api/tempconv.proto

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /tempconv ./cmd/server

# Run stage
FROM --platform=linux/amd64 alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /tempconv .
EXPOSE 8080 9090
CMD ["./tempconv"]
