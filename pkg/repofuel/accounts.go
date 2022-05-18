package repofuel

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type AccountsService service

//todo: change the userID from string to UserID
func (s *AccountsService) ProvderToken(ctx context.Context, userID, provider string) (*oauth2.Token, *http.Response, error) {
	u := fmt.Sprintf("users/%s/providers/%s/token", userID, provider)
	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var t oauth2.Token
	resp, err := s.client.Do(req, &t)
	return &t, resp, err
}
