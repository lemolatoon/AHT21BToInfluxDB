FROM --platform=$BUILDPLATFORM docker.io/golang:1.24-bookworm AS builder

ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} GOOS=linux go build -o sensor-to-db

FROM docker.io/debian:bookworm-slim

WORKDIR /root/
COPY --from=builder /app/sensor-to-db .

ENTRYPOINT ["/bin/sh", "-c", "./sensor-to-db & tail -f /dev/null"]

