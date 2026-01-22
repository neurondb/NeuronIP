FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO (required by some connectors like Oracle)
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled (required by Oracle driver)
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/neuronip-api ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/neuronip-api .

EXPOSE 8082

CMD ["./neuronip-api"]
