package main

import "testing"

func TestOperationPathExpressionUsesPathTemplateOrder(t *testing.T) {
	op := operation{
		MethodName: "RunVersion",
		Path:       "/v1/agents/run/{agent_id}/versions/{agent_version_id}/",
		Parameters: []parameter{
			{
				Name:     "versionID",
				Location: "path",
				WireName: "agent_version_id",
			},
			{
				Name:     "agentID",
				Location: "path",
				WireName: "agent_id",
			},
		},
	}

	got := operationPathExpression(op)
	want := `fmt.Sprintf("/v1/agents/run/%s/versions/%s/", agentID, versionID)`
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestZeroValueSupportsPointerReturnTypes(t *testing.T) {
	if got := zeroValue("*generated.Connection"); got != "nil" {
		t.Fatalf("expected nil zero value for pointer return type, got %s", got)
	}
}

func TestIsGeneratedAPISkipsManualNamespaces(t *testing.T) {
	// A namespace with a "manual" operation is hand-written; the generator must
	// skip it (return false) rather than panic on the unsupported kind.
	manual := apiSpec{
		Operations: []operation{{MethodName: "List", Kind: "manual"}},
	}
	if isGeneratedAPI(manual) {
		t.Fatal("expected manual namespace to be skipped")
	}

	// Namespaces with nested namespaces are hand-written too.
	nested := apiSpec{
		Namespaces: map[string]apiSpec{"versions": {}},
	}
	if isGeneratedAPI(nested) {
		t.Fatal("expected namespace with nested namespaces to be skipped")
	}

	// A flat namespace whose operations are all generated kinds is generated.
	generated := apiSpec{
		Operations: []operation{
			{MethodName: "List", Kind: "simple"},
			{MethodName: "Create", Kind: "body"},
		},
	}
	if !isGeneratedAPI(generated) {
		t.Fatal("expected generated namespace to be generated")
	}
}
