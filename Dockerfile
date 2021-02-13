FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY asm-webhook /app
COPY ssl ssl

CMD ["/app/asm-webhook"]