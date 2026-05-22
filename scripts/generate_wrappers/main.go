package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

const (
	readmeBlockStart = "<!-- ROE-SDK:GENERATED-FRIENDLY-APIS:START -->"
	readmeBlockEnd   = "<!-- ROE-SDK:GENERATED-FRIENDLY-APIS:END -->"
)

type contract struct {
	APIs map[string]apiSpec `yaml:"apis"`
}

type apiSpec struct {
	StructName string      `yaml:"struct_name"`
	FieldName  string      `yaml:"field_name"`
	Docstring  string      `yaml:"docstring"`
	Operations []operation `yaml:"operations"`
}

type operation struct {
	Kind       string      `yaml:"kind"`
	MethodName string      `yaml:"method_name"`
	Docstring  string      `yaml:"docstring"`
	Path       string      `yaml:"path"`
	ReturnType string      `yaml:"return_type"`
	BodyType   string      `yaml:"body_type"`
	Parameters []parameter `yaml:"parameters"`
}

type parameter struct {
	Name          string `yaml:"name"`
	WireName      string `yaml:"wire_name"`
	GoType        string `yaml:"go_type"`
	Optional      bool   `yaml:"optional"`
	OmitWhenEmpty bool   `yaml:"omit_when_empty"`
}

type readmeBlocks struct {
	Blocks struct {
		GeneratedFriendlyAPIs struct {
			Go string `yaml:"go"`
		} `yaml:"generated_friendly_apis"`
	} `yaml:"blocks"`
}

func main() {
	root, err := repoRoot()
	must(err)
	modulePath, err := readModulePath(root)
	must(err)

	spec := readContract(filepath.Join(root, "openapi", "wrappers.yml"))
	writeGeneratedAPIs(root, spec)
	for apiName, api := range spec.APIs {
		writeAPI(root, modulePath, apiName, api)
	}
	syncReadmeBlock(root)
	fmt.Printf("Generated %d friendly API wrapper modules from openapi/wrappers.yml\n", len(spec.APIs))
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("could not find go.mod from %s or any parent directory", wd)
		}
		wd = parent
	}
}

func readModulePath(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", err
	}
	modulePath := modfile.ModulePath(data)
	if modulePath == "" {
		return "", fmt.Errorf("go.mod must declare a module path")
	}
	return modulePath, nil
}

func readContract(path string) contract {
	data, err := os.ReadFile(path)
	must(err)
	var spec contract
	must(yaml.Unmarshal(data, &spec))
	if len(spec.APIs) == 0 {
		must(fmt.Errorf("%s must contain apis", path))
	}
	return spec
}

func writeGeneratedAPIs(root string, spec contract) {
	var buf bytes.Buffer
	writeHeader(&buf)
	apiNames := sortedAPINames(spec.APIs)

	buf.WriteString("type generatedAPIs struct {\n")
	for _, apiName := range apiNames {
		api := spec.APIs[apiName]
		fmt.Fprintf(&buf, "\t%s *%s\n", api.FieldName, api.StructName)
	}
	buf.WriteString("}\n\n")

	buf.WriteString("func newGeneratedAPIs(cfg Config, httpClient *httpClient) *generatedAPIs {\n")
	buf.WriteString("\treturn &generatedAPIs{\n")
	for _, apiName := range apiNames {
		api := spec.APIs[apiName]
		fmt.Fprintf(&buf, "\t\t%s: new%s(cfg, httpClient),\n", api.FieldName, api.StructName)
	}
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")

	writeFile(root, "generated_apis.go", buf.Bytes())
}

