FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN apk --no-cache add ca-certificates

COPY . .
ARG APP_VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X github.com/user/kareelio/backend/internal/model.Version=${APP_VERSION}" -o /app/server ./cmd/server

FROM scratch

WORKDIR /app

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER 1000:1000

EXPOSE 8080

CMD ["./server"]
