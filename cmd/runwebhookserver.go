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

	"github.com/clarkezone/previewd/pkg/jobmanager"
	"github.com/clarkezone/previewd/pkg/kubelayer"
	llrm "github.com/clarkezone/previewd/pkg/localrepomanager"
	"github.com/clarkezone/previewd/pkg/webhooklistener"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/rest"

	"github.com/clarkezone/previewd/internal"
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

		//	if serve || initialbuild || webhooklisten || initialclone {
		//		result := verifyFlags(repo, localRootDir, initialbuild, initialclone)
		//		if result != nil {
		//			return result
		//		}
		//	} else {
		//		return nil
		//	}
		// TODO args from flags
		// err := PerformActions(repo, localRootDir, initalBranchName, incluster, "jekyllpreviewv2", webhooklisten)
		return nil
	},
}

// PerformActions runs the webhook logic
func PerformActions(c *rest.Config, repo string, localRootDir string, initialBranch string,
	preformInCluster bool, namespace string, webhooklisten bool, serve bool, initialbuild bool, initialclone bool) error {
	sourceDir := path.Join(localRootDir, "sourceroot")
	fileinfo, res := os.Stat(sourceDir)
	if fileinfo != nil && res == nil {
		err := os.RemoveAll(sourceDir)
		if err != nil {
			return err
		}
	}

	var err error
	if webhooklisten || initialbuild {
		jm, err = jobmanager.Newjobmanager(c, namespace)
		if err != nil {
			return err
		}
	}
	lrm, err = llrm.CreateLocalRepoManager(localRootDir, nil, enableBranchMode, jm)
	if err != nil {
		clarkezoneLog.Debugf("Unable to create localrepomanager via CreateLocalRepoManager")
		return err
	}
	whl = webhooklistener.CreateWebhookListener(lrm)

	if initialclone {
		err = initialClone(repo, initialBranch)
		if err != nil {
			clarkezoneLog.Debugf("initialClone failed %v", err)
			return err
		}
	}

	if webhooklisten {
		whl.StartListen("")
	}

	if initialbuild {
		err = initialBuild(namespace)
		if err != nil {
			clarkezoneLog.Debugf("initialbuild failed: %v", err)
		}
		return err
	}
	return nil
}

func initialClone(repo string, initialBranch string) error {
	err := lrm.InitialClone(repo, "")
	if err != nil {
		return err
	}

	if initialBranch != "" {
		return lrm.SwitchBranch(initialBranch)
	}
	return nil
}

func initialBuild(namespace string) error {
	notifier := (func(job *batchv1.Job, typee jobmanager.ResourseStateType) {
		log.Printf("Got job in outside world %v", typee)

		if typee == jobmanager.Update && job.Status.Active == 0 && job.Status.Failed > 0 {
			log.Printf("Failed job detected")
		}
	})
	const rendername = "render"
	const sourcename = "source"
	render, err := jm.FindpvClaimByName(rendername, namespace)
	if err != nil {
		clarkezoneLog.Errorf("can't find pvcalim render %v", err)
		return err
	}
	if render == "" {
		clarkezoneLog.Errorf("render name empty")
		return fmt.Errorf("render name empty")
	}
	source, err := jm.FindpvClaimByName(sourcename, namespace)
	if err != nil {
		clarkezoneLog.Errorf("can't find pvcalim source %v", err)
		return err
	}
	if source == "" {
		clarkezoneLog.Errorf("source name empty")
		return fmt.Errorf("source name empty")
	}
	renderref := jm.CreatePvCMountReference(render, "/site", false)
	srcref := jm.CreatePvCMountReference(source, "/src", true)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	var imagePath string
	fmt.Printf("%v", runtime.GOARCH)
	if runtime.GOARCH == "amd64" {
		imagePath = "registry.hub.docker.com/clarkezone/jekyllbuilder:0.0.1.8"
	} else {
		imagePath = "registry.dev.clarkezone.dev/jekyllbuilder:arm"
	}
	command := []string{"sh", "-c", "--"}
	params := []string{"cd source;bundle install;bundle exec jekyll build -d /site JEKYLL_ENV=production"}
	_, err = jm.CreateJob("jekyll-render-container", namespace, imagePath, command,
		params, notifier, false, refs)
	if err != nil {
		log.Printf("Failed to create job: %v\n", err.Error())
	}
	return nil
}

//nolint
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
