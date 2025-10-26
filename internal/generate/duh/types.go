package duh

type TemplateData struct {
	Package      string
	ModulePath   string
	ProtoImport  string
	ProtoPackage string
	Operations   []Operation
	ListOps      []ListOperation
	HasListOps   bool
	Timestamp    string
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
