FROM golang:1.21-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /go/src/fowergram

# Install build tools
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /go/bin/app ./cmd/api

FROM alpine:latest
COPY --from=builder /go/bin/app /app
CMD ["/app"]