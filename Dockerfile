FROM alpine:3.13
RUN apk add --no-cache ca-certificates
COPY http/assets http/assets
COPY favicon.ico .
COPY posts posts
COPY posts.bleve posts.bleve
COPY http/html/templates http/html/templates
RUN mkdir db
COPY blog .
EXPOSE 80
ENTRYPOINT [ "./blog" ]