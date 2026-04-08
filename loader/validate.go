package loader

import (
	"encoding/json"
	"fmt"
	"strings"

	_ "embed" // for the embedded schema

	"github.com/sentrie-sh/sentrie/pack"
	"github.com/xeipuuv/gojsonschema"
)

//go:embed schema.json
var schemaJSON []byte

var (
	schemaLoader *gojsonschema.SchemaLoader
	schema       *gojsonschema.Schema
)

func init() {
	schemaLoader = gojsonschema.NewSchemaLoader()
	schemaLoader.Draft = gojsonschema.Draft7
	var err error
	schema, err = schemaLoader.Compile(gojsonschema.NewBytesLoader(schemaJSON))
	if err != nil {
		panic(fmt.Sprintf("failed to compile JSON schema: %v", err))
	}
}

func ValidatePackFile(packFile *pack.PackFile) error {
	// Convert PackFile to JSON
	jsonBytes, err := json.Marshal(packFile)
	if err != nil {
		return fmt.Errorf("failed to marshal pack file to JSON: %w", err)
	}

	// Validate against schema
	documentLoader := gojsonschema.NewBytesLoader(jsonBytes)
	result, err := validatePackDocument(documentLoader)
	if err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if !result.Valid() {
		var errMsgs []string
		for _, desc := range result.Errors() {
			field := desc.Field()
			if field == "(root)" {
				field = "root"
			}
			errMsgs = append(errMsgs, fmt.Sprintf("%s: %s", field, desc.Description()))
		}
		return fmt.Errorf("schema validation failed:\n  %s", strings.Join(errMsgs, "\n  "))
	}

	return nil
}

// validatePackDocument runs JSON-schema validation. Swappable in tests.
var validatePackDocument = func(loader gojsonschema.JSONLoader) (*gojsonschema.Result, error) {
	return schema.Validate(loader)
}
