package duh

import (
	"bytes"
	"go/format"
	"text/template"
	"time"
)

type Generator struct {
	templates *template.Template
	timestamp string
}

func NewGenerator() (*Generator, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	return &Generator{
		templates: tmpl,
		timestamp: generateTimestamp(),
	}, nil
}

func (g *Generator) RenderServer(data *TemplateData) ([]byte, error) {
	data.Timestamp = g.timestamp

	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "server.go.tmpl", data); err != nil {
		return nil, err
	}

	return g.FormatCode(buf.Bytes())
}

func (g *Generator) RenderIterator(data *TemplateData) ([]byte, error) {
	data.Timestamp = g.timestamp

	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "iterator.go.tmpl", data); err != nil {
		return nil, err
	}

	return g.FormatCode(buf.Bytes())
}

func (g *Generator) RenderClient(data *TemplateData) ([]byte, error) {
	data.Timestamp = g.timestamp

	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "client.go.tmpl", data); err != nil {
		return nil, err
	}

	return g.FormatCode(buf.Bytes())
}

func (g *Generator) FormatCode(code []byte) ([]byte, error) {
	return format.Source(code)
}

func generateTimestamp() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
}
