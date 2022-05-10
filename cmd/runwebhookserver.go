// Package cmd contains commands
/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	"temp.com/JekyllBlogPreview/jobmanager"
	llrm "temp.com/JekyllBlogPreview/localrepomanager"
	"temp.com/JekyllBlogPreview/webhooklistener"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

var (
	lrm              *llrm.LocalRepoManager
	jm               *jobmanager.Jobmanager
	enableBranchMode bool
	whl              *webhooklistener.WebhookListener
)

// runwebhookserverCmd represents the runwebhookserver command
var runwebhookserverCmd = &cobra.Command{
	Use:   "runwebhookserver",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		clarkezoneLog.Infof("runwebhookserver called")
		// TODO args from flags
		// err := PerformActions(repo, localRootDir, initalBranchName, incluster, "jekyllpreviewv2", webhooklisten)
	},
}

func PerformActions(repo string, localRootDir string, initialBranch string, preformInCluster bool, namespace string, webhooklisten bool) error {
	if serve || initialbuild || webhooklisten || initialclone {
		result := verifyFlags(repo, localRootDir, initialbuild, initialclone)
		if result != nil {
			return result
		}
	} else {
		return nil
	}

	sourceDir := path.Join(localRootDir, "sourceroot")
	fileinfo, res := os.Stat(sourceDir)
	if fileinfo != nil && res == nil {
		err := os.RemoveAll(sourceDir)
		if err != nil {
			return err
		}
	}

	var jm *jobmanager.Jobmanager
	var err error
	if webhooklisten || initialbuild {
		jm, err = jobmanager.Newjobmanager(preformInCluster, namespace)
		if err != nil {
			return err
		}
	}
	lrm = llrm.CreateLocalRepoManager(localRootDir, sharemgn, enableBranchMode, jm)
	whl = webhooklistener.CreateWebhookListener(lrm)

	if initialclone {
		err := lrm.InitialClone(repo, "")
		if err != nil {
			return err
		}

		if initialBranch != "" {
			return lrm.SwitchBranch(initialBranch)
		}
	}

	if webhooklisten {
		whl.StartListen("")
	}

	if initialbuild {
		notifier := (func(job *batchv1.Job, typee jobmanager.ResourseStateType) {
			log.Printf("Got job in outside world %v", typee)

			if typee == jobmanager.Update && job.Status.Active == 0 && job.Status.Failed > 0 {
				log.Printf("Failed job detected")
			}
		})
		var imagePath string
		fmt.Printf("%v", runtime.GOARCH)
		if runtime.GOARCH == "amd64" {
			imagePath = "registry.hub.docker.com/clarkezone/jekyllbuilder:0.0.1.8"
		} else {

			imagePath = "registry.dev.clarkezone.dev/jekyllbuilder:arm"
		}
		command := []string{"sh", "-c", "--"}
		params := []string{"cd source;bundle install;bundle exec jekyll build -d /site JEKYLL_ENV=production"}
		_, err = jm.CreateJob("jekyll-render-container", namespace, imagePath, command, params, notifier)
		if err != nil {
			log.Printf("Failed to create job: %v\n", err.Error())
		}

	}
	return nil
}

func verifyFlags(repo string, localRootDir string, build bool, clone bool) error {
	return nil
	//	if clone && repo == "" {
	//		return fmt.Errorf("repo must be provided in %v", reponame)
	//	}
	//
	//	if clone {
	//		if localRootDir == "" {
	//			return fmt.Errorf("localdir be provided in %v", localRootDir)
	//		} else {
	//			fileinfo, res := os.Stat(localRootDir)
	//			if res != nil {
	//				return fmt.Errorf("localdir must exist %v", localRootDir)
	//			}
	//			if !fileinfo.IsDir() {
	//				return fmt.Errorf("localdir must be a directory %v", localRootDir)
	//			}
	//		}
	//	}
	//	if build && !clone {
	//		return fmt.Errorf("cannont request initial build without an initial clone %v", reponame)
	//	}
	//	return nil
}

func init() {
	rootCmd.AddCommand(runwebhookserverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runwebhookserverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runwebhookserverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
