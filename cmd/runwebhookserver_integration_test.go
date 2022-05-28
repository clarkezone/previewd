//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package cmd

import (
	"testing"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/jobmanager"
	"github.com/clarkezone/previewd/pkg/kubelayer"
	corev1 "k8s.io/api/core/v1"
)

func PrepareEnvironment(t *testing.T) {
	ks, err := kubelayer.Newkubesession(GetTestConfig(t))
	if err != nil {
		t.Fatalf("Unable to create kubesession %v", err)
	}
	wait := make(chan bool)
	ks.CreateNamespace("testns", func(ns *corev1.Namespace, rt kubelayer.ResourseStateType) {
		if rt == kubelayer.Create {
			wait <- true
		}
	})
	source, render := createVolumes(err, ks, t)

	// create job to launch clone only previewd with persistent volumes bound
	renderref := jm.KubeSession().CreatePvCMountReference(render, "/site", false)
	srcref := jm.KubeSession().CreatePvCMountReference(source, "/src", true)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	// TODO: fix image path
	imagePath := "registry.hub.docker.com/clarkezone/jekyllbuilder:0.0.1.8"
	// TODO: verify entrypoint
	cmd := []string{"previewd"}
	// TODO: ensure initialclone implemented
	args := []string{" --initialclone true"}
	ks.CreateJob("populatepv", "testns", imagePath, cmd, args, nil, false, refs)
}

func TestSetupEnvironment(t *testing.T) {
	PrepareEnvironment(t)
}

func TestPerformActions(t *testing.T) {
	// TODO: create test namespace
	jm, err := jobmanager.Newjobmanager(GetTestConfig(t), testNamespace, false)
	if err != nil {
		t.Errorf("job manager create failed")
	}
	//	err = jm.KubeSession().CreateNamespace(testNamespace)
	//	if err != nil {
	//		t.Fatalf("unable to create namespace %v", err)
	//	}
	createVolumes(err, jm.KubeSession(), t)

	repo, localdir, _, _, _ := internal.Getenv(t)
	c := GetTestConfig(t)
	err = PerformActions(currentProvider, c, repo, localdir, "main", "testns", false, false, true, true)
	if err != nil {
		t.Fatalf("Performactions failed %v", err)
	}

	// TODO: call post to webhook to trigger job
}

func createVolumes(err error, jm *kubelayer.KubeSession, t *testing.T) (string, string) {
	const rendername = "render"
	const sourcename = "source"

	err = jm.CreatePersistentVolumeClaim(sourcename, testNamespace)
	if err != nil {
		t.Fatalf("unable to create persistent volume claim %v", err)
	}

	err = jm.CreatePersistentVolumeClaim(rendername, testNamespace)
	if err != nil {
		t.Fatalf("unable to create persistent volume claim %v", err)
	}
	return sourcename, rendername
}
