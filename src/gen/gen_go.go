package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
)

var genGoPkg = flag.String("gen_go_pkg", "main", "Go package")
var genGoFmt = flag.Bool("gen_go_fmt", true, "Run gofmt on output")
var genGoDbg = flag.Bool("gen_go_dbg", false, "Add debug to output code")

type GoGenerator struct {
	Package string
}

func NewGoGenerator() *GoGenerator {
	return &GoGenerator{
		Package: *genGoPkg,
	}
}

func methodOrGet(link *JsonLink) string {
	if link.Method != "" {
		return link.Method
	}
	return "GET"
}

func genLinkDoc(link *JsonLink, key string) string {
	doc := ""
	if link.Title != "" || link.Description != "" {
		doc += "\n\n"
	}
	if link.Title != "" {
		doc += "// " + link.Title + "\n"
	}
	if link.Description != "" {
		doc += "// " + link.Description + "\n"
	}
	if *genGoDbg {
		doc += "// (ApiGenCode: key=" + key + ")\n"
	}
	return doc
}

func genSchemaDoc(schema *JsonSchema, key string) string {
	doc := ""
	if schema.Title != "" || schema.Description != "" {
		doc += "\n\n"
	}
	if schema.Title != "" {
		doc += "// " + schema.Title + "\n"
	}
	if schema.Description != "" {
		doc += "// " + schema.Description + "\n"
	}
	if *genGoDbg {
		doc += "// (ApiGenCode: key=" + key + ")\n"
	}
	return doc
}

func (g *GoGenerator) GenCode(api *JsonSchema) []*GenFile {
	return []*GenFile{
		&GenFile{
			Name:    "types.go",
			Content: g.MaybeRunGoFmt(g.WrapFile(g.Objects(api))),
		},
		&GenFile{
			Name:    "paths.go",
			Content: g.MaybeRunGoFmt(g.WrapFile(g.Paths(api))),
		},
		&GenFile{
			Name:    "interface.go",
			Content: g.MaybeRunGoFmt(g.WrapFile(g.Interface(api))),
		},
		&GenFile{
			Name:    "handler.go",
			Content: g.MaybeRunGoFmt(g.WrapFile(g.Handler(api))),
		},
	}
}

func (g *GoGenerator) WrapFile(content string) string {
	return fmt.Sprintf("package %s\n\n%s", g.Package, content)
}

func (g *GoGenerator) EnumType(schema *JsonSchema) string {
	var common reflect.Kind = reflect.Invalid
	for _, obj := range schema.Enum {
		t := reflect.TypeOf(obj).Kind()
		if common == reflect.Invalid {
			common = t
			continue
		}
		if t != common {
			return "interface{}"
		}
	}
	switch common {
	case reflect.Bool:
		return "bool"
	case reflect.Float64:
		return "float64"
	case reflect.String:
		return "string"
	}
	return "interface{}"
}

func (g *GoGenerator) TypeName(path string, ptr bool, schema *JsonSchema) string {
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
		if ptr {
			return "*" + g.GoName(path)
		}
		return g.GoName(path)
	case "array":
		return g.GoName(path)
	}
	if schema.Ref != "" {
		if ptr {
			return "*" + g.GoName(schema.Ref)
		}
		return g.GoName(schema.Ref)
	}
	if len(schema.Enum) != 0 {
		return g.EnumType(schema)
	}
	return "/* SHOULD NOT COMPILE */"
}

// =======================================================

type structGenerator struct {
	g *GoGenerator
}

func (i *structGenerator) schema(path string, in *JsonSchema, parent *JsonLink) *line {
	name := path
	// Note: we only do this for the input. it is unlikely that a
	// custom output would be definede here (i.e., not with $ref)
	// but it is still possible. May be worth doing the same with
	// targetSchema.
	if parent != nil && strings.HasSuffix(path, "/schema") {
		if parent.Schema.Ref != "" {
			name = i.g.TypeName(path, true, parent.Schema)
		} else {
			name = i.g.GoName(parent.Title) + "Input"
		}
	}

	doc := genSchemaDoc(in, path)

	if in.Type == "object" {
		content := make([]string, 0, len(in.Properties))
		for fieldName, field := range in.Properties {
			t := i.g.TypeName(path+"/"+fieldName, true, field)
			content = append(content,
				fmt.Sprintf("%s %s `json:\"%s,omitempty\"`",
					i.g.GoName(fieldName), t, fieldName))
		}
		l := fmt.Sprintf("%stype %s struct {\n  %s\n}",
			doc, i.g.GoName(name), strings.Join(content, "\n  "))
		return &line{path, l}
	} else if in.Type == "array" {
		l := fmt.Sprintf("%stype %s []%s\n", doc,
			i.g.GoName(name), i.g.TypeName(path, true, in.Items))
		return &line{path, l}
	} else if len(in.Enum) > 0 {
		l := i.g.GoName(name)
		enumType := i.g.EnumType(in)
		values := make([]string, 0, len(in.Enum))
		for n, v := range in.Enum {
			key := i.enumVar(n, v)
			val := i.formatEnumValue(enumType, v)
			values = append(values,
				fmt.Sprintf("%s_%s %s = %s", l, key, l, val))
		}
		return &line{path, fmt.Sprintf(`
        %stype %s %s
        const (
          %s
        )`, doc, l, enumType, strings.Join(values, "\n"))}
	}
	return nil
}

