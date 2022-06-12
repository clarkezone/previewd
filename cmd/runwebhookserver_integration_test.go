//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package cmd

import (
	"log"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/jobmanager"
	"github.com/clarkezone/previewd/pkg/kubelayer"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	corev1 "k8s.io/api/core/v1"
)

const (
	renderPvName      = "render"
	sourcePvName      = "source"
	testNamespace     = "testns"
	previewdImagePath = "registry.hub.docker.com/clarkezone/previewd:0.0.3"
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
	cmd := []string{"./previewd"}
	args := []string{"runwebhookserver", "--targetrepo=https://github.com/clarkezone/selfhostinfrablog.git", "--localdir=/src", " --initialclone=true",
		"--initialbuild=false", "--webhooklisten=false", "--loglevel=debug"}
	_, err := ks.CreateJob("populatepv", testNamespace, previewdImagePath, cmd, args, nil, false, refs)
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

	// create job to render using jekyll docker image from prepared source
	renderref := ks.CreatePvCMountReference(renderPvName, "/site", false)
	srcref := ks.CreatePvCMountReference(sourcePvName, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}

	// cmd := []string{"sh", "-c", "--"}
	// params := []string{"sleep 100000"}
	cmd, params := internal.GetJekyllCommands()
	image := internal.GetJekyllImage()

	outputjob := runTestJod(ks, "jekyllrender", testNamespace,
		image,
		completechannel, deletechannel, t, cmd, params, notifier, refs)

	if outputjob.Status.Succeeded != 1 {
		t.Fatalf("Jobs didn't succeed")
	}
}

