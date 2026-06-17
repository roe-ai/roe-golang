package roe

import (
	"context"

	"github.com/roe-ai/roe-golang/generated"
)

// UsersAPI provides authenticated user account operations.
type UsersAPI struct {
	cfg        Config
	httpClient *httpClient
}

func newUsersAPI(cfg Config, httpClient *httpClient) *UsersAPI {
	return &UsersAPI{cfg: cfg, httpClient: httpClient}
}

// Me retrieves the currently authenticated user.
func (u *UsersAPI) Me() (generated.User, error) {
	return u.MeWithContext(context.Background())
}

// MeWithContext retrieves the currently authenticated user with a caller-supplied context.
func (u *UsersAPI) MeWithContext(ctx context.Context) (generated.User, error) {
	var resp generated.User
	if err := u.httpClient.getWithContext(ctx, "/v1/users/current_user/", nil, &resp); err != nil {
		return generated.User{}, err
	}
	return resp, nil
}
