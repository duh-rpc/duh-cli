package duh

import (
	"fmt"
	"strings"
	"time"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

type Parser struct {
	spec           *v3.Document
	config         *Config
	isFullTemplate bool
}

func NewParser(spec *v3.Document, config *Config, isFullTemplate bool) *Parser {
	return &Parser{
		spec:           spec,
		config:         config,
		isFullTemplate: isFullTemplate,
	}
}

func (p *Parser) Parse() (*TemplateData, error) {
	modulePath, err := p.config.DetectModulePath()
	if err != nil {
		return nil, err
	}

	operations, err := p.extractOperations()
	if err != nil {
		return nil, err
	}

	listOps, err := p.detectListOperations(operations)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")

	return &TemplateData{
		PackageImport:  p.config.ConstructPackageImport(modulePath),
		Package:        p.config.PackageName,
		ModulePath:     modulePath,
		ProtoImport:    p.config.ConstructProtoImport(modulePath),
		ProtoPackage:   p.config.DeriveProtoPackage(),
		Operations:     operations,
		ListOps:        listOps,
		HasListOps:     len(listOps) > 0,
		Timestamp:      timestamp,
		IsFullTemplate: p.isFullTemplate,
		GoModule:       modulePath,
	}, nil
}

func (p *Parser) extractOperations() ([]Operation, error) {
	var operations []Operation

	if p.spec.Paths == nil || p.spec.Paths.PathItems == nil {
		return operations, nil
	}

	for pair := orderedmap.First(p.spec.Paths.PathItems); pair != nil; pair = pair.Next() {
		path := pair.Key()
		pathItem := pair.Value()

		if pathItem.Post == nil {
			continue
		}

		operation := pathItem.Post
		operationName, err := GenerateOperationName(path)
		if err != nil {
			continue
		}

		requestType := ""
		if operation.RequestBody != nil && operation.RequestBody.Content != nil {
			for contentPair := orderedmap.First(operation.RequestBody.Content); contentPair != nil; contentPair = contentPair.Next() {
				mediaType := contentPair.Value()
				if mediaType.Schema != nil {
					if mediaType.Schema.IsReference() {
						ref := mediaType.Schema.GetReference()
						requestType = "pb." + extractSchemaName(ref)
						break
					} else {
						return nil, fmt.Errorf("inline schema not supported for request body in path %s", path)
					}
				}
			}
		}

		if requestType == "" {
			continue
		}

		responseType := ""
		if operation.Responses != nil && operation.Responses.Codes != nil {
			for responsePair := orderedmap.First(operation.Responses.Codes); responsePair != nil; responsePair = responsePair.Next() {
				response := responsePair.Value()
				if response.Content != nil {
					for contentPair := orderedmap.First(response.Content); contentPair != nil; contentPair = contentPair.Next() {
						mediaType := contentPair.Value()
						if mediaType.Schema != nil {
							if mediaType.Schema.IsReference() {
								ref := mediaType.Schema.GetReference()
								responseType = "pb." + extractSchemaName(ref)
								break
							} else {
								return nil, fmt.Errorf("inline schema not supported for response body in path %s", path)
							}
						}
					}
				}
				if responseType != "" {
					break
				}
			}
		}

		if responseType == "" {
			continue
		}

		summary := ""
		if operation.Summary != "" {
			summary = operation.Summary
		} else if operation.Description != "" {
			summary = operation.Description
		}

		operations = append(operations, Operation{
			MethodName:   operationName,
			Path:         path,
			ConstName:    GenerateConstName(operationName),
			Summary:      summary,
			RequestType:  requestType,
			ResponseType: responseType,
		})
	}

	return operations, nil
}

func (p *Parser) detectListOperations(ops []Operation) ([]ListOperation, error) {
	var listOps []ListOperation

	for _, op := range ops {
		requestSchema, responseSchema, err := p.getSchemas(op.Path)
		if err != nil {
			continue
		}

		if p.isListOperation(op.Path, requestSchema, responseSchema) {
			fieldName, itemType, found := p.findFirstArrayField(responseSchema)
			if !found {
				continue
			}

			listOps = append(listOps, ListOperation{
				Operation:     op,
				IteratorName:  op.MethodName + "Iter",
				FetcherName:   strings.TrimSuffix(itemType, "Response") + "PageFetcher",
				ItemType:      "*pb." + itemType,
				ResponseField: fieldName,
			})
		}
	}

	return listOps, nil
}

func (p *Parser) isListOperation(path string, requestSchema, responseSchema *base.SchemaProxy) bool {
	_, method, err := parseSubjectMethod(path)
	if err != nil {
		return false
	}

	if !strings.Contains(strings.ToLower(method), "list") {
		return false
	}

	if requestSchema == nil || requestSchema.Schema() == nil {
		return false
	}

	hasOffset := false
	if requestSchema.Schema().Properties != nil {
		for propPair := orderedmap.First(requestSchema.Schema().Properties); propPair != nil; propPair = propPair.Next() {
			if strings.ToLower(propPair.Key()) == "offset" {
				hasOffset = true
				break
			}
		}
	}

	if !hasOffset {
		return false
	}

	_, _, found := p.findFirstArrayField(responseSchema)
	return found
}

func (p *Parser) findFirstArrayField(schema *base.SchemaProxy) (fieldName, itemType string, found bool) {
	if schema == nil || schema.Schema() == nil {
		return "", "", false
	}

	schemaObj := schema.Schema()
	if schemaObj.Properties == nil {
		return "", "", false
	}

	for propPair := orderedmap.First(schemaObj.Properties); propPair != nil; propPair = propPair.Next() {
		propName := propPair.Key()
		propSchema := propPair.Value()

		if propSchema.Schema() != nil && propSchema.Schema().Type != nil {
			if len(propSchema.Schema().Type) > 0 && propSchema.Schema().Type[0] == "array" {
				if propSchema.Schema().Items != nil && propSchema.Schema().Items.IsA() {
					itemSchema := propSchema.Schema().Items.A
					if itemSchema.IsReference() {
						ref := itemSchema.GetReference()
						itemType = extractSchemaName(ref)
						capitalizedFieldName := capitalizeFirst(propName)
						return capitalizedFieldName, itemType, true
					}
				}
			}
		}
	}

	return "", "", false
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (p *Parser) getSchemas(path string) (*base.SchemaProxy, *base.SchemaProxy, error) {
	if p.spec.Paths == nil || p.spec.Paths.PathItems == nil {
		return nil, nil, fmt.Errorf("no paths found")
	}

	pathItem := p.spec.Paths.PathItems.GetOrZero(path)
	if pathItem == nil || pathItem.Post == nil {
		return nil, nil, fmt.Errorf("path not found: %s", path)
	}

	operation := pathItem.Post

	var requestSchema *base.SchemaProxy
	if operation.RequestBody != nil && operation.RequestBody.Content != nil {
		for contentPair := orderedmap.First(operation.RequestBody.Content); contentPair != nil; contentPair = contentPair.Next() {
			mediaType := contentPair.Value()
			if mediaType.Schema != nil {
				if mediaType.Schema.IsReference() {
					ref := mediaType.Schema.GetReference()
					requestSchema = p.resolveSchemaRef(ref)
				}
				break
			}
		}
	}

	var responseSchema *base.SchemaProxy
	if operation.Responses != nil && operation.Responses.Codes != nil {
		for responsePair := orderedmap.First(operation.Responses.Codes); responsePair != nil; responsePair = responsePair.Next() {
			response := responsePair.Value()
			if response.Content != nil {
				for contentPair := orderedmap.First(response.Content); contentPair != nil; contentPair = contentPair.Next() {
					mediaType := contentPair.Value()
					if mediaType.Schema != nil {
						if mediaType.Schema.IsReference() {
							ref := mediaType.Schema.GetReference()
							responseSchema = p.resolveSchemaRef(ref)
						}
						break
					}
				}
			}
			if responseSchema != nil {
				break
			}
		}
	}

	return requestSchema, responseSchema, nil
}

func (p *Parser) resolveSchemaRef(ref string) *base.SchemaProxy {
	schemaName := extractSchemaName(ref)
	if p.spec.Components != nil && p.spec.Components.Schemas != nil {
		return p.spec.Components.Schemas.GetOrZero(schemaName)
	}
	return nil
}

func extractSchemaName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}