func TestCreateJobRenderSimulateK8sDeployment(t *testing.T) {
	ks := getKubeSession(t)
	// create a job that launches previewd in cluster perferming initial build
	renderref := ks.CreatePvCMountReference(renderPvName, "/site", false)
	srcref := ks.CreatePvCMountReference(sourcePvName, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	cmd := []string{"./previewd"}
	args := []string{"runwebhookserver", "--targetrepo=https://github.com/clarkezone/clarkezone.github.io.git",
		"--localdir=/src", " --initialclone=false",
		"--initialbuild=true", "--webhooklisten=true", "--loglevel=debug"}
	_, err := ks.CreateJob("rendertopv", testNamespace, previewdImagePath, cmd, args, nil, false, refs)
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
}

// TOTO this is duplicate code but moving it into internal/testutils creates a cycle
// between jobmanager integration tests and internal/testutils

func newCompletionTrackingJobManager(towrap jobmanager.Jobxxx) *CompletionTrackingJobManager {
	wrapped := &CompletionTrackingJobManager{wrappedJob: towrap}
	wrapped.done = make(chan bool, 10)
	return wrapped
}

// TODO: figure out how to share / remove duplicate with jobmanagerintegrationtest
type CompletionTrackingJobManager struct {
	wrappedJob jobmanager.Jobxxx
	done       chan bool
}

func (o *CompletionTrackingJobManager) CreateJob(name string, namespace string,
	image string, command []string, args []string, notifier kubelayer.JobNotifier,
	autoDelete bool, mountlist []kubelayer.PVClaimMountRef) (*batchv1.Job, error) {
	return o.wrappedJob.CreateJob(name, namespace, image, command,
		args, notifier, autoDelete, mountlist)
}

func (o *CompletionTrackingJobManager) DeleteJob(name string, namespace string) error {
	err := o.wrappedJob.DeleteJob(name, namespace)
	o.done <- true
	return err
}

func (o *CompletionTrackingJobManager) FailedJob(name string, namespace string) {
	o.wrappedJob.FailedJob(name, namespace)
	o.done <- true
}

func (o *CompletionTrackingJobManager) InProgress() bool {
	return o.wrappedJob.InProgress()
}

func (o *CompletionTrackingJobManager) WaitDone(t *testing.T, numjobs int) {
	clarkezoneLog.Debugf("Begin wait done on mockjobmananger")
	for i := 0; i < numjobs; i++ {
		select {
		case <-o.done:
		case <-time.After(20 * time.Second):
			t.Fatalf("No done before 10 second timeout")
		}
	}
	clarkezoneLog.Debugf("End wait done on mockjobmananger")
}

type e2emockprovider struct {
	// mock.Mock
}

func (p *e2emockprovider) initialClone(a string, b string) error {
	// p.Called(a, b)
	return nil
}

func (p *e2emockprovider) initialBuild(a string) error {
	clarkezoneLog.Debugf("== initial build with '%v'", a)
	// p.Called(a)
	return nil
}

func (p *e2emockprovider) webhookListen() {
	clarkezoneLog.Debugf("webhookListen")
	// p.Called()
}

func (p *e2emockprovider) waitForInterupt() error {
	clarkezoneLog.Debugf("waitForInterupt")
	// p.Called()
	return nil
}

func (*e2emockprovider) needInitialization() bool {
	return false
}

func TestFullE2eTestWithWebhook(t *testing.T) {
	// launch previewd out of cluster with pre-cloned source valid for jekyll
	// previewd will create an initial in-cluster renderjob which should succeed
	// test calls webhook which creates a second job to re-render which should succeed
	// requires TestSetupEnvironment()
	// requires TestCreateJobForClone()

	localdir := t.TempDir()

	var cm *CompletionTrackingJobManager
	jm, cm = getCompletionTrackingJobManager(t)

	// TODO wrap xxxProvider to hook job completion and waitForInterrupt
	p := &xxxProvider{}
	cmd := getRunWebhookServerCmd(p)

	// targetrepo and localdir are unused as no initial clone
	// webhook will run job in cluster
	cmd.SetArgs([]string{"--targetrepo=http://foo",
		"--localdir=" + localdir, "--kubeconfigpath=" + internal.GetTestConfigPath(t), "--namespace=testns",
		"--initialclone=false",
		"--initialbuild=true", "--webhooklisten=true"})

	// TODO use goroutine
	// Execute will block until sigterm
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	// wait for initial render
	cm.WaitDone(t, 1)

	// TODO: fire webhook

	// wait for webhook triggered render
	cm.WaitDone(t, 1)

	// TODO: close webhooklisten goroutine via wrapped xxxProvider

}

// TODO this is duplicate code.  Find a way of sharing without cycles between testinternal and jobmanaber
func getCompletionTrackingJobManager(t *testing.T) (*jobmanager.Jobmanager, *CompletionTrackingJobManager) {
	internal.SetupGitRoot()
	path := internal.GetTestConfigPath(t)
	config, err := kubelayer.GetConfigOutofCluster(path)
	if err != nil {
		t.Fatalf("Can't get config %v", err)
	}
	jm, err := jobmanager.Newjobmanager(config, "testns", false, true)
	if err != nil {
		t.Fatalf("Can't create jobmanager %v", err)
	}
	wrappedProvider := newCompletionTrackingJobManager(jm.JobProvider)

	jm.JobProvider = wrappedProvider

	// Watchers must be started after the provider has been wrapped
	jm.StartWatchers(true)
	return jm, wrappedProvider
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
	cmd := []string{"./previewd"}
	args := []string{"runwebhookserver", "--targetrepo=https://github.com/clarkezone/clarkezone.github.io.git", "--localdir=/src", " --initialclone=false",
		"--initialbuild=true", "--webhooklisten=true", "--loglevel=debug"}
	_, err := ks.CreateJob("populatepv", testNamespace, previewdImagePath, cmd, args, nil, false, refs)
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
}

func createJobForTestServerWithMountedVols(t *testing.T, ks *kubelayer.KubeSession) {
	// create job to launch clone only previewd with persistent volumes bound
	renderref := ks.CreatePvCMountReference(renderPvName, "/site", false)
	srcref := ks.CreatePvCMountReference(sourcePvName, "/src", true)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	cmd := []string{"./previewd"}
	args := []string{"testserver"}
	_, err := ks.CreateJob("testserver", testNamespace, previewdImagePath, cmd, args, nil, false, refs)
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
