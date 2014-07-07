package {{.File.Package}}

import (
	"net/http"
)

func init() {
  {{range $s := .File.Service}}
    {{range $s.Method}}
	    http.HandleFunc("{{MethodPath $s .}}", {{.Name}})
    {{end}}
  {{end}}
}

{{range .File.Service}}
type {{.Name}} interface {
  {{range .Method}}
  {{.Name}}({{Type .InputType}}) ({{Type .OutputType}}, error)
  {{end}}
}
{{end}}
