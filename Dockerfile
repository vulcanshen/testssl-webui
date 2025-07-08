FROM ghcr.io/testssl/testssl.sh:3.2

COPY testssl-webui /usr/local/bin/testssl-webui
COPY public-back /app/public

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/testssl-webui"]
