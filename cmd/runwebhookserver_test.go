//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package cmd

import (
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/jobmanager"
	"k8s.io/client-go/rest"
)

var gitRoot string

func setup() {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic("couldn't read output from git command get gitroot")
	}
	gitRoot = string(output)
	gitRoot = strings.TrimSuffix(gitRoot, "\n")
}

func getTestConfig(t *testing.T) *rest.Config {
	configpath := path.Join(gitRoot, "integration/secrets/k3s-c2.yaml")
	c, err := jobmanager.GetConfigOutofCluster(configpath)
	if err != nil {
		t.Fatalf("Couldn't get config %v", err)
	}
	return c
}

func TestFindVolumeSuccess(t *testing.T) {
	repo, localdir, _, _, _ := internal.Getenv(t)
	c := getTestConfig(t)
	err := PerformActions(c, repo, localdir, "main", false, "testns", false, false, true, true)
	if err != nil {
		t.Fatalf("Performactions failed %v", err)
	}
}
