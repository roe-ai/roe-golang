package roe

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/roe-ai/roe-golang/generated"
)

// UsersAPI exposes user-scoped endpoints. The 37-op SDK surface only includes
// GET /v1/users/current_user/, which returns the authenticated user's info.
type UsersAPI struct {
	client *RoeClient
}

func newUsersAPI(client *RoeClient) *UsersAPI {
	return &UsersAPI{client: client}
}

func (u *UsersAPI) raw() *generated.ClientWithResponses { return u.client.raw }

// Me returns information about the currently authenticated user.
func (u *UsersAPI) Me() (*generated.UserInfo, error) {
	return u.MeWithContext(context.Background())
}

// MeWithContext returns information about the currently authenticated user
// with a caller-supplied context.
func (u *UsersAPI) MeWithContext(ctx context.Context) (*generated.UserInfo, error) {
	resp, err := u.raw().V1UsersCurrentUserRetrieveWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if err := errorFromResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	// The OpenAPI spec declares "no response body" for this endpoint, so the
	// generated wrapper exposes the body as raw bytes. The backend does send
	// a JSON UserInfo payload — decode it here.
	var info generated.UserInfo
	if len(resp.Body) == 0 {
		return nil, fmt.Errorf("retrieve current user: empty response body")
	}
	if err := json.Unmarshal(resp.Body, &info); err != nil {
		return nil, fmt.Errorf("retrieve current user: parse response body: %w", err)
	}
	return &info, nil
}
