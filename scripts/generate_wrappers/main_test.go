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
