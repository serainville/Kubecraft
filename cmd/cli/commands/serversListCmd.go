package commands

import "github.com/spf13/cobra"

var (
	serversListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Minecraft servers",
		Long:  "List Minecraft servers",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
)
