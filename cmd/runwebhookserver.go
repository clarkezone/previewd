// Package cmd contains commands
/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

// runwebhookserverCmd represents the runwebhookserver command
var runwebhookserverCmd = &cobra.Command{
	Use:   "runwebhookserver http://repotoclone.git",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
previewd runwebhookserver http://repo.git --localdir /tmp/foo
`,
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		clarkezoneLog.Infof("PreRunE with args: %v", args)
		return nil

	},
	Run: func(cmd *cobra.Command, args []string) {

		clarkezoneLog.Infof("runwebhookserver called")
	},
}

func init() {
	rootCmd.AddCommand(runwebhookserverCmd)

	runwebhookserverCmd.Flags().StringP("localdir", "d", "", "absolute path to local dir to clone into")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:jj
	// runwebhookserverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runwebhookserverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// if serve || initialbuild || webhooklisten || initialclone {
	//		result := verifyFlags(repo, localRootDir, initialbuild, initialclone)
	//		if result != nil {
	//			return result
	//		}
	//	} else {
	//		return nil
	//	}
	// TODO args from flags
	// err := PerformActions(repo, localRootDir, initalBranchName, incluster, "jekyllpreviewv2", webhooklisten)
}
