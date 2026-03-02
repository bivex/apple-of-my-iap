FROM golang:1.23-alpine AS builder
WORKDIR /app
# Copy all source (go.work workspace)
COPY . .
# Download deps and build
RUN go work sync && go build -o bin/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache wget
WORKDIR /app
COPY --from=builder /app/bin/server ./server
# Seed default plans — overridable via volume mount at /app/tmp/plans.json
RUN mkdir -p /app/tmp
COPY tmp/plans.json /app/tmp/plans.json
EXPOSE 9090
CMD ["./server"]
