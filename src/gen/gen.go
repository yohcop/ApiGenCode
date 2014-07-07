package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var jsonFile = flag.String("schema", "", "Path to json schema")
var lang = flag.String("lang", "go",
    `Language to generate code for. ["go", "html"]`)
var outDir = flag.String("out", "",
    "Output directory. If emtpy, prints to stdout")
var showParsed = flag.Bool("show_parsed", false,
	"Prints what was parsed from the schema file.")

func main() {
	flag.Parse()
	api, err := ioutil.ReadFile(*jsonFile)
	if err != nil {
		log.Panic(err)
	}
	jsonApi := new(JsonSchema)
	err = json.Unmarshal(api, jsonApi)
	if err != nil {
		log.Panic(err)
	}

	if *showParsed {
		out, _ := json.MarshalIndent(jsonApi, "", "  ")
		log.Printf("%s", out)
	}

	var gen Generator
	switch *lang {
	case "go":
		gen = NewGoGenerator()
	case "html":
		gen = NewHtmlFormGenerator()
	default:
		log.Panicf("Unknown language: %s", *lang)
	}

	if *outDir == "" {
		for _, f := range gen.GenCode(jsonApi) {
			fmt.Println("==== " + f.Name + " ====")
			fmt.Println(f.Content)
		}
	} else {
		for _, f := range gen.GenCode(jsonApi) {
			if err := os.MkdirAll(*outDir, 0755); err != nil {
				log.Panic(err)
			}
			err := ioutil.WriteFile(
				path.Join(*outDir, f.Name), []byte(f.Content), 0644)
			if err != nil {
				log.Panic(err)
			}
		}
	}
}
