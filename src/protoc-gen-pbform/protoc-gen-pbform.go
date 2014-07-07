package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"code.google.com/p/goprotobuf/proto"
	protocompiler "google/protobuf/compiler"
)

// Those options are read from the CodeGeneratorRequest.
// They can be bassed with the _out flag to protoc, e.g.:
//   --pbform_out=tpl_path=path/to/tpls,override_js=true:out/directory
var opts = map[string]string{
	"tpl_path":        "src/protoc-gen-pbform",
	"override_js":     "false",
	"gen_go_services": "false",
	"gen_html_form":   "false",
}

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

	// Get options.
	optStr := strings.Split(request.GetParameter(), ",")
	for _, opt := range optStr {
		vals := strings.SplitN(opt, "=", 2)
		if len(vals) == 2 {
			opts[vals[0]] = vals[1]
		}
	}

	// Process request and generate response.
	response := new(protocompiler.CodeGeneratorResponse)
	response.File = make([]*protocompiler.CodeGeneratorResponse_File, 0, 0)

	err = genForm(request, response)
	if err != nil {
		log.Panic(err, "Generating form")
	}
	err = genGoServices(request, response)
	if err != nil {
		log.Panic(err, "Generating go services")
	}

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
