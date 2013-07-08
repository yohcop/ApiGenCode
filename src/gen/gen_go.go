package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

var genGoPkg = flag.String("gen_go_pkg", "main", "Go package")

type GoGenerator struct {
	Package string
}

func NewGoGenerator() *GoGenerator {
	return &GoGenerator{
		Package: *genGoPkg,
	}
}

func (g *GoGenerator) GenCode(api *JsonApi) []*GenFile {
	typesContent := make([]string, 0, len(api.Schemas))
	for name, schema := range api.Schemas {
		typesContent = append(typesContent, g.Object(name, schema))
	}

	return []*GenFile{
		&GenFile{
			Name:    "types.go",
			Content: g.WrapFile(strings.Join(typesContent, "\n")),
		}, &GenFile{
			Name:    "interface.go",
			Content: g.WrapFile(g.Interface(api)),
		},
	}
}

func (g *GoGenerator) WrapFile(content string) string {
	return fmt.Sprintf("package %s\n\n%s", g.Package, content)
}

func (g *GoGenerator) TypeName(schema *JsonSchema) string {
	switch schema.Type {
	case "number":
		return "float32"
	case "string":
		return "string"
	case "integer":
		return "int"
	case "boolean":
		return "bool"
	case "object":
		return "*" + schema.Ref
	case "array":
		return "[]" + g.TypeName(schema.Items)
	}
	return "*" + schema.Ref
}

func (g *GoGenerator) Object(name string, schema *JsonSchema) string {
	if schema.Type == "object" {
		content := make([]string, 0, len(schema.Properties))
		for fieldName, field := range schema.Properties {
			t := g.TypeName(field)
			content = append(content,
				fmt.Sprintf("%s %s `json:\"%s,omitempty\"`",
					g.GoName(fieldName), t, fieldName))
		}
		return fmt.Sprintf("type %s struct {\n"+
			"  "+strings.Join(content, "\n  ")+
			"\n}", name)
	}
	log.Panicf("Can't generate an object of type %s", schema.Type)
	return ""
}

func (g *GoGenerator) Method(name string, method *JsonMethod) (req, resp, stub string) {
	req = g.Object(name+"Req", method.Request)
	resp = g.Object(name+"Resp", method.Response)
	stub = fmt.Sprintf("%s(*%sReq) (*%sResp, err)",
		name, name, name)
	return
}

func (g *GoGenerator) Interface(api *JsonApi) string {
	types := make([]string, 0, len(api.Methods))
	functions := make([]string, 0, len(api.Methods))
	for name, method := range api.Methods {
		req, resp, f := g.Method(g.GoName(name), method)
		types = append(types, req+"\n"+resp)
		functions = append(functions, f)
	}
	return strings.Join(types, "\n") +
		"\ntype " + g.GoName(api.Name) + " interface {\n" +
		"  " + strings.Join(functions, "\n  ") +
		"\n}"
}

func (g *GoGenerator) GoName(jsonName string) string {
	return strings.ToUpper(jsonName[0:1]) + jsonName[1:]
}
