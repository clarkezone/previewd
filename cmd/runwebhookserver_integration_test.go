//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package cmd

import (
	"log"
	"testing"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/kubelayer"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	corev1 "k8s.io/api/core/v1"
)

const (
	renderPvName  = "render"
	sourcePvName  = "source"
	testNamespace = "testns"
)

func TestSetupEnvironment(t *testing.T) {
	prepareEnvironment(t)
}

func TestCreateJobForClone(t *testing.T) {
	ks := getKubeSession(t)

	// create job to launch clone only previewd with persistent volumes bound
	renderref := ks.CreatePvCMountReference(renderPvName, "/site", false)
	srcref := ks.CreatePvCMountReference(sourcePvName, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	imagePath := "registry.hub.docker.com/clarkezone/previewd:0.0.3"
	cmd := []string{"./previewd"}
	args := []string{"runwebhookserver", "--targetrepo=https://github.com/clarkezone/selfhostinfrablog.git", "--localdir=/src", " --initialclone=true",
		"--initialbuild=false", "--webhooklisten=false", "--loglevel=debug"}
	_, err := ks.CreateJob("populatepv", testNamespace, imagePath, cmd, args, nil, false, refs)
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
	// TODO wait for job to complete
}

func TestCreateJobTestServerMountVols(t *testing.T) {
	ks := getKubeSession(t)
	createJobForTestServerWithMountedVols(t, ks)
}

func TestCreateJobUsinsingPreparedJekyll(t *testing.T) {
	ks := getKubeSession(t)
	completechannel, deletechannel, notifier := getNotifier()

	// create job to launch clone only previewd with persistent volumes bound
	renderref := ks.CreatePvCMountReference(renderPvName, "/site", false)
	srcref := ks.CreatePvCMountReference(sourcePvName, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}

	// cmd := []string{"sh", "-c", "--"}
	// params := []string{"sleep 100000"}
	cmd, params := getJekyllCommands()
	image := getJekyllImage()

	outputjob := runTestJod(ks, "jekyllrender", testNamespace,
		image,
		completechannel, deletechannel, t, cmd, params, notifier, refs)

	if outputjob.Status.Succeeded != 1 {
		t.Fatalf("Jobs didn't succeed")
	}
}

func TestCreateJobRenderSimulateK8sDeployment(t *testing.T) {
	ks := getKubeSession(t)
	// create job to launch clone only previewd with persistent volumes bound
	renderref := ks.CreatePvCMountReference(renderPvName, "/site", false)
	srcref := ks.CreatePvCMountReference(sourcePvName, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	imagePath := "registry.hub.docker.com/clarkezone/previewd:0.0.3"
	cmd := []string{"./previewd"}
	args := []string{"runwebhookserver", "--targetrepo=https://github.com/clarkezone/clarkezone.github.io.git",
		"--localdir=/src", " --initialclone=false",
		"--initialbuild=true", "--webhooklisten=true", "--loglevel=debug"}
	_, err := ks.CreateJob("rendertopv", testNamespace, imagePath, cmd, args, nil, false, refs)
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
}

func TestRunWebhookServercmd(t *testing.T) {
	// ? can this work because stuff isn't cloned
	localdir := t.TempDir()
	p := &xxxProvider{}
	cmd := getRunWebhookServerCmd(p)

	cmd.SetArgs([]string{"--targetrepo", "http://foo",
		"--localdir", localdir, "--kubeconfigpath", internal.GetTestConfigPath(t)})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

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

func ToDoFullE2eTestWithWebhook() {
	// pass args into cmd
	// TODO
}

func prepareEnvironment(t *testing.T) {
	ks := getKubeSession(t)
	wait := make(chan bool)
	ks.CreateNamespace(testNamespace, func(ns *corev1.Namespace, rt kubelayer.ResourseStateType) {
		if rt == kubelayer.Create {
			wait <- true
		}
	})
	_, _ = createVolumes(ks, t)
}

func createJobForClone(t *testing.T, ks *kubelayer.KubeSession) {
	// create job to launch clone only previewd with persistent volumes bound
	renderref := ks.CreatePvCMountReference(renderPvName, "/site", false)
	srcref := ks.CreatePvCMountReference(sourcePvName, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	imagePath := "registry.hub.docker.com/clarkezone/previewd:0.0.3"
	cmd := []string{"./previewd"}
	args := []string{"runwebhookserver", "--targetrepo=https://github.com/clarkezone/clarkezone.github.io.git", "--localdir=/src", " --initialclone=false",
		"--initialbuild=true", "--webhooklisten=true", "--loglevel=debug"}
	_, err := ks.CreateJob("populatepv", testNamespace, imagePath, cmd, args, nil, false, refs)
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
}

func createJobForTestServerWithMountedVols(t *testing.T, ks *kubelayer.KubeSession) {
	// create job to launch clone only previewd with persistent volumes bound
	renderref := ks.CreatePvCMountReference(renderPvName, "/site", false)
	srcref := ks.CreatePvCMountReference(sourcePvName, "/src", true)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	imagePath := "registry.hub.docker.com/clarkezone/previewd:0.0.3"
	cmd := []string{"./previewd"}
	args := []string{"testserver"}
	_, err := ks.CreateJob("testserver", testNamespace, imagePath, cmd, args, nil, false, refs)
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
}

// TODO: share with unit test
func runTestJod(ks *kubelayer.KubeSession, jobName string, testNamespace string, imageUrl string, completechannel chan batchv1.Job, deletechannel chan batchv1.Job,
	t *testing.T, command []string, args []string, notifier func(*batchv1.Job, kubelayer.ResourseStateType),
	mountlist []kubelayer.PVClaimMountRef) batchv1.Job {
	defer ks.Close()
	_, err := ks.CreateJob(jobName, testNamespace, imageUrl,
		command, args, notifier, false, mountlist)
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	outputjob := <-completechannel

	log.Println("Completed; attempting delete")
	err = ks.DeleteJob("alpinetest", testNamespace)
	if err != nil {
		t.Fatalf("Unable to delete job %v", err)
	}
	log.Println(("Deleted."))
	<-deletechannel

	return outputjob
}

func getNotifier() (chan batchv1.Job, chan batchv1.Job, func(job *batchv1.Job, typee kubelayer.ResourseStateType)) {
	completechannel := make(chan batchv1.Job)
	deletechannel := make(chan batchv1.Job)
	notifier := (func(job *batchv1.Job, typee kubelayer.ResourseStateType) {
		clarkezoneLog.Debugf("Got job in outside world %v", typee)

		if completechannel != nil && typee == kubelayer.Update && job.Status.Failed > 0 {
			clarkezoneLog.Debugf("Job failed")
			completechannel <- *job
			close(completechannel)
			completechannel = nil // avoid double close
		}

		if completechannel != nil && typee == kubelayer.Update && job.Status.Succeeded > 0 {
			clarkezoneLog.Debugf("Job succeeded")
			completechannel <- *job
			close(completechannel)
			completechannel = nil // avoid double close
		}

		if typee == kubelayer.Delete && deletechannel != nil {
			log.Printf("Deleted")
			close(deletechannel)
			deletechannel = nil
		}
	})
	return completechannel, deletechannel, notifier
}

func getKubeSession(t *testing.T) *kubelayer.KubeSession {
	ks, err := kubelayer.Newkubesession(GetTestConfig(t))
	if err != nil {
		t.Fatalf("Unable to create kubesession %v", err)
	}
	return ks
}

func createVolumes(jm *kubelayer.KubeSession, t *testing.T) (string, string) {
	err := jm.CreatePersistentVolumeClaim(sourcePvName, testNamespace)
	if err != nil {
		t.Fatalf("unable to create persistent volume claim %v", err)
	}

	err = jm.CreatePersistentVolumeClaim(renderPvName, testNamespace)
	if err != nil {
		t.Fatalf("unable to create persistent volume claim %v", err)
	}
	return sourcePvName, renderPvName
}
