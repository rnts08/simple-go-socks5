FROM golang:1.25-alpine AS builder

ARG GOARCH=amd64

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GOARCH=$GOARCH go build -o /go/bin/go-socks5 .

FROM scratch
COPY --from=builder /go/bin/go-socks5 /go/bin/go-socks5