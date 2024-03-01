package commands

import "github.com/spf13/cobra"

var (
	serversCmd = &cobra.Command{
		Use:   "servers",
		Short: "Manage Minecraft servers",
		Long:  "Manage Minecraft servers",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
)

func init() {
	serversCmd.AddCommand(serversListCmd)
}
