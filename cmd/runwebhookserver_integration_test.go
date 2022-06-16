//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package cmd

import (
	"log"
	"strings"
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
	// TODO: wait
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
		"--initialbuild=false", "--webhooklisten=true", "--loglevel=debug"}
	_, err := ks.CreateJob("rendertopv", testNamespace, previewdImagePath, cmd, args, nil, false, refs)
	if err != nil {
		t.Fatalf("create job failed: %v", err)
	}
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

	p := &xxxProvider{}
	wp := newE2mockprovider(p)
	cmd := getRunWebhookServerCmd(wp)

	// targetrepo and localdir are unused as no initial clone
	// webhook will run job in cluster
	// NOTE: no intialclone blows things up due to nil repomanager
	cmd.SetArgs([]string{"--targetrepo=https://github.com/clarkezone/selfhostinfrablog.git",
		"--localdir=" + localdir, "--kubeconfigpath=" + internal.GetTestConfigPath(t), "--namespace=testns",
		"--initialclone=true",
		"--initialbuild=true", "--webhooklisten=true"})

	waitExit := make(chan error)
	go func() {
		err := cmd.Execute()
		waitExit <- err
		close(waitExit)
	}()

	cm.WaitDone(t, 3)

	err := lrm.HandleWebhook("main", true, false)
	if err != nil {
		t.Fatalf("body read failed: %v", err)
	}

	// wait for webhook triggered render
	cm.WaitDone(t, 3)

	wp.signalDone()
	exitError := <-waitExit
	if exitError != nil {
		t.Fatalf(exitError.Error())
	}

	// TODO: use mock to verify calls to completiontrackingjobmanager
}

// TOTO this is duplicate code but moving it into internal/testutils creates a cycle
// between jobmanager integration tests and internal/testutils

type e2emockprovider struct {
	// mock.Mock
	doneChan        chan struct{}
	wrappedProvider providers
}

func newE2mockprovider(p providers) *e2emockprovider {
	provider := e2emockprovider{}
	provider.doneChan = make(chan struct{})
	provider.wrappedProvider = p
	return &provider
}

func (p *e2emockprovider) initialClone(a string, b string) error {
	// p.Called(a, b)
	clarkezoneLog.Debugf("initialClone with %v and %v", a, b)
	return p.wrappedProvider.initialClone(a, b)
}

func (p *e2emockprovider) initialBuild(a string) error {
	clarkezoneLog.Debugf("== initial build with '%v'", a)
	// p.Called(a)
	return p.wrappedProvider.initialBuild(a)
}

func (p *e2emockprovider) webhookListen() {
	clarkezoneLog.Debugf("webhookListen")
	p.wrappedProvider.webhookListen()
}

func (p *e2emockprovider) waitForInterupt() error {
	clarkezoneLog.Debugf("waitForInterupt")
	// p.Called()
	<-p.doneChan
	return nil
}

func (p *e2emockprovider) needInitialization() bool {
	clarkezoneLog.Debugf("needInitialization")
	return p.wrappedProvider.needInitialization()
}

