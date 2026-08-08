FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
ARG VITE_API_URL=/
RUN VITE_API_URL=$VITE_API_URL npm run build

FROM golang:1.26-alpine AS server-builder

WORKDIR /app

COPY server.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/frontend-server ./server.go

FROM scratch

WORKDIR /app

ENV PORT=8080

COPY --from=builder /app/dist ./dist
COPY --from=server-builder /out/frontend-server ./frontend-server

USER 101:101

EXPOSE 8080

CMD ["./frontend-server"]
