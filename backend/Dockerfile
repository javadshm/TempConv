# Dockerfile

FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

# other necessary commands for setup...

# Begin the proto generation block
RUN mkdir -p gen; \
    protoc -I. -I${GOPATH}/src \
    --go_out=plugins=grpc:${GOPATH}/src \
    --go_opt=paths=source_relative \
    --proto_path=$(GOOGLEAPIS) \
    --proto_path=. \
    <proto_files>
# End of proto generation block

# other remaining commands...
