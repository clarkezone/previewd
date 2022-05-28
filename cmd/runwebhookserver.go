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
	"k8s.io/client-go/rest"

	"github.com/clarkezone/previewd/internal"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

var (
	lrm                 *llrm.LocalRepoManager
	jm                  *jobmanager.Jobmanager
	enableBranchMode    bool
	whl                 *webhooklistener.WebhookListener
	currentProvider     providers
	runwebhookserverCmd *cobra.Command
)

type providers interface {
	initialClone(string, string) error
	initialBuild(string) error
	webhookListen()
	needInitialization() bool
}

type xxxProvider struct {
}

func getRunWebhookServerCmd(p providers) *cobra.Command {
	currentProvider = p
	command := &cobra.Command{
		// TODO: update documentation once flags stable
		Use:   "runwebhookserver --targetrepo <target repo URL> --localdir <path to local dir>",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
previewd runwebhookserver --targetrepo http://repo.git --localdir /tmp/foo
./bin/previewd runwebhookserver --targetrepo test --localdir /tmp --initialclone=false --initialbuild=false --webhooklisten=false
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := internal.ValidateEnv()
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			previewserver := false

			clarkezoneLog.Successf("runwebhookserver with port: %v, TargetRepo:%v, localdir:%v, initialbranch:%v, namespace:'%v'",
				internal.Port, internal.TargetRepo, internal.LocalDir, internal.InitialBranch, internal.Namespace)
			clarkezoneLog.Successf(" clone on run:%v, build on run:%v, start webhook server:%v, start preview server:%v",
				internal.InitialClone, internal.InitialBuild, internal.WebhookListen, previewserver)

			p := xxxProvider{}

			c, err := getConfig(internal.InitialBuild, internal.WebhookListen)
			if err != nil {
				return err
			}

			err = PerformActions(&p, c, internal.TargetRepo, internal.LocalDir,
				internal.InitialBranch,
				internal.Namespace, internal.WebhookListen, false, internal.InitialBuild, internal.InitialClone)
			return err
		},
	}

	err := setupFlags(command)
	if err != nil {
		panic(err)
	}
	return command
}

func setupFlags(command *cobra.Command) error {
	command.PersistentFlags().StringVarP(&internal.TargetRepo, internal.TargetRepoVar, "t",
		viper.GetString(internal.TargetRepoVar), "url to target repo to clone")

	command.PersistentFlags().StringVarP(&internal.LocalDir, internal.LocalDirVar, "d",
		viper.GetString(internal.LocalDirVar), "absolute path to local dir to clone into")

	command.PersistentFlags().StringVarP(&internal.KubeConfigPath, internal.KubeConfigPathVar, "k",
		viper.GetString(internal.KubeConfigPathVar), "absolute path to a valid kubeconfig file")
	err := viper.BindPFlag(internal.KubeConfigPathVar, command.PersistentFlags().Lookup(internal.KubeConfigPathVar))
	if err != nil {
		return err
	}

	command.PersistentFlags().StringVarP(&internal.Namespace, internal.NamespaceVar, "n",
		viper.GetString(internal.NamespaceVar), "Kube namespace for creating resources")
	err = viper.BindPFlag(internal.NamespaceVar, command.PersistentFlags().Lookup(internal.NamespaceVar))
	if err != nil {
		return err
	}

	command.PersistentFlags().BoolVarP(&internal.InitialClone, internal.InitialCloneVar, "c",
		viper.GetBool(internal.InitialCloneVar), "perform clone at startup")
	err = viper.BindPFlag(internal.InitialCloneVar, command.PersistentFlags().Lookup(internal.InitialCloneVar))
	if err != nil {
		return err
	}
	command.PersistentFlags().BoolVarP(&internal.InitialBuild, internal.InitialBuildVar, "b",
		viper.GetBool(internal.InitialBuildVar), "perform build at startup")
	err = viper.BindPFlag(internal.InitialBuildVar, command.PersistentFlags().Lookup(internal.InitialBuildVar))
	if err != nil {
		return err
	}

	command.PersistentFlags().BoolVarP(&internal.WebhookListen, internal.WebhookListenVar, "w",
		viper.GetBool(internal.WebhookListenVar), "start webhook listener on startup")
	err = viper.BindPFlag(internal.WebhookListenVar, command.PersistentFlags().Lookup(internal.WebhookListenVar))
	if err != nil {
		return err
	}
	return nil
}

