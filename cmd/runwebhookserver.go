// Package cmd contains commands
/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/clarkezone/previewd/pkg/config"
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
	lrm              *llrm.LocalRepoManager
	jm               *jobmanager.Jobmanager
	enableBranchMode bool
	whl              *webhooklistener.WebhookListener
)

type providers interface {
	initialClone(string, string) error
	initialBuild(string) error
	webhookListen()
	waitForInterupt() error
	needInitialization() bool
}

type xxxProvider struct {
}

func getRunWebhookServerCmd(p providers) *cobra.Command {
	// nolint
	command := &cobra.Command{
		// TODO: update documentation once flags stable
		Use:   "runwebhookserver --targetrepo=<target repo URL> --localdir=<path to local dir>",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

previewd runwebhookserver --targetrepo=http://repo.git --localdir=/tmp/foo
previewd runwebhookserver --targetrepo=test --localdir=/tmp --initialclone=false --initialbuild=false --webhooklisten=false
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
			clarkezoneLog.Infof("previewd version:%s hash:%s\n", config.VersionString, config.VersionHash)
			clarkezoneLog.Successf("runwebhookserver with port: %v, TargetRepo:%v, localdir:%v, initialbranch:%v, namespace:'%v'",
				internal.Port, internal.TargetRepo, internal.LocalDir, internal.InitialBranch, internal.Namespace)
			clarkezoneLog.Successf(" clone on run:%v, build on run:%v, start webhook server:%v, start preview server:%v",
				internal.InitialClone, internal.InitialBuild, internal.WebhookListen, previewserver)

			err := PerformActions(p, internal.TargetRepo, internal.LocalDir,
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
func PerformActions(provider providers, repo string, localRootDir string, initialBranch string,
	namespace string, webhooklisten bool, serve bool, initialbuild bool, initialclone bool) error {
	sourceDir := path.Join(localRootDir, "sourceroot")
	clarkezoneLog.Debugf("PerformActions() with providers:%v, repo:%v, localRootDir:%v, initialBranch:%v,",
		provider, repo, localRootDir, initialBranch)
	clarkezoneLog.Debugf(" namespace:%v, webhooklisten:%v, serve:%v, initialBuild:%v, initialClone:%v",
		namespace, webhooklisten, serve, initialbuild, initialclone)

	fileinfo, res := os.Stat(sourceDir)
	if fileinfo != nil && res == nil {
		err := os.RemoveAll(sourceDir)
		if err != nil {
			return err
		}
	}

	var err error

	// When running unit tests, don't initialize dependencies
	err = intitializeDependencies(provider, webhooklisten, initialbuild, namespace, localRootDir)
	if err != nil {
		return err
	}

	if initialclone {
		err = provider.initialClone(repo, initialBranch)
		if err != nil {
			clarkezoneLog.Debugf("initialClone failed %v", err)
			return err
		}
	}

	if webhooklisten {
		clarkezoneLog.Debugf("PerformActions() start webhookListen on provider")
		provider.webhookListen()
	}

	if initialbuild {
		clarkezoneLog.Debugf("PerformActions() initialBuild with namespace %v", namespace)
		err = provider.initialBuild(namespace)
		if err != nil {
			clarkezoneLog.Debugf("initialbuild failed: %v", err)
			return err
		}
	}

	if webhooklisten {
		clarkezoneLog.Debugf("PerformActions() calling waitforinterupt on provider")
		err = provider.waitForInterupt()
		if err != nil {
			return err
		}
	}

	return nil
}

func intitializeDependencies(provider providers, webhooklisten bool, initialbuild bool,
	namespace string, localRootDir string) error {
	if provider.needInitialization() {
		c, err := getConfig(internal.InitialBuild, internal.WebhookListen)
		if err != nil {
			return err
		}
		if webhooklisten || initialbuild {
			// possible that integration tests has preconfigured job manager
			if jm == nil {
				jm, err = jobmanager.Newjobmanager(c, namespace, true, false)
				if err != nil {
					return err
				}
			}
		}
		lrm, err = llrm.CreateLocalRepoManager(localRootDir, nil, enableBranchMode, jm, namespace)
		if err != nil {
			clarkezoneLog.Debugf("Unable to create localrepomanager via CreateLocalRepoManager")
			return err
		}
		whl = webhooklistener.CreateWebhookListener(lrm)
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

func (xxxProvider) waitForInterupt() error {
	return whl.WaitForInterupt()
}

func (xxxProvider) initialBuild(namespace string) error {
	clarkezoneLog.Debugf("initialbuild() with namespace %v", namespace)
	const rendername = "render"
	const sourcename = "source"
	render, err := jm.KubeSession().FindpvClaimByName(rendername, namespace)
	if err != nil {
		clarkezoneLog.Errorf("initialBuild() can't find pvcalim render %v", err)
		return err
	}
	if render == "" {
		clarkezoneLog.Errorf("initialBuild() render name empty")
		return fmt.Errorf("initialBuild() render name empty")
	}
	source, err := jm.KubeSession().FindpvClaimByName(sourcename, namespace)
	if err != nil {
		clarkezoneLog.Errorf("initialBuild() can't find pvcalim source %v", err)
		return err
	}
	if source == "" {
		clarkezoneLog.Errorf("initialBuild() source name empty")
		return fmt.Errorf("initialBuild() source name empty")
	}
	renderref := jm.KubeSession().CreatePvCMountReference(render, "/site", false)
	srcref := jm.KubeSession().CreatePvCMountReference(source, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	imagePath := internal.GetJekyllImage()
	command, params := internal.GetJekyllCommands()
	clarkezoneLog.Debugf("initialBuild() submitting job namespace:%v, imagePath:%v, command:%v, pararms:%v, refs:%v",
		namespace, imagePath, command, params, refs)
	err = jm.AddJobtoQueue("jekyll-render-container", namespace, imagePath, command,
		params, refs)
	if err != nil {
		clarkezoneLog.Errorf("Failed to create job: %v", err.Error())
	}
	return nil
}

func init() {
	p := &xxxProvider{}
	runwebhookserverCmd := getRunWebhookServerCmd(p)
	rootCmd.AddCommand(runwebhookserverCmd)
}
