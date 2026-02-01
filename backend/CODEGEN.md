# Code Generation

This document describes how to generate Go code from protobuf definitions.

## Prerequisites

Install the required tools:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
```

You also need `protoc` (the Protocol Buffer compiler):
- Download from: https://github.com/protocolbuffers/protobuf/releases
- Or install via package manager (e.g., `brew install protobuf` on macOS, `apt-get install protobuf-compiler` on Ubuntu)

## Generate Code

Run the following command from the `backend` directory:

```bash
protoc -I. -I$(go env GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/v2 \
  --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  --grpc-gateway_out=gen --grpc-gateway_opt=paths=source_relative \
  --grpc-gateway_opt=generate_unbound_methods=true \
  api/tempconv.proto
```

Or use the included script:
```bash
./generate.sh
```

This generates:
- `gen/api/tempconv.pb.go` - Protocol buffer message definitions
- `gen/api/tempconv_grpc.pb.go` - gRPC service definitions
- `gen/api/tempconv.pb.gw.go` - grpc-gateway HTTP reverse proxy
