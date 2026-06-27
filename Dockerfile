FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o cloudprobe-dashboard ./cmd/dashboard
RUN go build -ldflags="-s -w" -o cloudprobe-agent ./cmd/agent

FROM node:20-alpine AS web-builder
WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ .
RUN npm run build

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/cloudprobe-dashboard /app/cloudprobe-agent /app/
COPY --from=web-builder /app/web/dist /app/web/dist

RUN mkdir -p /app/data /app/config /etc/cloudprobe

ENV TZ=Asia/Shanghai
EXPOSE 8000 50051

ENTRYPOINT ["/app/cloudprobe-dashboard"]
