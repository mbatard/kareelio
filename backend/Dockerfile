FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

RUN addgroup -S kareelio && adduser -S kareelio -G kareelio
USER kareelio

EXPOSE 8080

CMD ["./server"]
