// Package localrepomanager manages lifecycle of local repos
package localrepomanager

import (
	"os"
	"path"
	"regexp"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/jobmanager"
	"github.com/clarkezone/previewd/pkg/kubelayer"
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

	clarkezoneLog.Debugf("CreateLocalRepoManager rootDir:%v, newBarnch:%v, enableBranchMode:%v, currentBranch:Master, namespace:%v",
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

//nolint
//lint:ignore U1000 called commented out
func (lrm *LocalRepoManager) HandleWebhook(branch string, runjek bool, sendNotify bool) error {
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
	// todo handle branch change
	lrm.startJob()

	if lrm.enableBranchMode && sendNotify && lrm.newBranchObs != nil {
		lrm.newBranchObs.NewBranch(lrm.legalizeBranchName(branch), renderDir)
	}
	return nil
}

func (lrm *LocalRepoManager) startJob() {
	// TODO extract job creation code into internal
	if lrm.jm == nil {
		clarkezoneLog.Infof("Skipping StartJob due to lack of jobmanager instance")
		return
	}
	const rendername = "render"
	const sourcename = "source"
	render, err := lrm.jm.KubeSession().FindpvClaimByName(rendername, lrm.kubenamespace)
	if err != nil {
		clarkezoneLog.Errorf("lrm::startJob () can't find pvcalim render %v", err)
	}
	if render == "" {
		clarkezoneLog.Errorf("ltm::startjob() render name empty")
	}
	source, err := lrm.jm.KubeSession().FindpvClaimByName(sourcename, lrm.kubenamespace)
	if err != nil {
		clarkezoneLog.Errorf("lrm::startjob() can't find pvcalim source %v", err)
	}
	if source == "" {
		clarkezoneLog.Errorf("lrm::startjob() source name empty")
	}
	renderref := lrm.jm.KubeSession().CreatePvCMountReference(render, "/site", false)
	srcref := lrm.jm.KubeSession().CreatePvCMountReference(source, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	imagePath := internal.GetJekyllImage()

	command, params := internal.GetJekyllCommands()
	err = lrm.jm.AddJobtoQueue("jekyll-render-container", lrm.kubenamespace, imagePath, command, params, refs)
	if err != nil {
		clarkezoneLog.Errorf("Failed to create job: %v\n", err.Error())
	}
}
