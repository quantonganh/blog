FROM alpine:3.12@sha256:549694ea68340c26d1d85c00039aa11ad835be279bfd475ff4284b705f92c24e as builder
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