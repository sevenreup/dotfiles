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
		Short: "CalDAV calendar integration",
	}
	cmd.AddCommand(
		calendarToggle(),
		calendarAccount(),
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

func calendarAccount() *cobra.Command {
	acct := &cobra.Command{
		Use:   "account",
		Short: "Manage CalDAV accounts",
	}

	var flagName, flagURL, flagUsername, flagPassword string

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a CalDAV account",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := bufio.NewReader(os.Stdin)

			if flagURL == "" {
				if flagUsername != "" && strings.Contains(flagUsername, "@gmail.com") {
					flagURL = calendar.GoogleCalDAVURL(flagUsername)
					fmt.Fprintf(os.Stderr, "Using Google CalDAV URL: %s\n", flagURL)
				} else {
					fmt.Fprint(os.Stderr, "CalDAV URL: ")
					flagURL, _ = r.ReadString('\n')
					flagURL = strings.TrimSpace(flagURL)
				}
			}
			if flagUsername == "" {
				fmt.Fprint(os.Stderr, "Username (email): ")
				flagUsername, _ = r.ReadString('\n')
				flagUsername = strings.TrimSpace(flagUsername)
			}
			if flagPassword == "" {
				fmt.Fprint(os.Stderr, "Password (App Password for Google): ")
				flagPassword, _ = r.ReadString('\n')
				flagPassword = strings.TrimSpace(flagPassword)
			}
			if flagName == "" {
				flagName = flagUsername
			}

			fmt.Fprintf(os.Stderr, "Verifying CalDAV connection to %s...\n", flagURL)
			mgr := calendar.New()
			if err := mgr.AddAccount(flagName, flagURL, flagUsername, flagPassword); err != nil {
				return err
			}
			fmt.Printf("✓ Added account %s (%s)\n", flagName, flagUsername)
			return nil
		},
	}
	addCmd.Flags().StringVar(&flagName, "name", "", "Display name")
	addCmd.Flags().StringVar(&flagURL, "url", "", "CalDAV URL")
	addCmd.Flags().StringVar(&flagUsername, "username", "", "Username / email")
	addCmd.Flags().StringVar(&flagPassword, "password", "", "Password / App Password")

	acct.AddCommand(addCmd)

	acct.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List configured accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			accounts, err := calendar.ListAccounts()
			if err != nil {
				return err
			}
			if len(accounts) == 0 {
				fmt.Println("No accounts. Run: doot calendar account add")
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
			if err := calendar.RemoveAccount(args[0]); err != nil {
				return err
			}
			fmt.Printf("Removed account %s\n", args[0])
			return nil
		},
	})

	return acct
}

func calendarSetup() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Show CalDAV setup instructions",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(`CalDAV Calendar Setup
=====================

Google Calendar
---------------
1. Enable 2-Factor Authentication on your Google account
2. Go to: https://myaccount.google.com/apppasswords
3. Create an App Password (select "Other" or "Mail")
4. Copy the 16-character password (no spaces needed)
5. Run: doot calendar account add --username you@gmail.com

   The URL will be auto-detected for Gmail addresses:
   https://apidata.google.com/caldav/v2/you@gmail.com/user

Other CalDAV servers (Nextcloud, iCloud, Fastmail, etc.)
---------------------------------------------------------
Run: doot calendar account add --url https://your-server/dav/calendars/user/ \
       --username your@email.com

iCloud:   https://caldav.icloud.com/
Fastmail: https://caldav.fastmail.com/dav/calendars/

`)
		},
	}
}
