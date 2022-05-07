package localrepomanager

import (
	"io/ioutil"
	"os"
	"testing"
)

func SkipCI(t *testing.T) {
	if os.Getenv("TEST_JEKPREV_TESTLOCALK8S") == "" {
		t.Skip("Skipping K8slocaltest")
	}
}

func TestSourceDir(t *testing.T) {
	lrm, err := CreateLocalRepoManager("test", nil, true, nil)
	if err != nil {
		t.Fatalf("CreateLRM failed %v", err)
	}

	res := lrm.getSourceDir()

	if res != "test/source" {
		t.Fatalf("Incorrect source dir")
	}

	err = os.RemoveAll("test")
	if err != nil {
		t.Fatalf("unable to remove dirs")
	}
}

func TestCreateLocalRepoManager(t *testing.T) {
	_, err := CreateLocalRepoManager("test", nil, true, nil)
	if err != nil {
		t.Fatalf("create localrepomanager failed")
	}

	_, err = ioutil.ReadDir("test")
	if err != nil {
		t.Fatalf("Directory didn't get created")
	}

	_, err = ioutil.ReadDir("test/source")
	if err != nil {
		t.Fatalf("Directory didn't get created")
	}

	err = os.RemoveAll("test")
	if err != nil {
		t.Fatalf("unable to remove dirs")
	}
}

func TestLegalizeBranchName(t *testing.T) {
	const branchname = "foo"
	lrm, err := CreateLocalRepoManager("test", nil, true, nil)
	if err != nil {
		t.Fatalf("create localrepomanager failed")
	}
	result := lrm.legalizeBranchName(branchname)
	if result != branchname {
		t.Fatalf("result incorrect")
	}

	result = lrm.legalizeBranchName("f-o-o")
	if result != branchname {
		t.Fatalf("result incorrect")
	}

	result = lrm.legalizeBranchName("f*o*o")
	if result != "foo" {
		t.Fatalf("result incorrect")
	}

	err = os.RemoveAll("test")
	if err != nil {
		t.Fatalf("unable to remove dirs")
	}
}

func TestGetCurrentBranchRender(t *testing.T) {
	lrm, err := CreateLocalRepoManager("test", nil, true, nil)
	if err != nil {
		t.Fatalf("create localrepomanager failed")
	}

	dir, err := lrm.getRenderDir()
	if err != nil {
		t.Fatalf("getrenderdir failed")
	}

	if dir != "test/master" {
		t.Fatalf("Wrong name")
	}

	_, err = ioutil.ReadDir("test/master")
	if err != nil {
		t.Fatalf("Directory didn't get created")
	}

	err = os.RemoveAll("test")
	if err != nil {
		t.Fatalf("unable to remove dirs")
	}
}

func TestLRMCheckout(t *testing.T) {
	//nolint
	repo, dirname, _, _, _ := Getenv(t)

	lrm, err := CreateLocalRepoManager(dirname, nil, true, nil)
	if err != nil {
		t.Fatalf("create localrepomanager failed")
	}

	err = lrm.InitialClone(repo, "")
	if err != nil {
		t.Fatalf("error in initial clonse")
	}

	err = os.RemoveAll(dirname)
	if err != nil {
		t.Fatalf("unable to remove dirs")
	}
}

// TODO unregister

// func TestLRMSwitchBranch(t *testing.T) {
// 	_, dirname, branch, secureRepo, pat := getenv()

// 	lrm := CreateLocalRepoManager(dirname)
// 	lrm.initialClone(secureRepo, pat)

// 	lrm.handleWebhook(branch, false, true)

// 	branchDir := lrm.getCurrentBranchRenderDir()

// 	if branchDir != path.Join(dirname, branch) {
// 		t.Fatalf("incorrect new dir")
// 	}

// 	os.RemoveAll(dirname)
// }

//  func TestLRMSwitchBranchBackToMain(t *testing.T) {
//	_, dirname, branch, secureRepo, pat := getenv()
//
//	sharemgn := createShareManager()
//
//	lrm := CreateLocalRepoManager(dirname, sharemgn, true)
//	lrm.initialClone(secureRepo, pat)
//
//	lrm.handleWebhook(branch, false, true)
//
//	branchDir := lrm.getRenderDir()
//
//	if branchDir != path.Join(dirname, branch) {
//		t.Fatalf("incorrect new dir")
//	}
//
//	lrm.handleWebhook("master", false, true)
//
//	os.RemoveAll(dirname)
//}
