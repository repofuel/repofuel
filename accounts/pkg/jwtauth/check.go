// Copyright (c) 2019. Suhaib Mujahid. All rights reserved.
// You cannot use this source code without a permission.

package jwtauth

import (
	"context"
	"crypto"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/repofuel/repofuel/accounts/pkg/permission"
)

var (
	ErrUnexpectedSignature = errors.New("unexpected signing method")
	ErrUnknownIssuer       = errors.New("unknown token issuer")
)

type AuthCheck struct {
	src KeySource
}

type LocalKeySource struct {
	keys map[string]crypto.PublicKey
}

func NewLocalKeySource(keys map[string]crypto.PublicKey) *LocalKeySource {
	if keys == nil {
		keys = make(map[string]crypto.PublicKey)
	}
	return &LocalKeySource{
		keys: keys,
	}
}

func (s *LocalKeySource) AddKey(issuer string, k crypto.PublicKey) *LocalKeySource {
	s.keys[issuer] = k
	return s
}

func (s *LocalKeySource) PublicKey(issuer string) (crypto.PublicKey, error) {
	k, ok := s.keys[issuer]
	if !ok {
		return nil, ErrUnknownIssuer
	}
	return k, nil
}

type KeySource interface {
	PublicKey(issuer string) (crypto.PublicKey, error)
}

func NewAuthCheck(src KeySource) *AuthCheck {
	return &AuthCheck{
		src: src,
	}
}

func (auth *AuthCheck) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := auth.AuthenticatedContext(r.Context(), tokenFromRequest(r))
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func tokenFromRequest(r *http.Request) string {
	accessToken := StripBearerToken(r.Header.Get("Authorization"))
	if accessToken != "" {
		return accessToken
	}

	return tokenFromCookie(r)
}

func (auth *AuthCheck) AuthenticatedContext(ctx context.Context, token string) (context.Context, error) {
	if token == "" {
		return nil, errors.New("missing bearer authentication token")
	}

	claims, err := auth.ClaimsFromToken(token)
	if err != nil {
		return nil, err
	}

	if claims.AccessInfo == nil {
		return nil, errors.New("no access info")
	}

	return context.WithValue(ctx, permission.ViewerCtxKey, claims.AccessInfo), nil
}

func (auth *AuthCheck) ClaimsFromToken(strToken string) (*AccessClaims, error) {
	var claims AccessClaims
	_, err := jwt.ParseWithClaims(strToken, &claims, auth.keyFunc)
	if err != nil {
		return nil, err
	}

	return &claims, nil
}

// Validate that the `alg` is the expect method and return the key.
func (auth *AuthCheck) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
		return nil, ErrUnexpectedSignature
	}
	c := token.Claims.(*AccessClaims)
	return auth.src.PublicKey(c.Issuer)
}

func StripBearerToken(auth string) string {
	const prefix = "Bearer "
	if len(auth) > len(prefix) && strings.EqualFold(auth[0:len(prefix)], prefix) {
		return auth[len(prefix):]
	}
	return ""
}

func tokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		return ""
	}
	return cookie.Value
}
