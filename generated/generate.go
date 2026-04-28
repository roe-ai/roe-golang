package generated

//go:generate go run ../scripts/normalize_openapi/main.go -input ../openapi/openapi.yml -output ../openapi/oapi-codegen-input.yml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config ../openapi/oapi-codegen.yaml ../openapi/oapi-codegen-input.yml
