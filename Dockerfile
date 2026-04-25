FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/go-socks5 .

FROM scratch
COPY --from=builder /go/bin/go-socks5 /go/bin/go-socks5
CMD ["/go/bin/go-socks5"]