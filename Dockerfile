FROM arm32v7/alpine:3.10

RUN mkdir /app

COPY assets /app/assets

COPY favicon.ico /app

COPY posts /app/posts

COPY templates/ /app/templates

COPY blog /app

WORKDIR /app

ENTRYPOINT [ "/app/blog" ]

EXPOSE 80
