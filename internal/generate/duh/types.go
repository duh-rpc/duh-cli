package duh

import "io"

type RunConfig struct {
	Writer       io.Writer
	SpecPath     string
	PackageName  string
	OutputDir    string
	ProtoPath    string
	ProtoImport  string
	ProtoPackage string
	FullFlag     bool
	Converter    ProtoConverter
}

type TemplateData struct {
	Package        string
	ModulePath     string
	ProtoImport    string
	ProtoPackage   string
	Operations     []Operation
	ListOps        []ListOperation
	HasListOps     bool
	Timestamp      string
	IsFullTemplate bool
	GoModule       string
}

type Operation struct {
	MethodName   string
	Path         string
	ConstName    string
	Summary      string
	RequestType  string
	ResponseType string
}

type ListOperation struct {
	Operation
	IteratorName  string
	FetcherName   string
	ItemType      string
	ResponseField string
}
