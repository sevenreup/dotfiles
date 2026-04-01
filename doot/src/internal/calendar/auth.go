package calendar

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"

	"golang.org/x/oauth2"
)

// AuthResult is returned when an OAuth2 flow completes or fails.
type AuthResult struct {
	Account *Account
	Err     error
}

// StartAuthFlow launches the browser OAuth2 flow asynchronously.
// Returns a channel that receives the result.
func StartAuthFlow() <-chan AuthResult {
	ch := make(chan AuthResult, 1)
	go func() {
		acc, err := runAuthFlow()
		ch <- AuthResult{Account: acc, Err: err}
	}()
	return ch
}

func runAuthFlow() (*Account, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Bind callback server first to get a stable port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("could not start callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	cfg.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	stateBytes := make([]byte, 16)
	rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			select {
			case errCh <- fmt.Errorf("OAuth state mismatch"):
			default:
			}
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		if e := r.URL.Query().Get("error"); e != "" {
			select {
			case errCh <- fmt.Errorf("OAuth denied: %s", e):
			default:
			}
			http.Error(w, e, http.StatusUnauthorized)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			select {
			case errCh <- fmt.Errorf("missing authorization code"):
			default:
			}
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, authSuccessHTML)
		select {
		case codeCh <- code:
		default:
		}
	})

	go srv.Serve(listener)
	defer srv.Close()

	exec.Command("xdg-open", authURL).Start()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("auth timed out after 5 minutes")
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	// Fetch user profile to get email + name
	client := cfg.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	var info struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}
	if info.Email == "" {
		return nil, fmt.Errorf("could not determine account email")
	}

	// Store token securely in keyring
	if err := StoreToken(info.Email, token); err != nil {
		return nil, fmt.Errorf("failed to store token in keyring: %w", err)
	}

	// Save non-secret account info to disk
	acc := &Account{
		Name:     info.Name,
		URL:      GoogleCalDAVURL(info.Email),
		Username: info.Email,
	}
	if err := SaveAccount(acc); err != nil {
		return nil, fmt.Errorf("failed to save account: %w", err)
	}
	return acc, nil
}

const authSuccessHTML = `<!doctype html>
<html><head><title>Signed in — doot</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, sans-serif; background: #1e1e2e; color: #cdd6f4;
         display: flex; align-items: center; justify-content: center; height: 100vh; }
  .card { text-align: center; padding: 2rem; }
  .check { font-size: 3rem; margin-bottom: 1rem; }
  h2 { color: #a6e3a1; margin-bottom: .5rem; }
  p { color: #6c7086; font-size: .9rem; }
</style></head>
<body><div class="card">
  <div class="check">✓</div>
  <h2>Signed in successfully</h2>
  <p>You can close this tab and return to your desktop.</p>
</div></body></html>`
