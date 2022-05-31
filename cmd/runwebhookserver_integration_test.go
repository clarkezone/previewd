//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package cmd

import (
	"testing"

	"github.com/clarkezone/previewd/pkg/kubelayer"
	corev1 "k8s.io/api/core/v1"
)

func PrepareEnvironment(t *testing.T) {
	ks := getKubeSession(t)
	wait := make(chan bool)
	ks.CreateNamespace("testns", func(ns *corev1.Namespace, rt kubelayer.ResourseStateType) {
		if rt == kubelayer.Create {
			wait <- true
		}
	})
	_, _ = createVolumes(ks, t)

}

func createJobForClone(t *testing.T, ks *kubelayer.KubeSession) {
	const render = "render"
	const source = "source"
	// create job to launch clone only previewd with persistent volumes bound
	renderref := ks.CreatePvCMountReference(render, "/site", false)
	srcref := ks.CreatePvCMountReference(source, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	imagePath := "registry.hub.docker.com/clarkezone/previewd:webhookcmdline"
	cmd := []string{"./previewd"}
	args := []string{"runwebhookserver", "--targetrepo=https://github.com/clarkezone/clarkezone.github.io.git", "--localdir=/src", " --initialclone=false",
		"--initialbuild=true", "--webhooklisten=true", "--loglevel=debug"}
	_, err := ks.CreateJob("populatepv", "testns", imagePath, cmd, args, nil, false, refs)
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
}

func createJobForTestServer(t *testing.T, ks *kubelayer.KubeSession) {
	const render = "render"
	const source = "source"
	// create job to launch clone only previewd with persistent volumes bound
	renderref := ks.CreatePvCMountReference(render, "/site", false)
	srcref := ks.CreatePvCMountReference(source, "/src", true)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	imagePath := "registry.hub.docker.com/clarkezone/previewd:webhookcmdline"
	cmd := []string{"./previewd"}
	args := []string{"testserver"}
	_, err := ks.CreateJob("testserver", "testns", imagePath, cmd, args, nil, false, refs)
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
}

func TestCreateJobForClone(t *testing.T) {
	ks := getKubeSession(t)
	createJobForClone(t, ks)
}

func TestCreateJobTestServer(t *testing.T) {
	ks := getKubeSession(t)
	createJobForTestServer(t, ks)
}

func getKubeSession(t *testing.T) *kubelayer.KubeSession {
	ks, err := kubelayer.Newkubesession(GetTestConfig(t))
	if err != nil {
		t.Fatalf("Unable to create kubesession %v", err)
	}
	return ks
}

func TestSetupEnvironment(t *testing.T) {
	PrepareEnvironment(t)
}

func TestPerformActions(t *testing.T) {
	// 	// TODO: create test namespace
	// 	jm, err := jobmanager.Newjobmanager(GetTestConfig(t), testNamespace, false)
	// 	if err != nil {
	// 		t.Errorf("job manager create failed")
	// 	}
	// 	//	err = jm.KubeSession().CreateNamespace(testNamespace)
	// 	//	if err != nil {
	// 	//		t.Fatalf("unable to create namespace %v", err)
	// 	//	}
	// 	createVolumes(err, jm.KubeSession(), t)
	//
	// 	repo, localdir, _, _, _ := internal.Getenv(t)
	// 	c := GetTestConfig(t)
	// 	err = PerformActions(currentProvider, c, repo, localdir, "main", "testns", false, false, true, true)
	// 	if err != nil {
	// 		t.Fatalf("Performactions failed %v", err)
	// 	}
	//
	// 	// TODO: call post to webhook to trigger job
}

func createVolumes(jm *kubelayer.KubeSession, t *testing.T) (string, string) {
	const rendername = "render"
	const sourcename = "source"

	err := jm.CreatePersistentVolumeClaim(sourcename, testNamespace)
	if err != nil {
		t.Fatalf("unable to create persistent volume claim %v", err)
	}

	err = jm.CreatePersistentVolumeClaim(rendername, testNamespace)
	if err != nil {
		t.Fatalf("unable to create persistent volume claim %v", err)
	}
	return sourcename, rendername
}
