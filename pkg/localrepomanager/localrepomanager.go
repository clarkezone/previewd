// Package localrepomanager manages lifecycle of local repos
package localrepomanager

import (
	"os"
	"path"
	"regexp"

	"github.com/clarkezone/previewd/pkg/jobmanager"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	"github.com/go-git/go-git/v5"
)

type newBranchHandler interface {
	NewBranch(branch string, dir string)
}

// LocalRepoManager is a type for managing local git repos
type LocalRepoManager struct {
	currentBranch    string
	repoSourceDir    string
	localRootDir     string
	repo             *gitlayer
	newBranchObs     newBranchHandler
	enableBranchMode bool
	jm               *jobmanager.Jobmanager
	kubenamespace    string
}

// CreateLocalRepoManager is a factory method for creating a new LRM instance
func CreateLocalRepoManager(rootDir string,
	newBranch newBranchHandler, enableBranchMode bool,
	jm *jobmanager.Jobmanager, namespace string) (*LocalRepoManager, error) {
	clarkezoneLog.Debugf("CreateLocalRepoManager rootDir:%v, newBarnch:%v, enableBranchMode:%v,"+
		" currentBranch:Master, namespace:%v",
		rootDir, newBranch, enableBranchMode, namespace)
	var lrm = &LocalRepoManager{currentBranch: "master", localRootDir: rootDir}
	lrm.newBranchObs = newBranch
	lrm.enableBranchMode = enableBranchMode
	lrm.jm = jm
	lrm.kubenamespace = namespace
	// TODO: replace with an error check for missing dir
	//nolint
	os.RemoveAll(rootDir) // ignore error since it may not exist
	dir, err := lrm.ensureDir("source")
	if err != nil {
		return nil, err
	}
	lrm.repoSourceDir = dir
	return lrm, nil
}

func (lrm *LocalRepoManager) ensureDir(subDir string) (string, error) {
	var currentPath = path.Join(lrm.localRootDir, subDir)
	var _, err = os.Stat(currentPath)
	if err != nil {
		err = os.MkdirAll(currentPath, os.ModePerm)
		if err != nil {
			clarkezoneLog.Debugf("Couldn't create sourceDir: %v", err.Error())
			return "", err
		}
	}

	return currentPath, nil
}

func (lrm *LocalRepoManager) getSourceDir() string {
	return lrm.repoSourceDir
}

func (lrm *LocalRepoManager) getRenderDir() (string, error) {
	if lrm.enableBranchMode {
		branchName := lrm.legalizeBranchName(lrm.currentBranch)
		return lrm.ensureDir(branchName)
	}
	return lrm.ensureDir("output")
}

func (lrm *LocalRepoManager) legalizeBranchName(name string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(name, "")
}

// InitialClone performs clone on given repo
func (lrm *LocalRepoManager) InitialClone(repo string, repopat string) error {
	//TODO: this function should ensure branch name is correct
	clarkezoneLog.Debugf("Initial clone for\n repo: %v\n local dir:%v", repo, lrm.repoSourceDir)
	if repopat != "" {
		clarkezoneLog.Debugf(" with Personal Access Token.\n")
	} else {
		clarkezoneLog.Debugf(" with no authentication.\n")
	}

	re, err := clone(repo, lrm.repoSourceDir)
	if err != nil {
		clarkezoneLog.Errorf("EXITING: Fatal Error in initial clone: %v\n", err.Error())
		os.Exit(1)
	}
	lrm.repo = re
	clarkezoneLog.Infof("Clone Done.")
	return err
}

// SwitchBranch changes to a new branch on current repo
func (lrm *LocalRepoManager) SwitchBranch(branch string) error {
	clarkezoneLog.Debugf("SwitchingBranch: resetting with hard")
	re := git.ResetOptions{Mode: git.HardReset}
	err := lrm.repo.wt.Reset(&re)
	if err != nil {
		clarkezoneLog.Errorf("LocalRepoManager::SwitchBranch reset failed with %v", err)
		return err
	}

	if branch != lrm.currentBranch {
		clarkezoneLog.Debugf("Switching branch befween current %v and %v", lrm.currentBranch, branch)

		err := lrm.repo.checkout(branch)
		if err != nil {
			clarkezoneLog.Errorf("LocalRepoManager::Switchbranch checkout failed %v", err)
			return err
		}

		lrm.currentBranch = branch
	}

	err = lrm.repo.pull(branch)
	if err != nil {
		clarkezoneLog.Errorf("LocalRepoManager::SwitchBranch pull failed for %v with %v", branch, err)
		return err
	}
	return nil
}

// HandleWebhook called by webhook machinery to trigger new job
func (lrm *LocalRepoManager) HandleWebhook(branch string, sendNotify bool) error {
	clarkezoneLog.Debugf("LocalRepoManager::HandleWebhook branch: %v", branch)
	err := lrm.SwitchBranch(branch)
	if err != nil {
		clarkezoneLog.Errorf("LocalRepoManager::HandleWebhook %v", err)
		return err
	}

	renderDir, err := lrm.getRenderDir()
	if err != nil {
		clarkezoneLog.Errorf("LocalRepoManager::HandleWebhook %v", err)
		return err
	}

	if lrm.jm == nil {
		clarkezoneLog.Infof("Skipping StartJob due to lack of jobmanager instance")
	} else {
		err = jobmanager.CreateJekyllJob(lrm.kubenamespace, lrm.jm.KubeSession(), lrm.jm)
	}

	if lrm.enableBranchMode && sendNotify && lrm.newBranchObs != nil {
		lrm.newBranchObs.NewBranch(lrm.legalizeBranchName(branch), renderDir)
	}
	return err
}
