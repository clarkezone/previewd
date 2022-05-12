//go:build integration
// +build integration

// to setup vscode for debugging integration tests see:
// https://www.ryanchapin.com/configuring-vscode-to-use-build-tags-in-golang-to-separate-integration-and-unit-test-code/

package jobmanager

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/rest"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	"github.com/sirupsen/logrus"
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

func SkipCI(t *testing.T) {
	if os.Getenv("TEST_JEKPREV_TESTLOCALK8S") == "" {
		t.Skip("Skipping K8slocaltest")
	}
}

// TestMain initizlie all tests
func TestMain(m *testing.M) {
	clarkezoneLog.Init(logrus.DebugLevel)
	setup()
	code := m.Run()
	os.Exit(code)
}

func GetJobManager(t *testing.T) (*Jobmanager, string) {
	c := getTestConfig(t)
	const ns = "testns"
	jm, err := Newjobmanager(c, ns)
	if err != nil {
		t.Errorf("job manager create failed")
	}
	return jm, ns
}

func RunTestJob(jm *Jobmanager, ns string, completechannel chan batchv1.Job, deletechannel chan batchv1.Job,
	t *testing.T, command []string, notifier func(*batchv1.Job, ResourseStateType), mountlist []PVClaimMountRef) batchv1.Job {
	// SkipCI(t)
	defer jm.Close()

	_, err := jm.CreateJob("alpinetest", "testns", "alpine", command, nil, notifier, false)
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	outputjob := <-completechannel

	log.Println("Completed; attempting delete")
	err = jm.DeleteJob("alpinetest", ns)
	if err != nil {
		t.Fatalf("Unable to delete job %v", err)
	}
	log.Println(("Deleted."))
	<-deletechannel

	return outputjob
}

func getNotifier() (chan batchv1.Job, chan batchv1.Job, func(job *batchv1.Job, typee ResourseStateType)) {
	completechannel := make(chan batchv1.Job)
	deletechannel := make(chan batchv1.Job)
	notifier := (func(job *batchv1.Job, typee ResourseStateType) {
		clarkezoneLog.Debugf("Got job in outside world %v", typee)

		if completechannel != nil && typee == Update && job.Status.Failed > 0 {
			clarkezoneLog.Debugf("Job failed")
			completechannel <- *job
			close(completechannel)
			completechannel = nil // avoid double close
		}

		if completechannel != nil && typee == Update && job.Status.Succeeded > 0 {
			clarkezoneLog.Debugf("Job succeeded")
			completechannel <- *job
			close(completechannel)
			completechannel = nil // avoid double close
		}

		if typee == Delete && deletechannel != nil {
			log.Printf("Deleted")
			close(deletechannel)
			deletechannel = nil
		}
	})
	return completechannel, deletechannel, notifier
}

// TODO test autodelete

// TODO test find volumes

// TODO test mount volumes

func getTestConfig(t *testing.T) *rest.Config {
	configpath := path.Join(gitRoot, "integration/secrets/k3s-c2.yaml")
	c, err := GetConfigOutofCluster(configpath)
	if err != nil {
		t.Fatalf("Couldn't get config %v", err)
	}
	return c
}

func TestCreateAndSucceed(t *testing.T) {
	t.Logf("TestCreateAndSucceed")
	// SkipCI(t)
	completechannel, deletechannel, notifier := getNotifier()
	jm, ns := GetJobManager(t)
	outputjob := RunTestJob(jm, ns, completechannel, deletechannel, t, nil, notifier, nil)
	if outputjob.Status.Succeeded != 1 {
		t.Fatalf("Jobs didn't succeed")
	}
}

func TestCreateAndErrorWork(t *testing.T) {
	t.Logf("TestCreateAndSucceed")
	// SkipCI(t)
	completechannel, deletechannel, notifier := getNotifier()
	command := []string{"error"}
	jm, ns := GetJobManager(t)
	outputjob := RunTestJob(jm, ns, completechannel, deletechannel, t, command, notifier, nil)
	if outputjob.Status.Failed != 1 {
		t.Fatalf("Jobs didn't fail")
	}
}

func TestCreateJobwithVolumes(t *testing.T) {
	t.Logf("TestCreateJobwithVolumes")
	// SkipCI(t)
	completechannel, deletechannel, notifier := getNotifier()
	// find render vol by name
	jm, ns := GetJobManager(t)
	render := jm.FindpvClaimByName("render", ns)
	source := jm.FindpvClaimByName("source", ns)
	renderref := jm.CreatePvCMountReference(render, "/site")
	srcref := jm.CreatePvCMountReference(source, "/src")
	refs := []PVClaimMountRef{renderref, srcref}

	// find source vol by name
	// create volumemount with name and path
	outputjob := RunTestJob(jm, ns, completechannel, deletechannel, t, nil, notifier, refs)
	if outputjob.Status.Succeeded != 1 {
		t.Fatalf("Jobs didn't succeed")
	}
}

func TestGetConfig(t *testing.T) {
	// SkipCI(t)
	t.Logf("TestGetConfig")

	c := getTestConfig(t)

	if c == nil {
		t.Fatalf("Unable to get config")
	}
	// TODO flag for job to autodelete
	// TODO wait for error exit
}

func TestCreateJobExitsError(t *testing.T) {

}

// test for other objects created doesn't fire job completion
// test for simple job create and exit

// test for job error state

// test for job that never returns and manually terminated
// test for job already exists
