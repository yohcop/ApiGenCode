package main

import (
	"bytes"
	"flag"
	"log"
	"strings"
	"text/template"
)

var genHtmlFormUrl = flag.String("gen_form_url", "/",
	"Html form path")

var templates = `<html>
<head>
<script>
function get(field) {
  if (document.forms[0][field].value) {
    return document.forms[0][field].value;
  }
}
function getN(field) {
  if (document.forms[0][field].value) {
    return Number(document.forms[0][field].value);
  }
}
function f() {
  var q = {};
  {{.Js}}
  return JSON.stringify(q);
}
function go() {
  window.open('{{.UrlPath}}/{{.Name}}?q=' + f());
}
</script>
</head>
<body>
<a href="index.html">forms</a>
<h1>{{.Name}}</h1>
<form>
{{.Form}}
<button onclick="go();return false()">Send</button>
</form>`

type HtmlFormGenerator struct {
	UrlPath   string
	Templates *template.Template
}

func NewHtmlFormGenerator() *HtmlFormGenerator {
	return &HtmlFormGenerator{
		UrlPath:   *genHtmlFormUrl,
		Templates: template.Must(template.New("foo").Parse(templates)),
	}
}

func (g *HtmlFormGenerator) GenCode(api *JsonApi) []*GenFile {
	files := []*GenFile{
		&GenFile{
			Name:    "index.html",
			Content: g.GenIndex(api),
		},
	}

	for name, method := range api.Methods {
		files = append(files, &GenFile{
			Name:    name + ".html",
			Content: g.GenMethodForm(name, method, api),
		})
	}
	return files
}

func (g *HtmlFormGenerator) GenField(
	name string, schema *JsonSchema, api *JsonApi) (
	title, input, js string) {

	switch schema.Type {
	case "string":
		return name, `<input name="` + name + `"></input>`,
			"q." + name + " = get\"" + name + "\");\n"
	case "number":
		return name, `<input name="` + name + `"></input>`,
			"q." + name + " = getN(\"" + name + "\");\n"
	case "object":
		form := ""
		jss := ""
		for sub, field := range schema.Properties {
			title, input, js := g.GenField(name+"."+sub, field, api)
			form += "<li><label>" + title + "</label>" + input + "</li>"
			jss += js + "\n"
		}
		return name, "<ul>" + form + "</ul>", jss
	}

	if len(schema.Ref) > 0 {
		s := api.Schemas[schema.Ref]
		if s != nil {
			title, input, js := g.GenField(name, s, api)
			return title, input, "q." + name + " = {};\n" + js
		}
		log.Panic("Unknown reference: " + schema.Ref)
	}
	log.Panic("Unknown type: " + schema.Type)
	return
}

func (g *HtmlFormGenerator) GenMethodForm(
	name string, method *JsonMethod, api *JsonApi) string {

	form := ""
	jss := ""
	for name, field := range method.Request.Properties {
		title, input, js := g.GenField(name, field, api)
		form += "<li><label>" + title + "</label>" + input + "</li>\n"
		jss += js
	}

	var out bytes.Buffer
	g.Templates.Execute(&out, struct {
		Js, Name, Form, UrlPath string
	}{
		jss, name, form, g.UrlPath,
	})
	return out.String()
}

func (g *HtmlFormGenerator) GenMethodLink(
	name string, method *JsonMethod) string {

	return "<li><a href=\"" + name + ".html\">" + name + "</a></li>"
}

func (g *HtmlFormGenerator) GenIndex(api *JsonApi) string {
	functions := make([]string, 0, len(api.Methods))
	for name, method := range api.Methods {
		f := g.GenMethodLink(name, method)
		functions = append(functions, f)
	}
	return "<ul>" + strings.Join(functions, "\n  ") + "</ul>"
}
