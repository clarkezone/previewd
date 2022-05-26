//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package cmd

import (
	"testing"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/kubelayer"
	"k8s.io/client-go/rest"
)

const (
	testNamespace = "testns"
)

// GetTestConfig returns a local testing config for k8s
func GetTestConfig(t *testing.T) *rest.Config {
	p := internal.GetTestConfigPath(t)
	c, err := kubelayer.GetConfigOutofCluster(p)
	if err != nil {
		t.Fatalf("Couldn't get config %v", err)
	}
	return c
}

func TestPerformActions(t *testing.T) {
	//	// TODO: create test namespace
	//	jm, err := jobmanager.Newjobmanager(GetTestConfig(t), testNamespace)
	//	if err != nil {
	//		t.Errorf("job manager create failed")
	//	}
	//	const rendername = "render"
	//	const sourcename = "source"
	//
	//	err = jm.CreateNamespace(testNamespace)
	//	if err != nil {
	//		t.Fatalf("unable to create namespace %v", err)
	//	}
	//
	//	err = jm.CreatePersistentVolumeClaim(sourcename, testNamespace)
	//	if err != nil {
	//		t.Fatalf("unable to create persistent volume claim %v", err)
	//	}
	//
	//	err = jm.CreatePersistentVolumeClaim(rendername, testNamespace)
	//	if err != nil {
	//		t.Fatalf("unable to create persistent volume claim %v", err)
	//	}
	//
	//	repo, localdir, _, _, _ := internal.Getenv(t)
	//	c := GetTestConfig(t)
	//	err = PerformActions(c, repo, localdir, "main", false, "testns", false, false, true, true)
	//	if err != nil {
	//		t.Fatalf("Performactions failed %v", err)
	//	}
}
