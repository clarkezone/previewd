// Package cmd contains the cli command definitions for previewd:w
package cmd

/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/

import (
	"fmt"
	"net/http"

	"github.com/clarkezone/previewd/cmd/testserver"
	"github.com/spf13/cobra"
)

var bs = testserver.CreateBasicServer()

// testserverCmd represents the testserver command
var testserverCmd = &cobra.Command{
	Use:   "testserver",
	Short: "Starts a test server to test logging and metrics",
	Long: `Starts a listener that will
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		//TODO: flag with default for port

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			message := fmt.Sprintln("Hello World<BR>")
			_, err := w.Write([]byte(message))
			if err != nil {
				panic(err)
			}
		})

		bs.StartListen("")
		//TODO: implement
		return bs.WaitforInterupt()
	},
}

func init() {
	rootCmd.AddCommand(testserverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testserverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testserverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
