# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23 AS builder 
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the metadata service
WORKDIR /src/metadata
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/metadata ./cmd

# Run
FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
COPY --from=builder /out/metadata /app
EXPOSE 8081
ENTRYPOINT ["/app"]