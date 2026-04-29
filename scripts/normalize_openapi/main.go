package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	inputPath := flag.String("input", "", "Path to the source OpenAPI spec")
	outputPath := flag.String("output", "", "Path to write the normalized OpenAPI spec")
	flag.Parse()

	if *inputPath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "normalize_openapi requires --input and --output")
		os.Exit(1)
	}

	raw, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read spec: %v\n", err)
		os.Exit(1)
	}

	var document map[string]any
	if err := yaml.Unmarshal(raw, &document); err != nil {
		fmt.Fprintf(os.Stderr, "parse yaml: %v\n", err)
		os.Exit(1)
	}

	normalizeNode(document)

	if openapi, ok := document["openapi"].(string); ok && strings.HasPrefix(openapi, "3.1") {
		document["openapi"] = "3.0.3"
	}

	output, err := yaml.Marshal(document)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal yaml: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputPath, output, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write normalized spec: %v\n", err)
		os.Exit(1)
	}
}

func normalizeNode(value any) {
	switch typed := value.(type) {
	case map[string]any:
		if schemaType, ok := typed["type"].([]any); ok {
			normalizedType, nullable, convertible := normalizeNullableType(schemaType)
			if convertible {
				typed["type"] = normalizedType
				if nullable {
					typed["nullable"] = true
				}
			}
		}
		normalizeNullableUnion(typed, "oneOf")
		normalizeNullableUnion(typed, "anyOf")
		for _, child := range typed {
			normalizeNode(child)
		}
	case []any:
		for _, child := range typed {
			normalizeNode(child)
		}
	}
}

func normalizeNullableType(schemaType []any) (string, bool, bool) {
	if len(schemaType) != 2 {
		return "", false, false
	}

	var nonNullType string
	var sawNull bool

	for _, entry := range schemaType {
		typeName, ok := entry.(string)
		if !ok {
			return "", false, false
		}
		if typeName == "null" {
			sawNull = true
			continue
		}
		if nonNullType != "" {
			return "", false, false
		}
		nonNullType = typeName
	}

	if !sawNull || nonNullType == "" {
		return "", false, false
	}
	return nonNullType, true, true
}

func normalizeNullableUnion(node map[string]any, key string) {
	union, ok := node[key].([]any)
	if !ok || len(union) != 2 {
		return
	}

	var (
		nullIndex = -1
		other     map[string]any
	)

	for index, entry := range union {
		schema, ok := entry.(map[string]any)
		if !ok {
			return
		}
		if typeName, ok := schema["type"].(string); ok && typeName == "null" {
			nullIndex = index
			continue
		}
		other = schema
	}

	if nullIndex == -1 || other == nil {
		return
	}

	delete(node, key)
	for childKey, childValue := range other {
		node[childKey] = childValue
	}
	node["nullable"] = true
}
