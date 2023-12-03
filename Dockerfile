FROM alpine:3.13
WORKDIR /app
RUN apk add --no-cache ca-certificates git
RUN mkdir db
COPY blog .
EXPOSE 8009
ENTRYPOINT [ "./blog" ]