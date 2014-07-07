package main

import (
	"fmt"
	"sort"
	"strings"
)

type line struct {
	DedupeKey string
	Line      string
}

type lineGenerator interface {
	schema(path string, in *JsonSchema, parent *JsonLink) *line
	link(path string, link *JsonLink, parent *JsonSchema) *line
}

func genAllLines(path string, in *JsonSchema, g lineGenerator) []*line {
	lines := make([]*line, 0)

	for name, schema := range in.Definitions {
		p := path + "/definitions/" + name
		lines = append(lines, g.schema(p, schema, nil))
		for _, m := range genAllLines(p, schema, g) {
			lines = append(lines, m)
		}
	}
	for name, schema := range in.Properties {
		p := path + "/properties/" + name
		lines = append(lines, g.schema(p, schema, nil))
		for _, m := range genAllLines(p, schema, g) {
			lines = append(lines, m)
		}
	}

	for i, link := range in.Links {
		p := fmt.Sprintf("%s/links[%d]", path, i)
		f := g.link(p, link, in)
		lines = append(lines, f)

		if link.Schema != nil {
			lines = append(lines, g.schema(p+"/schema", link.Schema, link))
		}
		if link.TargetSchema != nil {
			lines = append(lines, g.schema(p+"/targetSchema", link.TargetSchema, link))
		}
	}
	return lines
}

func GenLines(in *JsonSchema, g lineGenerator) string {
	filtered := make(map[string]*line)
	for _, l := range genAllLines("#", in, g) {
		if l != nil {
			filtered[l.DedupeKey] = l
		}
	}

	list := make([]string, len(filtered))
	for _, l := range filtered {
		list = append(list, l.Line)
	}

	sort.Strings(list)
	return strings.Join(list, "\n  ")
}
