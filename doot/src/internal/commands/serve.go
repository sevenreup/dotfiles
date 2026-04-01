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

	// calendar.accounts — return list of connected accounts (no passwords)
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

	// calendar.account.add — verify CalDAV credentials and save account
	srv.Handle("calendar.account.add", func(params json.RawMessage) (any, error) {
		var p struct {
			Name     string `json:"name"`
			URL      string `json:"url"`
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if p.URL == "" || p.Username == "" || p.Password == "" {
			return nil, fmt.Errorf("url, username and password are required")
		}
		if p.Name == "" {
			p.Name = p.Username
		}
		if err := mgr.AddAccount(p.Name, p.URL, p.Username, p.Password); err != nil {
			return nil, err
		}
		return map[string]string{"status": "added", "username": p.Username}, nil
	})

	// calendar.account.remove — remove a stored account
	srv.Handle("calendar.account.remove", func(params json.RawMessage) (any, error) {
		var p struct {
			Username string `json:"username"`
		}
		if err := json.Unmarshal(params, &p); err != nil || p.Username == "" {
			return nil, fmt.Errorf("username is required")
		}
		return nil, mgr.RemoveAccount(p.Username)
	})
}