func (p *e2emockprovider) signalDone() {
	close(p.doneChan)
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
	wrappedProvider := newCompletionTrackingJobManager(jm.JobProvider.(jobmanager.Jobxxx), 60)

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

//nolint
func GetBody() *strings.Reader {
	//nolint
	body := `{
		"ref": "refs/heads/master",
		"before": "cfbda0818c286970d373bab5e700599e572d3c40",
		"after": "e2cdd9288b113800293027bc5dae2c7d47b36189",
		"compare_url": "http://gitea.homelab.clarkezone.dev:3000/clarkezone/testfoobar2/compare/cfbda0818c286970d373bab5e700599e572d3c40...e2cdd9288b113800293027bc5dae2c7d47b36189",
		"commits": [
		  {
			"id": "e2cdd9288b113800293027bc5dae2c7d47b36189",
			"message": "Update 'test.txt'\n",
			"url": "http://gitea.homelab.clarkezone.dev:3000/clarkezone/testfoobar2/commit/e2cdd9288b113800293027bc5dae2c7d47b36189",
			"author": {
			  "name": "clarkezone",
			  "email": "james@clarkezone.io",
			  "username": "clarkezone"
			},
			"committer": {
			  "name": "clarkezone",
			  "email": "james@clarkezone.io",
			  "username": "clarkezone"
			},
			"verification": null,
			"timestamp": "2022-04-10T09:35:51Z",
			"added": [],
			"removed": [],
			"modified": [
			  "test.txt"
			]
		  }
		],
		"head_commit": {
		  "id": "e2cdd9288b113800293027bc5dae2c7d47b36189",
		  "message": "Update 'test.txt'\n",
		  "url": "http://gitea.homelab.clarkezone.dev:3000/clarkezone/testfoobar2/commit/e2cdd9288b113800293027bc5dae2c7d47b36189",
		  "author": {
			"name": "clarkezone",
			"email": "james@clarkezone.io",
			"username": "clarkezone"
		  },
		  "committer": {
			"name": "clarkezone",
			"email": "james@clarkezone.io",
			"username": "clarkezone"
		  },
		  "verification": null,
		  "timestamp": "2022-04-10T09:35:51Z",
		  "added": [],
		  "removed": [],
		  "modified": [
			"test.txt"
		  ]
		},
		"repository": {
		  "id": 1,
		  "owner": {"id":1,"login":"clarkezone","full_name":"","email":"james@clarkezone.io","avatar_url":"http://gitea.homelab.clarkezone.dev:3000/user/avatar/clarkezone/-1","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2021-11-21T18:43:19Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"clarkezone"},
		  "name": "testfoobar2",
		  "full_name": "clarkezone/testfoobar2",
		  "description": "",
		  "empty": false,
		  "private": false,
		  "fork": false,
		  "template": false,
		  "parent": null,
		  "mirror": false,
		  "size": 21,
		  "html_url": "http://gitea.homelab.clarkezone.dev:3000/clarkezone/testfoobar2",
		  "ssh_url": "ssh://git@gitea.homelab.clarkezone.dev:2222/clarkezone/testfoobar2.git",
		  "clone_url": "http://gitea.homelab.clarkezone.dev:3000/clarkezone/testfoobar2.git",
		  "original_url": "",
		  "website": "",
		  "stars_count": 0,
		  "forks_count": 0,
		  "watchers_count": 1,
		  "open_issues_count": 0,
		  "open_pr_counter": 0,
		  "release_counter": 0,
		  "default_branch": "master",
		  "archived": false,
		  "created_at": "2021-11-21T18:50:53Z",
		  "updated_at": "2022-04-10T09:32:02Z",
		  "permissions": {
			"admin": true,
			"push": true,
			"pull": true
		  },
		  "has_issues": true,
		  "internal_tracker": {
			"enable_time_tracker": true,
			"allow_only_contributors_to_track_time": true,
			"enable_issue_dependencies": true
		  },
		  "has_wiki": true,
		  "has_pull_requests": true,
		  "has_projects": true,
		  "ignore_whitespace_conflicts": false,
		  "allow_merge_commits": true,
		  "allow_rebase": true,
		  "allow_rebase_explicit": true,
		  "allow_squash_merge": true,
		  "default_merge_style": "merge",
		  "avatar_url": "",
		  "internal": false,
		  "mirror_interval": ""
		},
		"pusher": {"id":1,"login":"clarkezone","full_name":"","email":"james@clarkezone.io","avatar_url":"http://gitea.homelab.clarkezone.dev:3000/user/avatar/clarkezone/-1","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2021-11-21T18:43:19Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"clarkezone"},
		"sender": {"id":1,"login":"clarkezone","full_name":"","email":"james@clarkezone.io","avatar_url":"http://gitea.homelab.clarkezone.dev:3000/user/avatar/clarkezone/-1","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2021-11-21T18:43:19Z","restricted":false,"active":false,"prohibit_login":false,"location":"","website":"","description":"","visibility":"public","followers_count":0,"following_count":0,"starred_repos_count":0,"username":"clarkezone"}
	  }
	`
	reader := strings.NewReader(body)
	return reader
}

type CompletionTrackingJobManager struct {
	wrappedJob      jobmanager.Jobxxx
	done            chan bool
	timeoutinterval int
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
			clarkezoneLog.Debugf("WaitDone: o done received")
		case <-time.After(time.Duration(o.timeoutinterval) * time.Second):
			t.Fatalf("No done before %v second timeout", o.timeoutinterval)
		}
	}
	clarkezoneLog.Debugf("End wait done on mockjobmananger")
}

func newCompletionTrackingJobManager(towrap jobmanager.Jobxxx, timeout int) *CompletionTrackingJobManager {
	wrapped := &CompletionTrackingJobManager{wrappedJob: towrap}
	wrapped.timeoutinterval = timeout
	wrapped.done = make(chan bool)
	return wrapped
}
