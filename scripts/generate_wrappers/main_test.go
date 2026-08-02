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

func TestOmitWhenEmptyConditionSupportsTypedMaps(t *testing.T) {
	// map[string]string reached the contract via roe-main's connector-level
	// dynamic runtime inputs and panicked the generator, stalling the Go
	// release fan-out. Any map type must produce a condition, not an error.
	cases := map[string]string{
		"map[string]any":    "dynamicInputs != nil",
		"map[string]string": "len(dynamicInputs) > 0",
		"map[string]int":    "len(dynamicInputs) > 0",
	}

	for goType, want := range cases {
		got, err := omitWhenEmptyCondition(parameter{Name: "dynamicInputs", GoType: goType})
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", goType, err)
		}
		if got != want {
			t.Fatalf("%s: expected %s, got %s", goType, want, got)
		}
	}

	if _, err := omitWhenEmptyCondition(parameter{Name: "thing", GoType: "chan int"}); err == nil {
		t.Fatal("expected an error for a genuinely unsupported go_type")
	}
}
