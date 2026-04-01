package calendar

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

// savingTokenSource wraps an oauth2.TokenSource and persists refreshed tokens
// back to the keyring so they survive process restarts.
type savingTokenSource struct {
	src   oauth2.TokenSource
	email string
	last  *oauth2.Token
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.src.Token()
	if err != nil {
		return nil, err
	}
	// Persist only when the token actually changed (i.e. was refreshed).
	if s.last == nil || t.AccessToken != s.last.AccessToken {
		if serr := StoreToken(s.email, t); serr != nil {
			fmt.Printf("[calendar] warning: could not persist refreshed token for %s: %v\n", s.email, serr)
		}
		s.last = t
	}
	return t, nil
}

// newBearerClient returns an *http.Client that authenticates CalDAV requests
// with the OAuth2 Bearer token stored in the keyring for the given email.
// Refreshes are handled automatically and persisted back to the keyring.
func newBearerClient(email string) (*http.Client, error) {
	stored, err := LoadToken(email)
	if err != nil {
		return nil, err
	}

	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	base := cfg.TokenSource(context.Background(), stored)
	src := &savingTokenSource{src: base, email: email, last: stored}
	return oauth2.NewClient(context.Background(), src), nil
}
