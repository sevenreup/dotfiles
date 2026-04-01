package commands

import (
	"doot/src/internal/server"
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

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig

			srv.Stop()
			return nil
		},
	}
}
