FROM alpine
WORKDIR /app

COPY asm-webhook /app
COPY ssl ssl
COPY ssl/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/asm-webhook"]