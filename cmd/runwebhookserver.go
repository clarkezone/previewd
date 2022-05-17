// Package cmd contains commands
/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/clarkezone/previewd/internal"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

// runwebhookserverCmd represents the runwebhookserver command
var runwebhookserverCmd = &cobra.Command{
	Use:   "runwebhookserver --targetrepo <target repo URL> --localdir <path to local dir>",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
previewd runwebhookserver --targetrepo http://repo.git --localdir /tmp/foo
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := internal.ValidateEnv()
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		clarkezoneLog.Successf("RunE with port: %v TargetRepo:%v localdir:%v",
			internal.Port, internal.TargetRepo, internal.LocalDir)
		clarkezoneLog.Infof("runwebhookserver called")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runwebhookserverCmd)

	runwebhookserverCmd.PersistentFlags().StringVarP(&internal.TargetRepo, internal.TargetRepoVar, "t",
		viper.GetString(internal.TargetRepoVar), "url to target repo to clone")

	runwebhookserverCmd.PersistentFlags().StringVarP(&internal.LocalDir, internal.LocalDirVar, "d",
		viper.GetString(internal.LocalDirVar), "absolute path to local dir to clone into")

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
