# syntax=docker/dockerfile:1
# --- Build stage
FROM golang:1.23 AS builder 
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build just the metadata binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /out/app ./metadata/cmd

# --- Runtime stage (small, has busybox wget for healthchecks)
FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /out/app /app
EXPOSE 8081
ENTRYPOINT ["/app"]