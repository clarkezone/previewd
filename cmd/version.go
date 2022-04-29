package cmd

import (
	"fmt"

	"github.com/clarkezone/previewd/pkg/config"
	"github.com/spf13/cobra"
)

// Show current version
var versionCommand = getVersionCommand()

func init() {
	rootCmd.AddCommand(versionCommand)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runwebhookserverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runwebhookserverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show previewd version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "rk version:%s hash:%s\n", config.VersionString, config.VersionHash)
			if err != nil {
				return err
			}
			return nil
		},
	}
}