func getConfig(ib bool, wl bool) (*rest.Config, error) {
	// if not doing initial build and not webhook,
	// don't get / load a kube config
	if !ib && !wl {
		return nil, nil
	}
	var c *rest.Config
	var err error
	if internal.KubeConfigPath == "" {
		c, err = kubelayer.GetConfigIncluster()
		clarkezoneLog.Successf("launching inside kubernetes cluster with cluster config")
	} else {
		c, err = kubelayer.GetConfigOutofCluster(internal.KubeConfigPath)
		clarkezoneLog.Successf("launching from outside kubernetes cluster with config %v",
			internal.KubeConfigPath)
	}
	return c, err
}

// PerformActions runs the webhook logic
func PerformActions(provider providers, c *rest.Config, repo string, localRootDir string, initialBranch string,
	namespace string, webhooklisten bool, serve bool, initialbuild bool, initialclone bool) error {
	sourceDir := path.Join(localRootDir, "sourceroot")
	fileinfo, res := os.Stat(sourceDir)
	if fileinfo != nil && res == nil {
		err := os.RemoveAll(sourceDir)
		if err != nil {
			return err
		}
	}

	var err error

	// When running unit tests, don't initialize dependencies
	if currentProvider.needInitialization() {
		if webhooklisten || initialbuild {
			jm, err = jobmanager.Newjobmanager(c, namespace, true)
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
	}

	if initialclone {
		err = currentProvider.initialClone(repo, initialBranch)
		if err != nil {
			clarkezoneLog.Debugf("initialClone failed %v", err)
			return err
		}
	}

	if webhooklisten {
		currentProvider.webhookListen()
	}

	if initialbuild {
		err = currentProvider.initialBuild(namespace)
		if err != nil {
			clarkezoneLog.Debugf("initialbuild failed: %v", err)
		}
		return err
	}
	return nil
}

func (xxxProvider) needInitialization() bool {
	return true
}

func (xxxProvider) initialClone(repo string, initialBranch string) error {
	err := lrm.InitialClone(repo, "")
	if err != nil {
		return err
	}

	if initialBranch != "" {
		return lrm.SwitchBranch(initialBranch)
	}
	return nil
}

func (xxxProvider) webhookListen() {
	whl.StartListen("")
}

func (xxxProvider) initialBuild(namespace string) error {
	const rendername = "render"
	const sourcename = "source"
	render, err := jm.KubeSession().FindpvClaimByName(rendername, namespace)
	if err != nil {
		clarkezoneLog.Errorf("can't find pvcalim render %v", err)
		return err
	}
	if render == "" {
		clarkezoneLog.Errorf("render name empty")
		return fmt.Errorf("render name empty")
	}
	source, err := jm.KubeSession().FindpvClaimByName(sourcename, namespace)
	if err != nil {
		clarkezoneLog.Errorf("can't find pvcalim source %v", err)
		return err
	}
	if source == "" {
		clarkezoneLog.Errorf("source name empty")
		return fmt.Errorf("source name empty")
	}
	renderref := jm.KubeSession().CreatePvCMountReference(render, "/site", false)
	srcref := jm.KubeSession().CreatePvCMountReference(source, "/src", true)
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
	err = jm.AddJobtoQueue("jekyll-render-container", namespace, imagePath, command,
		params, refs)
	if err != nil {
		log.Printf("Failed to create job: %v\n", err.Error())
	}
	return nil
}

func init() {
	p := &xxxProvider{}
	runwebhookserverCmd = getRunWebhookServerCmd(p)
	rootCmd.AddCommand(runwebhookserverCmd)
}
