FROM quantonganh/golang:1.13-alpine3.10-arm

COPY blog /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/blog" ]
