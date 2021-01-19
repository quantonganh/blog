FROM alpine:3.12 as builder
RUN apk add --no-cache ca-certificates

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY http/assets http/assets
COPY favicon.ico .
COPY posts posts
COPY posts.bleve posts.bleve
COPY http/html/templates http/html/templates
COPY blog .
EXPOSE 80
ENTRYPOINT [ "./blog" ]