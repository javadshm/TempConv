# Build stage
FROM --platform=linux/amd64 golang:1.24-alpine AS builder

# Install protoc and required tools (add git)
RUN apk add --no-cache protobuf protobuf-dev curl unzip git

WORKDIR /app

# Go deps
COPY go.mod go.sum ./
RUN go mod download

# Install protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Copy source
COPY . .

# Generate protobuf code (robust googleapis lookup)
RUN set -e; \
    GW_ROOT=$(go env GOPATH)/pkg/mod; \
    GOOGLEAPIS=$(find "$GW_ROOT" -maxdepth 3 -type d -path "*/grpc-gateway@*/third_party/googleapis" | head -n1); \
    echo "GOPATH mod root: $GW_ROOT"; \
    echo "googleapis path: $GOOGLEAPIS"; \
    [ -n "$GOOGLEAPIS" ] || { echo "googleapis include not found"; find "$GW_ROOT" -maxdepth 4 -type d -name googleapis; exit 1; }; \
    protoc -I. -I="$GOOGLEAPIS" \
      --go_out=gen --go_opt=paths=source_relative \ 
      --go-grpc_out=gen --go-grpc_opt=paths=source_relative \ 
      --grpc-gateway_out=gen --grpc-gateway_opt=paths=source_relative \ 
      --grpc-gateway_opt=generate_unbound_methods=true \
      api/tempconv.proto

# Build server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /tempconv ./cmd/server

# Run stage
FROM --platform=linux/amd64 alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /tempconv .
EXPOSE 8080 9090
CMD ["./tempconv"]