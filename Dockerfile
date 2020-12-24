FROM alpine:3.12@sha256:549694ea68340c26d1d85c00039aa11ad835be279bfd475ff4284b705f92c24e

RUN mkdir /app

COPY assets /app/assets

COPY favicon.ico /app

COPY posts /app/posts

COPY templates/ /app/templates

COPY blog /app

WORKDIR /app

ENTRYPOINT [ "/app/blog" ]

EXPOSE 80
