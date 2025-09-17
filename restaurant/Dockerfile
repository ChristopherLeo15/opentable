# syntax=docker/dockerfile:1

# Build
FROM golang:1.23 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/restaurant ./restaurant/cmd

# Run
FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
COPY --from=builder /out/restaurant /app
EXPOSE 8082
ENTRYPOINT ["/app"]