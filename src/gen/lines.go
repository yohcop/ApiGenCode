package main

import (
	"sort"
	"strings"
)

type line struct {
	DedupeKey string
	Line      string
}

type lineGenerator interface {
	schema(prefixes []string, in *JsonSchema, parent *JsonLink) *line
	link(prefixes []string, link *JsonLink, parent *JsonSchema) *line
}

func genAllLines(prefixes []string, in *JsonSchema, g lineGenerator) []*line {
	lines := make([]*line, 0)

	for name, schema := range in.Definitions {
		lines = append(lines, g.schema(append(prefixes, name), schema, nil))
		for _, m := range genAllLines(append(prefixes, name), schema, g) {
			lines = append(lines, m)
		}
	}
	for name, schema := range in.Properties {
		lines = append(lines, g.schema(append(prefixes, name), schema, nil))
		for _, m := range genAllLines(append(prefixes, name), schema, g) {
			lines = append(lines, m)
		}
	}

	for _, link := range in.Links {
		f := g.link(prefixes, link, in)
		lines = append(lines, f)

		if link.Schema != nil {
			lines = append(lines, g.schema(append(prefixes, link.Title), link.Schema, link))
		}
		if link.TargetSchema != nil {
			lines = append(lines, g.schema(append(prefixes, link.Title), link.TargetSchema, link))
		}
	}
	return lines
}

func GenLines(in *JsonSchema, g lineGenerator) string {
	filtered := make(map[string]*line)
	for _, l := range genAllLines([]string{}, in, g) {
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
