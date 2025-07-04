FROM ghcr.io/testssl/testssl.sh:3.2

# 設定工作目錄
WORKDIR /app

COPY testssl-webui /usr/local/bin/testssl-webui

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/testssl-webui"]
