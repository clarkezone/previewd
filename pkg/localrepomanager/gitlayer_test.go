package localrepomanager

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/clarkezone/previewd/internal"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	"github.com/sirupsen/logrus"
)

// TestMain initizlie all tests
func TestMain(m *testing.M) {
	clarkezoneLog.Init(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func TestAllReadEnvTest(t *testing.T) {
	t.Logf("TestAllReadEnvTest")
	repo, localdr, testbranchswitch, _, _ := internal.Getenv(t)
	if repo == "" || localdr == "" || testbranchswitch == "" {
		t.Fatalf("Test environment variables not configured repo:%v, localdr:%v, testbranchswitch:%v,\n",
			repo, localdr, testbranchswitch)
	}
}

func TestCloneNoAuth(t *testing.T) {
	t.Logf("TestCloneNoAuth")
	//nolint
	reponame, dirName, _, _, _ := internal.Getenv(t)

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
	reponame, dirName, branch, _, _ := internal.Getenv(t)

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

	const expectedcount = 23
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
