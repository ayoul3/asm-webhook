FROM scratch
WORKDIR /app
COPY ssm-webhook .
COPY ssl ssl
COPY /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/app/ssm-webhook"]