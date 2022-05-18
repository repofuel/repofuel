// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package tokens

import (
	"golang.org/x/oauth2"
	"time"
)

type TokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

func OauthToTokenJSON(t *oauth2.Token) *TokenJSON {
	return &TokenJSON{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		ExpiresIn:    expiryIn(t.Expiry),
	}
}

func expiryIn(t time.Time) int32 {
	if t.IsZero() {
		return 0
	}

	return int32(t.Sub(time.Now()).Seconds())
}
