package commands

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

func Waybar() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "waybar",
		Short: "Waybar controls",
	}

	cmd.AddCommand(waybarRestart())

	return cmd
}

func waybarRestart() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Kill and relaunch waybar",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Stopping waybar...")
			stop := exec.Command("killall", "-w", "waybar")
			// killall returns non-zero if no process found — that's fine
			_ = stop.Run()

			time.Sleep(100 * time.Millisecond)

			fmt.Println("Starting waybar...")
			start := exec.Command("waybar")
			start.Stdin = nil
			start.Stdout = nil
			start.Stderr = nil
			if err := start.Start(); err != nil {
				return fmt.Errorf("failed to start waybar: %w", err)
			}

			fmt.Println("Waybar restarted.")
			return nil
		},
	}
}
