// Package localrepomanager manages lifecycle of local repos
package localrepomanager

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"

	"github.com/clarkezone/previewd/pkg/jobmanager"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	batchv1 "k8s.io/api/batch/v1"
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
}

// CreateLocalRepoManager is a factory method for creating a new LRM instance
func CreateLocalRepoManager(rootDir string,
	newBranch newBranchHandler, enableBranchMode bool,
	jm *jobmanager.Jobmanager) (*LocalRepoManager, error) {
	var lrm = &LocalRepoManager{currentBranch: "master", localRootDir: rootDir}
	lrm.newBranchObs = newBranch
	lrm.enableBranchMode = enableBranchMode
	lrm.jm = jm
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
	clarkezoneLog.Infof("Initial clone for\n repo: %v\n local dir:%v", repo, lrm.repoSourceDir)
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
	if branch != lrm.currentBranch {
		clarkezoneLog.Infof("Fetching\n")

		err := lrm.repo.checkout(branch)
		if err != nil {
			clarkezoneLog.Errorf("LocalRepoManager::Switchbranch %v", err)
			return err
		}

		lrm.currentBranch = branch
	}

	err := lrm.repo.pull(branch)
	if err != nil {
		clarkezoneLog.Errorf("LocalRepoManager::SwitchBranch %v", err)
		return err
	}
	return nil
}

//nolint
//lint:ignore U1000 called commented out
func (lrm *LocalRepoManager) HandleWebhook(branch string, runjek bool, sendNotify bool) error {
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

// nolint
func (lrm *LocalRepoManager) startJob() {
	if lrm.jm == nil {
		clarkezoneLog.Infof("Skipping StartJob due to lack of jobmanager instance")
		return
	}
	namespace := "jekyllpreviewv2"
	notifier := (func(job *batchv1.Job, typee jobmanager.ResourseStateType) {
		clarkezoneLog.Debugf("Got job in outside world %v", typee)

		if typee == jobmanager.Update && job.Status.Active == 0 && job.Status.Failed > 0 {
			clarkezoneLog.Debugf("Failed job detected")
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
	log.Fatalf("fix this")
	_, err := lrm.jm.CreateJob("jekyll-render-container", namespace, imagePath, command, params, notifier, true, nil)
	if err != nil {
		clarkezoneLog.Errorf("Failed to create job: %v\n", err.Error())
	}
}
