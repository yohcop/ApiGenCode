package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"path"
	"strings"
	"text/template"
)

var genHtmlFormUrl = flag.String("gen_form_url", "/",
	"Html form path")
var genHtmlTplPath = flag.String("gen_form_src", "src/gen",
	"Path to directory containing html templates")
var genHtmlOverrideJs = flag.Bool("gen_form_gen_js", true,
	"Precents overriding js file. Useful for dev.")

type HtmlFormGenerator struct {
	UrlPath   string
	Templates *template.Template
}

func NewHtmlFormGenerator() *HtmlFormGenerator {
	return &HtmlFormGenerator{
		UrlPath: *genHtmlFormUrl,
		Templates: template.Must(template.New("foo").ParseFiles(
			path.Join(*genHtmlTplPath, "form.html"))),
	}
}

func (g *HtmlFormGenerator) GenCode(api *JsonSchema) []*GenFile {
	files := []*GenFile{
		&GenFile{
			Name:    "index.html",
			Content: g.GenIndex(api),
		},
	}

	if *genHtmlOverrideJs {
		js, _ := ioutil.ReadFile(path.Join(*genHtmlTplPath, "form.js"))
		files = append(files, &GenFile{
			Name:    "form.js",
			Content: string(js),
		})
	}

	for name, def := range api.Definitions {
		for _, link := range def.Links {
			files = append(files, &GenFile{
				Name:    name + ".html",
				Content: g.GenLinkForm(name, link, api),
			})
		}

	}
	return files
}

func (g *HtmlFormGenerator) GenLinkForm(
	name string, method *JsonLink, api *JsonSchema) string {

	rs, _ := json.Marshal(method.Schema)
	schemas, _ := json.Marshal(api.Definitions)

	var out bytes.Buffer
	g.Templates.ExecuteTemplate(&out, "form.html", struct {
		Name, UrlPath, MethodPath, Req, Schemas string
	}{
		name, g.UrlPath, method.Href, string(rs), string(schemas),
	})
	return out.String()
}

func (g *HtmlFormGenerator) GenMethodLink(
	name string, method *JsonLink) string {

	return "<li><a href=\"" + name + ".html\">" + name + "</a></li>"
}

func (g *HtmlFormGenerator) GenIndex(api *JsonSchema) string {
	functions := make([]string, 0)
	for name, def := range api.Definitions {
		for _, link := range def.Links {
			f := g.GenMethodLink(name, link)
			functions = append(functions, f)
		}
	}
	return "<ul>" + strings.Join(functions, "\n  ") + "</ul>"
}
