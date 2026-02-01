#!/bin/bash
set -e

echo "Generating Go code from protobuf definitions..."

# Ensure we're in the backend directory
cd "$(dirname "$0")"

# Get the grpc-gateway path for googleapis annotations
GRPC_GATEWAY_PATH=$(go list -f '{{ .Dir }}' -m github.com/grpc-ecosystem/grpc-gateway/v2 2>/dev/null || echo "")

if [ -z "$GRPC_GATEWAY_PATH" ]; then
  echo "Error: grpc-gateway module not found. Run 'go mod download' first."
  exit 1
fi

# Create output directory
mkdir -p gen/api

# Generate code
protoc -I. -Ithird_party \
  --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=gen --grpc-gateway_opt=paths=source_relative \
  --grpc-gateway_opt=generate_unbound_methods=true \
  api/tempconv.proto

echo "Code generation complete! Generated files in gen/"
