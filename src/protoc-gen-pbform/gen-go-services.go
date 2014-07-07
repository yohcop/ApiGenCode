package main

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"

	"code.google.com/p/goprotobuf/proto"
	protobuf "google/protobuf"
	protocompiler "google/protobuf/compiler"
)

func genGoServices(
	request *protocompiler.CodeGeneratorRequest,
	response *protocompiler.CodeGeneratorResponse) error {
	if opts["gen_go_services"] != "true" {
		return nil
	}

	funcMap := template.FuncMap{
		"Type": func(tpe string) string {
			s := strings.Split(tpe, ".")
			return s[len(s)-1]
		},
		"MethodPath": func(service *protobuf.ServiceDescriptorProto,
			method *protobuf.MethodDescriptorProto) string {
			return url(service, method)
		},
	}
	tpl := template.Must(template.New("foo").Funcs(funcMap).ParseFiles(
		path.Join(opts["tpl_path"], "interfaces.go.tpl")))

	for _, f := range request.ProtoFile {
		var out bytes.Buffer
		err := tpl.ExecuteTemplate(&out, "interfaces.go.tpl", struct {
			File *protobuf.FileDescriptorProto
		}{
			f,
		})
		if err != nil {
			return err
		}

		interfaces := new(protocompiler.CodeGeneratorResponse_File)
		interfaces.Name = proto.String(
			fmt.Sprintf("%s_interfaces.go", f.GetName()))
		interfaces.Content = proto.String(out.String())
		response.File = append(response.File, interfaces)
	}
	return nil
}
