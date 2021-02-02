FROM alpine:3.13
RUN apk add --no-cache ca-certificates
COPY http/assets/css http/assets/css
COPY favicon.ico .
COPY http/html/templates http/html/templates
RUN mkdir db
COPY blog .
EXPOSE 80
ENTRYPOINT [ "./blog" ]