package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"code.google.com/p/goprotobuf/proto"
	protobuf "google/protobuf"
	protocompiler "google/protobuf/compiler"

	"pbform"
)

func genForm(request *protocompiler.CodeGeneratorRequest,
	response *protocompiler.CodeGeneratorResponse) error {
	if opts["gen_html_form"] != "true" {
		return nil
	}

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
	if opts["override_js"] == "true" {
		js := new(protocompiler.CodeGeneratorResponse_File)
		js.Name = proto.String("pbform.js")
		jsFile, _ := ioutil.ReadFile(
			path.Join(opts["tpl_path"], "pbform.js"))
		js.Content = proto.String(string(jsFile))
		response.File = append(response.File, js)
	}

	return nil
}

func indexPage(files []*protobuf.FileDescriptorProto) string {
	tpl := template.Must(template.New("foo").ParseFiles(
		path.Join(opts["tpl_path"], "index.html")))

	var out bytes.Buffer
	tpl.ExecuteTemplate(&out, "index.html", struct {
		Files []*protobuf.FileDescriptorProto
	}{
		files,
	})
	return out.String()
}

func url(service *protobuf.ServiceDescriptorProto,
	method *protobuf.MethodDescriptorProto) string {
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

	path := ""
	if method.GetOptions() != nil {
		i, err := proto.GetExtension(
			method.GetOptions(), pbform.E_Method)
		if err == nil && i != nil {
			opts := i.(*pbform.MethodOptions)
			if opts.Path != nil {
				path = *opts.Path
			}
		}
	}
	return url + path
}
