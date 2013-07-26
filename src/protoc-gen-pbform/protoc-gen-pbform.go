package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"code.google.com/p/goprotobuf/proto"
	protobuf "google/protobuf"
	protocompiler "google/protobuf/compiler"

	"pbform"
)

var genHtmlTplPath = flag.String("gen_form_src",
	"src/protoc-gen-pbform",
	"Path to directory containing html templates")
var genHtmlOverrideJs = flag.Bool("gen_form_gen_js", false,
	"Prevents overriding js file. Useful for dev, after linking the work file in place of the generated file.")

func main() {
	flag.Parse()

	// Read the request from stdin.
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Panic(err, "reading input")
	}
	request := new(protocompiler.CodeGeneratorRequest)
	if err := proto.Unmarshal(data, request); err != nil {
		log.Panic(err, "parsing input proto")
	}

	// Process request and generate response.
	response := genForm(request)

	// Write the response to stdout.
	data, err = proto.Marshal(response)
	if err != nil {
		log.Panic(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Panic(err, "failed to write output proto")
	}
}

func genForm(request *protocompiler.CodeGeneratorRequest) *protocompiler.CodeGeneratorResponse {
	response := new(protocompiler.CodeGeneratorResponse)

	response.File = make([]*protocompiler.CodeGeneratorResponse_File, 0, len(request.ProtoFile)+2)

	// Files containing proto bufs are serialized to jsonp.
	for _, desc := range request.ProtoFile {
		file := new(protocompiler.CodeGeneratorResponse_File)
		file.Name = proto.String(
			fmt.Sprintf("%s.js", *desc.Name))
		c, _ := json.Marshal(desc)

		methodPaths := make([]string, 0)
		for _, service := range desc.Service {
			url := ""
			if service.GetOptions() != nil {
				i, err := proto.GetExtension(
					service.GetOptions(), pbform.E_Service)
				if err == nil && i != nil {
					opts := i.(*pbform.ServiceOptions)
					if opts.Url != nil {
						url = *opts.Url
					}
				}
			}

			for _, method := range service.Method {
				if method.GetOptions() != nil {
					i, err := proto.GetExtension(
						method.GetOptions(), pbform.E_Method)
					if err == nil && i != nil {
						opts := i.(*pbform.MethodOptions)
						if opts.Path != nil {
							p := fmt.Sprintf(`setServiceUrl(".%s.%s.%s", "%s%s");`,
								desc.GetPackage(), service.GetName(),
								method.GetName(), url, opts.GetPath())
							methodPaths = append(methodPaths, p)
						}
					}
				}
			}
		}

		content := fmt.Sprintf("setup(%s);\n%s", string(c),
			strings.Join(methodPaths, "\n"))
		file.Content = proto.String(content)
		response.File = append(response.File, file)
	}

	// Index file includes all the above jsonp files.
	index := new(protocompiler.CodeGeneratorResponse_File)
	index.Name = proto.String("index.html")
	index.Content = proto.String(indexPage(request.ProtoFile))
	response.File = append(response.File, index)

	// form.js file has the javascript to interpret all that.
	if *genHtmlOverrideJs {
		js := new(protocompiler.CodeGeneratorResponse_File)
		js.Name = proto.String("pbform.js")
		jsFile, _ := ioutil.ReadFile(
			path.Join(*genHtmlTplPath, "pbform.js"))
		js.Content = proto.String(string(jsFile))
		response.File = append(response.File, js)
	}

	return response
}

func indexPage(files []*protobuf.FileDescriptorProto) string {
	tpl := template.Must(template.New("foo").ParseFiles(
		path.Join(*genHtmlTplPath, "index.html")))

	var out bytes.Buffer
	tpl.ExecuteTemplate(&out, "index.html", struct {
		Files []*protobuf.FileDescriptorProto
	}{
		files,
	})
	return out.String()
}
