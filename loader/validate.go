package loader

import (
	"encoding/json"
	"fmt"
	"strings"

	_ "embed" // for the embedded schema

	"github.com/pkg/errors"
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
		return errors.Wrap(err, "failed to marshal pack file to JSON")
	}

	// Validate against schema
	documentLoader := gojsonschema.NewBytesLoader(jsonBytes)
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return errors.Wrap(err, "schema validation failed")
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
		return errors.Errorf("schema validation failed:\n  %s", strings.Join(errMsgs, "\n  "))
	}

	return nil
}
