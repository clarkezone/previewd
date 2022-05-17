package internal

import (
	"os"
	"testing"
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
