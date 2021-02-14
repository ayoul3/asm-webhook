FROM golang:1.15-alpine AS builder

RUN apk add --update --no-cache git make
WORKDIR /build
COPY . /build

RUN make build

FROM alpine

WORKDIR /app

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /build/asm-webhook /app/
COPY ssl ssl

CMD ["/app/asm-webhook"]