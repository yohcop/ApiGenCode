package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var genGoPkg = flag.String("gen_go_pkg", "main", "Go package")
var genGoFmt = flag.Bool("gen_go_fmt", true, "Run gofmt on output")

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
			Name: "types.go",
			Content: g.MaybeRunGoFmt(
				g.WrapFile(strings.Join(typesContent, "\n"))),
		},
		&GenFile{
			Name:    "paths.go",
			Content: g.MaybeRunGoFmt(g.WrapFile(g.Paths(api))),
		},
		&GenFile{
			Name:    "interface.go",
			Content: g.MaybeRunGoFmt(g.WrapFile(g.Interface(api))),
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
	if schema.Ref != "" {
		return "*" + schema.Ref
	} else if len(schema.Enum) != 0 {
		return "interface{}"
	}
	return "/* SHOULD NOT COMPILE */"
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
	stub = fmt.Sprintf("%s(*%sReq) (*%sResp, error)",
		name, name, name)
	return
}

func (g *GoGenerator) Paths(api *JsonApi) string {
	paths := make([]string, 0, len(api.Methods))
	for name, method := range api.Methods {
		p := `"` + g.GoName(name) + `": "` + method.Path + `",`
		paths = append(paths, p)
	}
	return "var " + g.GoName(api.Name) + "Paths = map[string]string{\n  " +
		strings.Join(paths, "\n  ") + "\n}"
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

func (g *GoGenerator) MaybeRunGoFmt(in string) string {
	if *genGoFmt {
		return g.RunGoFmt(in)
	}
	return in
}

func (g *GoGenerator) RunGoFmt(in string) string {
	cmd := exec.Command("gofmt")
	cmd.Stdin = strings.NewReader(in)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if err != nil {
		log.Println(errOut.String())
		log.Println(err)
	}
	return out.String()
}
