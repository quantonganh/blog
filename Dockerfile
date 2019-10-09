FROM quantonganh/golang:1.13-alpine3.10-arm

RUN mkdir /app

COPY assets /app/assets

COPY favicon.ico /app

COPY posts /app/posts

COPY templates/ /app/templates

COPY blog /app

WORKDIR /app

ENTRYPOINT [ "/app/blog" ]

EXPOSE 80
