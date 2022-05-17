//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package cmd

import (
	"testing"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/jobmanager"
	"k8s.io/client-go/rest"
)

// GetTestConfig returns a local testing config for k8s
func GetTestConfig(t *testing.T) *rest.Config {
	p := internal.GetTestConfigPath(t)
	c, err := jobmanager.GetConfigOutofCluster(p)
	if err != nil {
		t.Fatalf("Couldn't get config %v", err)
	}
	return c
}

func TestFindVolumeSuccess(t *testing.T) {
	// TODO: create test namespace
	repo, localdir, _, _, _ := internal.Getenv(t)
	c := GetTestConfig(t)
	err := PerformActions(c, repo, localdir, "main", false, "testns", false, false, true, true)
	if err != nil {
		t.Fatalf("Performactions failed %v", err)
	}
}
