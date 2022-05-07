package localrepomanager

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	"github.com/sirupsen/logrus"
)

const (
	testreponame         = "TEST_GITLAYER_REPO_NOAUTHURL"
	testlocaldirname     = "TEST_GITLAYER_LOCALDIR"
	testbranchswitchname = "TEST_GITLAYER_BRANCHSWITCH"
	testsecurereponame   = "TEST_GITLAYER_SECURE_REPO_NOAUTH"
	//nolint
	testsecureclonepwname = "TEST_GITLAYER_SECURECLONEPWNAME"
)

// configure environment variables by:
// 1. command palette: open settings (json)
// 2. append the following
// "go.testEnvFile": "/home/james/.previewd_test.env",
// 3. contents of file
// TEST_GITLAYER_REPO_NOAUTHURL="https:/"
// TEST_GITLAYER_LOCALDIR=""
// TEST_GITLAYER_BRANCHSWITCH=""
// TEST_GITLAYER_SECURE_REPO_NOAUTH=""
// TEST_GITLAYER_SECURECLONEPW=""
// TEST_GITLAYER_TESTLOCALK8S=""

// TestMain initizlie all tests
func TestMain(m *testing.M) {
	clarkezoneLog.Init(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func TestAllReadEnvTest(t *testing.T) {
	t.Logf("TestAllReadEnvTest")
	repo, localdr, testbranchswitch, _, _ := Getenv(t)
	if repo == "" || localdr == "" || testbranchswitch == "" {
		t.Fatalf("Test environment variables not configured repo:%v, localdr:%v, testbranchswitch:%v,\n",
			repo, localdr, testbranchswitch)
	}
}

func TestCloneNoAuth(t *testing.T) {
	t.Logf("TestCloneNoAuth")
	//nolint
	reponame, dirName, _, _, _ := Getenv(t)

	err := os.RemoveAll(dirName)
	if err != nil {
		t.Fatalf("Dir already exists")
	}

	_, err = clone(reponame, dirName)

	if err != nil {
		t.Fatalf("Clone failed %v", err)
	}

	if _, err := os.Stat(dirName); err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Clone failed %v no files were copied", err)
		}
	}

	infos, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Fatalf("TestCloneNoAuth: clone failed %v", err.Error())
	}

	if len(infos) < 8 {
		log.Fatalf("TestCloneNoAuth: clone failed expected %v, found %v", 9, len(infos))
	}

	err = os.RemoveAll(dirName)
	if err != nil {
		t.Fatalf("Unable to remove dir %v", err)
	}
}

func TestPullBranch(t *testing.T) {
	t.Logf("TestPullBranch")
	reponame, dirName, branch, _, _ := Getenv(t)

	err := os.RemoveAll(dirName)
	if err != nil {
		log.Fatal("TestPullBranch: removeallfailed")
	}

	repo, err := clone(reponame, dirName)
	if err != nil {
		log.Fatal("TestPullBranch: clone failed")
	}

	err = repo.checkout(branch)
	if err != nil {
		log.Fatal("checkout failed")
	}

	err = repo.pull(branch)
	if err != nil {
		log.Fatal("pull failed")
	}

	infos, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Fatal("pull failed")
	}

	const expectedcount = 22
	if len(infos) != expectedcount { // One extra for .git
		log.Fatalf("pull failed file mismatch error expected %v found %v", expectedcount, len(infos))
	}

	err = os.RemoveAll(dirName)
	if err != nil {
		log.Fatal("TestPullBranch: removeallfailed")
	}
}

// func TestCloneAuth(t *testing.T) {
// 	t.Logf("TestCloneAuth")
// 	_, dirname, _, secureproname, pw := Getenv()
// 	// reponame, dirname, branch, pw := Getenv()
//
// 	if pw == "unused" {
// 		return
// 	}
//
// 	err := os.RemoveAll(dirname)
// 	if err != nil {
// 		log.Fatal("TestCloneAuth: removeallfailed")
// 	}
//
// 	_, err = secureClone(secureproname, dirname, pw)
// 	// repo, err := clone(reponame, dirname, "", pw)
// 	if err != nil {
// 		log.Fatal("TestCloneAuth: clone failed")
// 	}
//
// 	// err = repo.checkout(branch)
// 	// if err != nil {
// 	// 	log.Fatal("checkout failed")
// 	// }
//
// 	// err = repo.pull(branch)
// 	// if err != nil {
// 	// 	log.Fatal("pull failed")
// 	// }
//
// 	infos, err := ioutil.ReadDir(dirname)
// 	if err != nil {
// 		log.Fatal("pull failed")
// 	}
//
// 	if len(infos) != 3 { // One extra for .git
// 		log.Fatalf("pull failed file mismatch error")
// 	}
//
// 	err = os.RemoveAll(dirname)
// 	if err != nil {
// 		log.Fatal("TestCloneAuth: removeallfailed")
// 	}
// }

func Getenv(t *testing.T) (string, string, string, string, string) {
	repo := os.Getenv(testreponame)
	localdr := os.Getenv(testlocaldirname)
	testbranchswitch := os.Getenv(testbranchswitchname)
	reposecure := os.Getenv(testsecurereponame)
	secureclonepw := os.Getenv(testsecureclonepwname)
	if repo == "" || localdr == "" || testbranchswitch == "" {
		t.Fatalf("Test environment variables not configured repo:%v, localdr:%v, testbranchswitch:%v,\n",
			repo, localdr, testbranchswitch)
	}
	return repo, localdr, testbranchswitch, reposecure, secureclonepw
}
