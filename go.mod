module github.com/quantonganh/blog

go 1.13

require (
	github.com/Depado/bfchroma v1.1.2
	github.com/alecthomas/chroma v0.6.7
	github.com/astaxie/beego v1.12.0
	github.com/blevesearch/bleve v1.0.14
	github.com/bmatcuk/doublestar/v2 v2.0.4
	github.com/flosch/pongo2 v0.0.0-20190707114632-bbf5a6c351f4
	github.com/gorilla/handlers v1.4.1
	github.com/gorilla/mux v1.7.3
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/russross/blackfriday.v2 v2.0.1
	gopkg.in/yaml.v2 v2.2.4
)

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
