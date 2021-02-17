FROM alpine:3.13
RUN apk add --no-cache ca-certificates
RUN mkdir db
COPY blog .
EXPOSE 80
ENTRYPOINT [ "./blog" ]