func (i *structGenerator) link(path string, link *JsonLink, parent *JsonSchema) *line {
	return nil
}

func (i *structGenerator) enumVar(n int, val interface{}) string {
	switch v := val.(type) {
	case nil:
		return "_nil_"
	case int:
		return fmt.Sprintf("%d", v)
	case string:
		return i.g.GoName(v)
	case bool:
		return fmt.Sprintf("%t", v)
	}
	return fmt.Sprintf("_%d_", n)
}

func (i *structGenerator) formatEnumValue(tpe string, val interface{}) string {
	switch {
	case tpe == "string":
		return fmt.Sprintf("\"%s\"", val)
	case tpe == "bool":
		return fmt.Sprintf("%t", val)
	case tpe == "float64":
		return fmt.Sprintf("%f", val)
	}
	data, _ := json.Marshal(val)
	return fmt.Sprintf("\"%s\"", data)
}

func (g *GoGenerator) Objects(schema *JsonSchema) string {
	gen := &structGenerator{g}
	return GenLines(schema, gen)
}

// =======================================================

type interfaceGenerator struct {
	g *GoGenerator
}

func (i *interfaceGenerator) schema(path string, in *JsonSchema, parent *JsonLink) *line {
	return nil
}

func (i *interfaceGenerator) link(path string, link *JsonLink, parent *JsonSchema) *line {
	name := i.g.GoName(link.Title)
	var req, resp string
	if link.Schema != nil {
		if link.Schema.Ref != "" {
			req = i.g.TypeName(path+"/schema", true, link.Schema)
		} else {
			req = "*" + name + "Input"
		}
	}
	if link.TargetSchema != nil {
		if link.TargetSchema.Ref != "" {
			resp = i.g.TypeName(path+"/targetSchema", true, link.TargetSchema)
		} else {
			resp = name + "Output"
		}
	}
	params := make([]string, 0)
	for _, extraParam := range i.Placeholders(link, parent) {
		params = append(params, extraParam[0]+" "+extraParam[1])
	}
	if len(req) > 0 {
		params = append(params, "input "+req)
	}
	re := regexp.MustCompile("\\{([^}]+)\\}")
	key := methodOrGet(link) + " " + re.ReplaceAllString(link.Href, "{}")
	return &line{
		DedupeKey: key,
		Line: fmt.Sprintf("%s%s(%s) (%s, error)",
			genLinkDoc(link, key), name, strings.Join(params, ","), resp),
	}
}

func (i *interfaceGenerator) Placeholders(method *JsonLink, parent *JsonSchema) [][]string {
	params := make([][]string, 0)
	re := regexp.MustCompile("\\{([^}]+)\\}")
	placeholders := re.FindAllStringSubmatch(method.Href, -1)
	for _, match := range placeholders {
		name := match[1]
		if schema, known := parent.Properties[name]; known {
			params = append(params, []string{name, i.g.TypeName("", true, schema)})
		} else {
			fmt.Println("Unknwon " + name)
		}
	}
	return params
}

func (g *GoGenerator) Interface(api *JsonSchema) string {
	gen := &interfaceGenerator{g}

	return "type " + g.GoName(api.Title) + " interface {\n" +
		"  " + GenLines(api, gen) +
		"\n}"
}

// =======================================================

type handlerGenerator struct {
	g     *GoGenerator
	class string
}

func (i *handlerGenerator) schema(path string, in *JsonSchema, parent *JsonLink) *line {
	return nil
}

