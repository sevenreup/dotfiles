package commands

import (
	"bufio"
	"doot/src/internal/calendar"
	"doot/src/internal/server"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func Calendar() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Google Calendar integration",
	}
	cmd.AddCommand(
		calendarToggle(),
		calendarAuth(),
		calendarSetup(),
	)
	return cmd
}

// calendarToggle sends a toggle request to the running doot daemon.
func calendarToggle() *cobra.Command {
	return &cobra.Command{
		Use:   "toggle",
		Short: "Toggle the calendar widget",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := net.Dial("unix", server.SocketPath())
			if err != nil {
				return fmt.Errorf("doot daemon not running (%w)", err)
			}
			defer conn.Close()

			req, _ := json.Marshal(map[string]any{
				"type":   "request",
				"id":     "cli-toggle",
				"method": "calendar.toggle",
				"data":   map[string]any{},
			})
			fmt.Fprintf(conn, "%s\n", req)
			bufio.NewScanner(conn).Scan()
			return nil
		},
	}
}

func calendarAuth() *cobra.Command {
	acct := &cobra.Command{
		Use:   "auth",
		Short: "Manage Google Calendar accounts",
	}

	acct.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List connected accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			accounts, err := calendar.ListAccounts()
			if err != nil {
				return err
			}
			if len(accounts) == 0 {
				fmt.Println("No accounts. Run: doot calendar auth add")
				return nil
			}
			for _, a := range accounts {
				fmt.Printf("  %-20s  %s\n  → %s\n\n", a.Name, a.Username, a.URL)
			}
			return nil
		},
	})

	acct.AddCommand(&cobra.Command{
		Use:   "remove [username]",
		Short: "Remove an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := calendar.New()
			if err := mgr.RemoveAccount(args[0]); err != nil {
				return err
			}
			fmt.Printf("Removed account %s\n", args[0])
			return nil
		},
	})

	return acct
}

func calendarSetup() *cobra.Command {
	setup := &cobra.Command{
		Use:   "setup",
		Short: "Setup and configuration",
	}

	var flagClientID, flagClientSecret string

	credsCmd := &cobra.Command{
		Use:   "credentials",
		Short: "Store Google OAuth2 client credentials in the keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := bufio.NewReader(os.Stdin)

			if flagClientID == "" {
				fmt.Fprint(os.Stderr, "Google OAuth2 Client ID: ")
				flagClientID, _ = r.ReadString('\n')
				flagClientID = strings.TrimSpace(flagClientID)
			}
			if flagClientSecret == "" {
				fmt.Fprint(os.Stderr, "Google OAuth2 Client Secret: ")
				flagClientSecret, _ = r.ReadString('\n')
				flagClientSecret = strings.TrimSpace(flagClientSecret)
			}
			if flagClientID == "" || flagClientSecret == "" {
				return fmt.Errorf("client ID and secret are required")
			}

			if err := calendar.StoreCredentials(flagClientID, flagClientSecret); err != nil {
				return fmt.Errorf("failed to store credentials: %w", err)
			}
			fmt.Println("✓ Credentials stored in keyring (visible in Seahorse under 'doot')")
			return nil
		},
	}
	credsCmd.Flags().StringVar(&flagClientID, "client-id", "", "Google OAuth2 Client ID")
	credsCmd.Flags().StringVar(&flagClientSecret, "client-secret", "", "Google OAuth2 Client Secret")

	setup.AddCommand(credsCmd)

	setup.AddCommand(&cobra.Command{
		Use:   "instructions",
		Short: "Show setup instructions",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(`Google Calendar Setup
=====================

1. Go to: https://console.cloud.google.com/
2. Create a project (or select an existing one)
3. Enable the "Calendar API" (for CalDAV) or "People API" (for user info)
4. Go to: APIs & Services → Credentials
5. Create an OAuth 2.0 Client ID (Desktop app)
6. Download or copy the Client ID and Client Secret
7. Run: doot calendar setup credentials

Then sign in:
  doot calendar toggle   (opens the calendar widget with a Sign In button)
  — or —
  Connect via the doot daemon (the widget will launch the browser flow)

`)
		},
	})

	return setup
}