func writeAPI(root, modulePath, apiName string, api apiSpec) {
	var buf bytes.Buffer
	writeHeader(&buf)
	buf.WriteString("import (\n")
	buf.WriteString("\t\"context\"\n")
	if apiNeedsFmt(api) {
		buf.WriteString("\t\"fmt\"\n")
	}
	buf.WriteString("\n")
	fmt.Fprintf(&buf, "\t%q\n", modulePath+"/generated")
	buf.WriteString(")\n\n")

	fmt.Fprintf(&buf, "// %s %s\n", api.StructName, sentence(api.Docstring))
	fmt.Fprintf(&buf, "type %s struct {\n", api.StructName)
	buf.WriteString("\tcfg        Config\n")
	buf.WriteString("\thttpClient *httpClient\n")
	buf.WriteString("}\n\n")

	fmt.Fprintf(&buf, "func new%s(cfg Config, httpClient *httpClient) *%s {\n", api.StructName, api.StructName)
	fmt.Fprintf(&buf, "\treturn &%s{cfg: cfg, httpClient: httpClient}\n", api.StructName)
	buf.WriteString("}\n\n")

	for index, operation := range api.Operations {
		if index > 0 {
			buf.WriteString("\n")
		}
		switch kind(operation.Kind) {
		case "simple":
			renderSimpleOperation(&buf, api.StructName, operation)
		case "table_upload":
			renderTableUploadOperation(&buf, api.StructName, operation)
		default:
			must(fmt.Errorf("%s.%s has unsupported kind %q", apiName, operation.MethodName, operation.Kind))
		}
	}

	writeFile(root, fmt.Sprintf("%s.go", apiName), buf.Bytes())
}

func renderSimpleOperation(buf *bytes.Buffer, receiver string, op operation) {
	params := goParams(op.Parameters)
	callArgs := goParamNames(op.Parameters)
	if callArgs != "" {
		callArgs = ", " + callArgs
	}

	fmt.Fprintf(buf, "// %s %s\n", op.MethodName, sentence(op.Docstring))
	fmt.Fprintf(buf, "func (a *%s) %s(%s) (%s, error) {\n", receiver, op.MethodName, params, op.ReturnType)
	fmt.Fprintf(buf, "\treturn a.%sWithContext(context.Background()%s)\n", op.MethodName, callArgs)
	buf.WriteString("}\n\n")

	fmt.Fprintf(buf, "// %sWithContext %s\n", op.MethodName, sentence(op.Docstring))
	fmt.Fprintf(buf, "func (a *%s) %sWithContext(ctx context.Context", receiver, op.MethodName)
	for _, param := range op.Parameters {
		fmt.Fprintf(buf, ", %s %s", param.Name, param.GoType)
	}
	fmt.Fprintf(buf, ") (%s, error) {\n", op.ReturnType)
	writeQueryMap(buf, op.Parameters)
	fmt.Fprintf(buf, "\tvar resp %s\n", op.ReturnType)
	fmt.Fprintf(buf, "\tif err := a.httpClient.getWithContext(ctx, %q, query, &resp); err != nil {\n", op.Path)
	fmt.Fprintf(buf, "\t\treturn %s{}, err\n", op.ReturnType)
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn resp, nil\n")
	buf.WriteString("}\n")
}

func renderTableUploadOperation(buf *bytes.Buffer, receiver string, op operation) {
	fmt.Fprintf(buf, "// %s %s\n", op.MethodName, sentence(op.Docstring))
	fmt.Fprintf(buf, "func (a *%s) %s(tableName string, file FileUpload, withHeaders bool) (%s, error) {\n", receiver, op.MethodName, op.ReturnType)
	fmt.Fprintf(buf, "\treturn a.%sWithContext(context.Background(), tableName, file, withHeaders)\n", op.MethodName)
	buf.WriteString("}\n\n")

	fmt.Fprintf(buf, "// %sWithContext %s\n", op.MethodName, sentence(op.Docstring))
	fmt.Fprintf(buf, "func (a *%s) %sWithContext(ctx context.Context, tableName string, file FileUpload, withHeaders bool) (%s, error) {\n", receiver, op.MethodName, op.ReturnType)
	buf.WriteString("\tinputs := map[string]any{\n")
	buf.WriteString("\t\t\"table_name\": tableName,\n")
	buf.WriteString("\t\t\"file\": file,\n")
	buf.WriteString("\t\t\"with_headers\": withHeaders,\n")
	buf.WriteString("\t\t\"organization_id\": a.cfg.OrganizationID,\n")
	buf.WriteString("\t}\n")
	fmt.Fprintf(buf, "\tvar resp %s\n", op.ReturnType)
	fmt.Fprintf(buf, "\tif err := a.httpClient.postDynamicInputsWithContext(ctx, %q, inputs, nil, &resp, nil); err != nil {\n", op.Path)
	fmt.Fprintf(buf, "\t\treturn %s{}, err\n", op.ReturnType)
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn resp, nil\n")
	buf.WriteString("}\n")
}