func (i *handlerGenerator) link(path string, link *JsonLink, parent *JsonSchema) *line {
	name := i.g.GoName(link.Title)
	var req string
	if link.Schema != nil {
		if link.Schema.Ref != "" {
			req = i.g.TypeName(path+"/schema", false, link.Schema)
		} else {
			req = name + "Input"
		}
	}
	params := make([]string, 0)
	for _, extraParam := range i.Placeholders(link, parent) {
		params = append(params, "_"+extraParam[0])
	}
	re := regexp.MustCompile("\\{([^}]+)\\}")
	key := methodOrGet(link) + " " + re.ReplaceAllString(link.Href, "{}")

	var args []string
	for i, p := range params {
		args = append(args, fmt.Sprintf(`%s := matches[0][%d]`, p, i + 1))
	}
	var matches = ""
	if len(args) > 0 {
		matches = "matches := re.FindAllStringSubmatch(r.URL.Path, -1)"
	}

	if len(req) > 0 {
		params = append(params, "input")
		args = append(args, fmt.Sprintf(`input := &%s{}
          body, _ := ioutil.ReadAll(r.Body)
          json.Unmarshal(body, input)`,
			req))
	}

	return &line{
		DedupeKey: key,
		Line: fmt.Sprintf(`%sfunc (h *%s) _%s(w http.ResponseWriter, r *http.Request) (bool, error) {
        re := regexp.MustCompile("^%s$")
        if r.Method == "%s" && re.MatchString(r.URL.Path) {
          %s
          %s
          r, err := h.Api.%s(%s)
          if err != nil {
            return true, err
          }
          if resp, err := json.Marshal(r); err == nil {
            w.Write(resp)
            return true, nil
          } else {
            return true, err
          }
        }
        return false, nil
      }`,
			genLinkDoc(link, key), i.class, name,
			re.ReplaceAllString(link.Href, "([^/]+)"),
			methodOrGet(link),
			matches,
			strings.Join(args, "\n"),
			name, strings.Join(params, ", ")),
	}
}

func (i *handlerGenerator) Placeholders(method *JsonLink, parent *JsonSchema) [][]string {
	params := make([][]string, 0)
	re := regexp.MustCompile("\\{([^}]+)\\}")
	placeholders := re.FindAllStringSubmatch(method.Href, -1)
	for _, match := range placeholders {
		name := match[1]
		if schema, known := parent.Properties[name]; known {
			params = append(params, []string{name, i.g.TypeName("", true, schema)})
		} else {
			fmt.Println("Unknwon " + name)
		}
	}
	return params
}

type dispatcherGenerator struct {
	g *GoGenerator
}

func (i *dispatcherGenerator) schema(path string, in *JsonSchema, parent *JsonLink) *line {
	return nil
}

func (i *dispatcherGenerator) link(path string, link *JsonLink, parent *JsonSchema) *line {
	name := i.g.GoName(link.Title)
	re := regexp.MustCompile("\\{([^}]+)\\}")
	key := methodOrGet(link) + " " + re.ReplaceAllString(link.Href, "{}")
	return &line{
		DedupeKey: key,
		Line: fmt.Sprintf(`if ok, err := s._%s(w, r); ok {
        return true, err
      }`, name),
	}
}

func (g *GoGenerator) Handler(api *JsonSchema) string {
	gen := &handlerGenerator{g, g.GoName(api.Title) + "Handler"}
	dispatch := &dispatcherGenerator{g}

	return fmt.Sprintf(`import (
      "net/http"
      "regexp"
      "encoding/json"
      "io/ioutil"
    )

    type %s struct {
      Api %s
    }

    func (s *%sHandler) Dispatch(w http.ResponseWriter, r *http.Request) (bool, error) {
      %s
      return false, nil
    }

    %s
    `,
		gen.class,
		g.GoName(api.Title),
		g.GoName(api.Title),
		GenLines(api, dispatch),
		GenLines(api, gen))
}

// =======================================================

type pathsGenerator struct {
	g *GoGenerator
}

func (i *pathsGenerator) schema(path string, in *JsonSchema, parent *JsonLink) *line {
	return nil
}

func (i *pathsGenerator) link(path string, link *JsonLink, parent *JsonSchema) *line {
	method := methodOrGet(link)
	f := fmt.Sprintf(`{
      "%s", "%s", "%s",
    },`, i.g.GoName(link.Title), method, link.Href)
	re := regexp.MustCompile("\\{([^}]+)\\}")
	key := method + "." + re.ReplaceAllString(link.Href, "{}")
	return &line{key, f}
}

func (g *GoGenerator) Paths(api *JsonSchema) string {
	gen := &pathsGenerator{g}

	strukt := `type pathDef struct {
    Name string
    Method string
    Href string
  }
  `

	return strukt + "var " + g.GoName(api.Title) + "Paths = []pathDef{" + GenLines(api, gen) + "\n}\n"
}

// =======================================================

func camelcase(in string, splits []string) string {
	for _, split := range splits {
		if strings.Contains(in, split) {
			path := strings.Split(in, split)
			for i, p := range path {
				path[i] = strings.ToUpper(p[0:1]) + p[1:]
			}
			in = strings.Join(path, "")
		}
	}
	return in
}

func (g *GoGenerator) GoName(jsonName string) string {
	if strings.Contains(jsonName, "/") {
		path := strings.Split(jsonName, "/")
		jsonName = path[len(path)-1]
	}
	jsonName = camelcase(jsonName, []string{" ", "_"})
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
		log.Println("=== gofmt returned an error. input was: ===")
		for i, l := range strings.Split(in, "\n") {
			fmt.Printf("%3.d  %s\n", i+1, l)
		}
		log.Println("=== gofmt output: ===")
		log.Println(errOut.String())
		log.Println(err)
		log.Println("=== end gofmt error ===")
	}
	return out.String()
}
