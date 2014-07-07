package main

// Structures used to represent the schema.
type JsonSchema struct {
	Title       string                 `json:"title,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Definitions map[string]*JsonSchema `json:"definitions,omitempty"`

	// If type is 'object'
	Properties map[string]*JsonSchema `json:"properties,omitempty"`

	// If type is 'array'
	Items *JsonSchema `json:"items,omitempty"`

	// If type is not present, implies object of type Ref, or enum values.
	Enum []interface{} `json:"enum,omitempty"`
	Ref  string        `json:"$ref,omitempty"`

	Links []*JsonLink `json:"links,omitempty"`
}

type JsonLink struct {
	Href         string      `json:"href,omitempty"`
	Rel          string      `json:"rel,omitempty"`
	Title        string      `json:"title,omitempty"`
	Method       string      `json:"method,omitempty"`
	EncType      string      `json:"encType,omitempty"`
	Schema       *JsonSchema `json:"schema,omitempty"`
	TargetSchema *JsonSchema `json:"targetSchema,omitempty"`
}