func writeQueryMap(buf *bytes.Buffer, params []parameter) {
	buf.WriteString("\tquery := map[string]string{}\n")
	for _, param := range params {
		wireName := param.WireName
		if wireName == "" {
			wireName = param.Name
		}
		if param.OmitWhenEmpty {
			condition, err := omitWhenEmptyCondition(param)
			must(err)
			fmt.Fprintf(buf, "\tif %s {\n", condition)
			fmt.Fprintf(buf, "\t\tquery[%q] = fmt.Sprint(%s)\n", wireName, param.Name)
			buf.WriteString("\t}\n")
		} else {
			fmt.Fprintf(buf, "\tquery[%q] = fmt.Sprint(%s)\n", wireName, param.Name)
		}
	}
}

func omitWhenEmptyCondition(param parameter) (string, error) {
	switch param.GoType {
	case "string":
		return fmt.Sprintf("%s != \"\"", param.Name), nil
	case "bool":
		return param.Name, nil
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64":
		return fmt.Sprintf("%s != 0", param.Name), nil
	default:
		return "", fmt.Errorf("%s has omit_when_empty with unsupported go_type %q", param.Name, param.GoType)
	}
}

func apiNeedsFmt(api apiSpec) bool {
	for _, op := range api.Operations {
		if len(op.Parameters) > 0 {
			return true
		}
	}
	return false
}

func goParams(params []parameter) string {
	parts := make([]string, 0, len(params))
	for _, param := range params {
		parts = append(parts, fmt.Sprintf("%s %s", param.Name, param.GoType))
	}
	return strings.Join(parts, ", ")
}

func goParamNames(params []parameter) string {
	names := make([]string, 0, len(params))
	for _, param := range params {
		names = append(names, param.Name)
	}
	return strings.Join(names, ", ")
}

func syncReadmeBlock(root string) {
	data, err := os.ReadFile(filepath.Join(root, "openapi", "readme_blocks.yml"))
	must(err)
	var blocks readmeBlocks
	must(yaml.Unmarshal(data, &blocks))
	block := strings.TrimSpace(blocks.Blocks.GeneratedFriendlyAPIs.Go)
	if block == "" {
		must(fmt.Errorf("openapi/readme_blocks.yml must contain blocks.generated_friendly_apis.go"))
	}

	readmePath := filepath.Join(root, "README.md")
	readmeBytes, err := os.ReadFile(readmePath)
	must(err)
	readme := string(readmeBytes)
	start := strings.Index(readme, readmeBlockStart)
	end := strings.Index(readme, readmeBlockEnd)
	if start < 0 || end < start {
		must(fmt.Errorf("README.md must contain %s and %s", readmeBlockStart, readmeBlockEnd))
	}
	end += len(readmeBlockEnd)
	updated := readme[:start] + readmeBlockStart + "\n" + block + "\n" + readmeBlockEnd + readme[end:]
	must(os.WriteFile(readmePath, []byte(updated), 0o644))
}

func writeHeader(buf *bytes.Buffer) {
	buf.WriteString("// Code generated by scripts/generate-sdk from openapi/wrappers.yml. DO NOT EDIT.\n")
	buf.WriteString("package roe\n\n")
}

func writeFile(root, relpath string, data []byte) {
	must(os.WriteFile(filepath.Join(root, relpath), data, 0o644))
}

func sortedAPINames(apis map[string]apiSpec) []string {
	names := make([]string, 0, len(apis))
	for name := range apis {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func kind(value string) string {
	if value == "" {
		return "simple"
	}
	return value
}

func sentence(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "wraps a generated Roe API operation."
	}
	if strings.HasSuffix(value, ".") {
		return value
	}
	return value + "."
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
