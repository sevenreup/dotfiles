package commands

import (
	"doot/src/internal/calendar"
	"doot/src/internal/server"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func Serve() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the doot daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			srv := server.New()

			if err := srv.Start(); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "doot daemon listening on %s\n", server.SocketPath())

			registerCalendarHandlers(srv)

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig

			srv.Stop()
			return nil
		},
	}
}

func registerCalendarHandlers(srv *server.Server) {
	mgr := calendar.New()

	// calendar.toggle — broadcast toggle event to all clients
	srv.Handle("calendar.toggle", func(_ json.RawMessage) (any, error) {
		srv.Broadcast("calendar.toggle", map[string]bool{})
		return map[string]string{"status": "ok"}, nil
	})

	// calendar.accounts — return list of connected accounts
	srv.Handle("calendar.accounts", func(_ json.RawMessage) (any, error) {
		return mgr.Accounts()
	})

	// calendar.events — fetch events for {year, month}
	srv.Handle("calendar.events", func(params json.RawMessage) (any, error) {
		var p struct {
			Year  int `json:"year"`
			Month int `json:"month"`
		}
		if err := json.Unmarshal(params, &p); err != nil || p.Year == 0 || p.Month == 0 {
			return nil, fmt.Errorf("year and month are required")
		}
		return mgr.Events(p.Year, p.Month)
	})

	// calendar.auth.start — launch OAuth2 browser flow asynchronously
	srv.Handle("calendar.auth.start", func(_ json.RawMessage) (any, error) {
		if !calendar.HasCredentials() {
			return nil, fmt.Errorf("Google OAuth credentials not configured — run: doot calendar setup credentials")
		}
		go func() {
			result := <-calendar.StartAuthFlow()
			if result.Err != nil {
				srv.Broadcast("calendar.authError", map[string]string{"error": result.Err.Error()})
				return
			}
			srv.Broadcast("calendar.authComplete", map[string]string{
				"email": result.Account.Username,
				"name":  result.Account.Name,
			})
		}()
		return map[string]string{"status": "started"}, nil
	})

	// calendar.auth.remove — remove an account and its token
	srv.Handle("calendar.auth.remove", func(params json.RawMessage) (any, error) {
		var p struct {
			Username string `json:"username"`
		}
		if err := json.Unmarshal(params, &p); err != nil || p.Username == "" {
			return nil, fmt.Errorf("username is required")
		}
		return nil, mgr.RemoveAccount(p.Username)
	})
}
