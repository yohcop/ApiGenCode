package main

// Structures used to represent the schema.

type JsonApi struct {
	Title   string                 `json:"title,omitempty"`
	Name    string                 `json:"name,omitempty"`
	Schemas map[string]*JsonSchema `json:"schemas,omitempty"`
	Methods map[string]*JsonMethod `json:"methods,omitempty"`
}

type JsonSchema struct {
	Type string `json:"type,omitempty"`
	// If type is 'object'
	Properties map[string]*JsonSchema `json:"properties,omitempty"`
	// If type is 'array'
	Items *JsonSchema `json:"items,omitempty"`
	// If type is not present, implies object of type Ref.
	Ref string `json:"$ref,omitempty"`
}

type JsonMethod struct {
	Description string      `json:"description,omitempty"`
	Request     *JsonSchema `json:"request,omitempty"`
	Response    *JsonSchema `json:"response,omitempty"`
}
