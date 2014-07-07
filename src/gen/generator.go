package main

type GenFile struct {
	Name    string
	Content string
}

type Generator interface {
	GenCode(api *JsonSchema) []*GenFile
}
