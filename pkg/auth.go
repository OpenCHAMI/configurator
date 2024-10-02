package configurator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/OpenCHAMI/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func VerifyScope(testScopes []string, r *http.Request) (bool, error) {
	// extract the scopes from JWT
	var scopes []string
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return false, fmt.Errorf("failed to get claim(s) from token: %v", err)
	}

	appendScopes := func(slice []string, scopeClaim any) []string {
		switch scopeClaim.(type) {
		case []any:
			// convert all scopes to str and append
			for _, s := range scopeClaim.([]any) {
				switch s.(type) {
				case string:
					slice = append(slice, s.(string))
				}
			}
		case []string:
			slice = append(slice, scopeClaim.([]string)...)
		}
		return slice
	}

	// check for and append both "scp" and "scope" claims
	v, ok := claims["scp"]
	if ok {
		scopes = appendScopes(scopes, v)
	}
	v, ok = claims["scope"]
	if ok {
		scopes = appendScopes(scopes, v)
	}

	// check for both 'scp' and 'scope' claims for scope
	scopeClaim, ok := claims["scp"]
	if ok {
		switch scopeClaim.(type) {
		case []any:
			// convert all scopes to str and append
			for _, s := range scopeClaim.([]any) {
				switch s.(type) {
				case string:
					scopes = append(scopes, s.(string))
				}
			}
		case []string:
			scopes = append(scopes, scopeClaim.([]string)...)
		}
	}
	scopeClaim, ok = claims["scope"]
	if ok {
		scopes = append(scopes, scopeClaim.([]string)...)
	}

	// verify that each of the test scopes are included
	for _, testScope := range testScopes {
		index := slices.Index(scopes, testScope)
		if index < 0 {
			return false, fmt.Errorf("invalid or missing scope")
		}
	}
	// NOTE: should this be ok if no scopes were found?
	return true, nil
}

func FetchPublicKeyFromURL(url string) (*jwtauth.JWTAuth, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	set, err := jwk.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	jwks, err := json.Marshal(set)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWKS: %v", err)
	}
	tokenAuth, err := jwtauth.NewKeySet(jwks)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWKS: %v", err)
	}

	return tokenAuth, nil
}

func LoadAccessToken() {

}
