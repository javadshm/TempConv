# Build stage (target linux/amd64 for GKE)
FROM --platform=linux/amd64 golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /tempconv .

# Run stage
FROM --platform=linux/amd64 alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /tempconv .
EXPOSE 8080
CMD ["./tempconv"]
