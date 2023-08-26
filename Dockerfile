FROM alpine:3.13
WORKDIR /app
RUN apk add --no-cache ca-certificates
ADD posts.tar.xz .
RUN mkdir db
COPY blog .
EXPOSE 80
ENTRYPOINT [ "./blog" ]