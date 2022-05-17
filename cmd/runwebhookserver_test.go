//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package cmd

import (
	"path"
	"testing"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/jobmanager"
	"k8s.io/client-go/rest"
)

// GetTestConfig returns a local testing config for k8s
func GetTestConfig(t *testing.T) *rest.Config {
	p := internal.GetTestConfigPath(t)
	configpath := path.Join(p, "integration/secrets/k3s-c2.yaml")
	c, err := jobmanager.GetConfigOutofCluster(configpath)
	if err != nil {
		t.Fatalf("Couldn't get config %v", err)
	}
	return c
}

func TestFindVolumeSuccess(t *testing.T) {
	repo, localdir, _, _, _ := internal.Getenv(t)
	c := GetTestConfig(t)
	err := PerformActions(c, repo, localdir, "main", false, "testns", false, false, true, true)
	if err != nil {
		t.Fatalf("Performactions failed %v", err)
	}
}
