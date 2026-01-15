FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o judge-service ./cmd/server

FROM docker:24-dind

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/judge-service /usr/local/bin/
COPY sandbox/Dockerfile /sandbox/Dockerfile
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8081

ENTRYPOINT ["/entrypoint.sh"]
CMD ["judge-service"]
