module github.com/quantonganh/blog

go 1.13

require (
	github.com/Depado/bfchroma v1.1.2
	github.com/alecthomas/chroma v0.6.7
	github.com/astaxie/beego v1.12.0
	github.com/flosch/pongo2 v0.0.0-20190707114632-bbf5a6c351f4
	github.com/julienschmidt/httprouter v1.3.0
	github.com/pkg/errors v0.8.1
	gopkg.in/russross/blackfriday.v2 v2.0.1
	gopkg.in/yaml.v2 v2.2.4
)

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
