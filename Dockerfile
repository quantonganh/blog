FROM scratch

COPY assets assets
COPY favicon.ico .
COPY posts posts
COPY templates templates
COPY blog blog

EXPOSE 80

ENTRYPOINT [ "./blog" ]