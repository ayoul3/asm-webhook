FROM alpine
WORKDIR /app
COPY ssm-webhook /app
COPY ssl ssl
COPY ssl/ca-certificates.crt /etc/ssl/certs/
CMD ["/app/ssm-webhook"]