package commands

import "github.com/spf13/cobra"

func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "doot",
		Short: "Dotfiles utility — shell helpers and service commands",
	}

	root.AddCommand(
		Serve(),
		Waybar(),
	)

	return root
}
