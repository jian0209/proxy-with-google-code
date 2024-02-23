FROM alpine:3.14

WORKDIR /usr/local

COPY ./build/proxy_server_386 .
COPY config.json.example ./config.json

ENTRYPOINT ["./proxy_server_386 -s"]