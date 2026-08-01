FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG APP_VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X github.com/user/kareelio/backend/internal/model.Version=${APP_VERSION}" -o /app/server ./cmd/server

FROM alpine:3.24

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

RUN addgroup -S kareelio && adduser -S kareelio -G kareelio
USER kareelio

EXPOSE 8080

CMD ["./server"]
