package commands

import "github.com/spf13/cobra"

var (
	rootCmd = &cobra.Command{
		Use:   "kubecraft",
		Short: "Kubecraft CLI for creating and managing Minecraft servers on Kubernetes",
		Long:  "Kubecraft CLI interacts with the Kubecraft Kubernetes operator to create and manage Minecraft servers on Kubernetes",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(serversCmd)
}
