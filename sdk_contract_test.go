package roe

import (
	"os"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

type wrapperContract struct {
	APIs map[string]wrapperAPI `yaml:"apis"`
}

type wrapperAPI struct {
	Operations []wrapperOperation    `yaml:"operations"`
	Namespaces map[string]wrapperAPI `yaml:"namespaces"`
}

type wrapperOperation struct {
	MethodName string `yaml:"method_name"`
	Method     string `yaml:"method"`
	Path       string `yaml:"path"`
}

type operationTransport struct {
	Method string
	Path   string
}

func TestHandMaintainedWrapperContractsExposeMethods(t *testing.T) {
	data, err := os.ReadFile("openapi/wrappers.yml")
	if err != nil {
		t.Fatalf("read wrappers contract: %v", err)
	}

	var contract wrapperContract
	if err := yaml.Unmarshal(data, &contract); err != nil {
		t.Fatalf("parse wrappers contract: %v", err)
	}

	handMaintained := map[string]reflect.Type{
		"agents":   reflect.TypeOf(&AgentsAPI{}),
		"policies": reflect.TypeOf(&PoliciesAPI{}),
	}
	namespaces := map[string]reflect.Type{
		"agents.jobs":       reflect.TypeOf(&AgentJobsAPI{}),
		"agents.versions":   reflect.TypeOf(&AgentVersionsAPI{}),
		"policies.versions": reflect.TypeOf(&PolicyVersionsAPI{}),
	}

	for apiName, api := range contract.APIs {
		apiType, ok := handMaintained[apiName]
		if !ok {
			continue
		}
		assertWrapperMethods(t, apiName, apiType, api.Operations)
		assertWrapperTransports(t, apiName, api.Operations)
		for namespaceName, namespace := range api.Namespaces {
			label := apiName + "." + namespaceName
			namespaceType, ok := namespaces[label]
			if !ok {
				t.Fatalf("no public Go type registered for hand-maintained wrapper namespace %s", label)
			}
			assertWrapperMethods(t, label, namespaceType, namespace.Operations)
			assertWrapperTransports(t, label, namespace.Operations)
		}
	}
}

func assertWrapperMethods(t *testing.T, label string, apiType reflect.Type, operations []wrapperOperation) {
	t.Helper()
	for _, operation := range operations {
		if operation.MethodName == "" {
			t.Fatalf("%s contains an operation without method_name", label)
		}
		if _, ok := apiType.MethodByName(operation.MethodName); !ok {
			t.Fatalf("%s contract method %s is not implemented on %s", label, operation.MethodName, apiType)
		}
	}
}

func assertWrapperTransports(t *testing.T, label string, operations []wrapperOperation) {
	t.Helper()
	expectedByMethod, ok := handMaintainedTransports[label]
	if !ok {
		t.Fatalf("no transport contract registered for hand-maintained wrapper %s", label)
	}
	seen := map[string]bool{}
	for _, operation := range operations {
		if operation.Method == "" && operation.Path == "" {
			continue
		}
		seen[operation.MethodName] = true
		expected, ok := expectedByMethod[operation.MethodName]
		if !ok {
			t.Fatalf("%s contract method %s has no expected transport", label, operation.MethodName)
		}
		if operation.Method != expected.Method || operation.Path != expected.Path {
			t.Fatalf(
				"%s.%s transport drift: contract has %s %s, expected %s %s",
				label,
				operation.MethodName,
				operation.Method,
				operation.Path,
				expected.Method,
				expected.Path,
			)
		}
	}
	for methodName := range expectedByMethod {
		if !seen[methodName] {
			t.Fatalf("%s expected transport %s is not declared in wrapper contract", label, methodName)
		}
	}
}

var handMaintainedTransports = map[string]map[string]operationTransport{
	"agents": {
		"List":           {Method: "GET", Path: "/v1/agents/"},
		"Retrieve":       {Method: "GET", Path: "/v1/agents/{agent_id}/"},
		"Create":         {Method: "POST", Path: "/v1/agents/"},
		"Update":         {Method: "PUT", Path: "/v1/agents/{agent_id}/"},
		"Delete":         {Method: "DELETE", Path: "/v1/agents/{agent_id}/"},
		"Duplicate":      {Method: "POST", Path: "/v1/agents/{agent_id}/duplicate/"},
		"Run":            {Method: "POST", Path: "/v1/agents/run/{agent_id}/async/"},
		"RunMany":        {Method: "POST", Path: "/v1/agents/run/{agent_id}/async/many/"},
		"RunSync":        {Method: "POST", Path: "/v1/agents/run/{agent_id}/"},
		"RunVersion":     {Method: "POST", Path: "/v1/agents/run/{agent_id}/versions/{agent_version_id}/async/"},
		"RunVersionSync": {Method: "POST", Path: "/v1/agents/run/{agent_id}/versions/{agent_version_id}/"},
	},
	"agents.versions": {
		"List":                    {Method: "GET", Path: "/v1/agents/{agent_id}/versions/"},
		"Retrieve":                {Method: "GET", Path: "/v1/agents/{agent_id}/versions/{agent_version_id}/"},
		"RetrieveCurrent":         {Method: "GET", Path: "/v1/agents/{agent_id}/versions/current/"},
		"Create":                  {Method: "POST", Path: "/v1/agents/{agent_id}/versions/"},
		"Update":                  {Method: "PUT", Path: "/v1/agents/{agent_id}/versions/{agent_version_id}/"},
		"Delete":                  {Method: "DELETE", Path: "/v1/agents/{agent_id}/versions/{agent_version_id}/"},
		"ListPaginated":           {Method: "GET", Path: "/v1/agents/{agent_id}/versions/"},
		"RetrieveCurrentWithEval": {Method: "GET", Path: "/v1/agents/{agent_id}/versions/current/"},
	},
	"agents.jobs": {
		"RetrieveStatus":     {Method: "GET", Path: "/v1/agents/jobs/{job_id}/status/"},
		"RetrieveResult":     {Method: "GET", Path: "/v1/agents/jobs/{agent_job_id}/result/"},
		"Cancel":             {Method: "POST", Path: "/v1/agents/jobs/{job_id}/cancel/"},
		"CancelAll":          {Method: "POST", Path: "/v1/agents/{agent_id}/jobs/cancel-all/"},
		"DeleteData":         {Method: "POST", Path: "/v1/agents/jobs/{job_id}/delete-data/"},
		"RetrieveStatusMany": {Method: "POST", Path: "/v1/agents/jobs/statuses/"},
		"RetrieveResultMany": {Method: "POST", Path: "/v1/agents/jobs/results/"},
		"DownloadReference":  {Method: "GET", Path: "/v1/agents/jobs/{agent_job_id}/references/{resource_id}/"},
	},
	"policies": {
		"List":     {Method: "GET", Path: "/v1/policies/"},
		"Retrieve": {Method: "GET", Path: "/v1/policies/{id}/"},
		"Create":   {Method: "POST", Path: "/v1/policies/"},
		"Update":   {Method: "PATCH", Path: "/v1/policies/{id}/"},
		"Delete":   {Method: "DELETE", Path: "/v1/policies/{id}/"},
	},
	"policies.versions": {
		"List":     {Method: "GET", Path: "/v1/policies/{policy_id}/versions/"},
		"Retrieve": {Method: "GET", Path: "/v1/policies/{policy_id}/versions/{version_id}/"},
		"Create":   {Method: "POST", Path: "/v1/policies/{policy_id}/versions/"},
	},
}
