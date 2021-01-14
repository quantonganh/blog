FROM scratch

COPY http/assets assets
COPY favicon.ico .
COPY posts posts
COPY posts.bleve posts.bleve
COPY templates templates
COPY blog .

EXPOSE 80

ENTRYPOINT [ "./blog" ]