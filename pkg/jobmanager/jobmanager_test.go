//go:build integration
// +build integration

package jobmanager

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	clienttesting "k8s.io/client-go/testing"

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

func RunTestJob(completechannel chan struct{}, deletechannel chan struct{},
	t *testing.T, command []string, notifier func(*batchv1.Job, ResourseStateType)) {
	// SkipCI(t)
	c := getTestConfig(t)
	jm, err := Newjobmanager(c, "testns")
	if err != nil {
		t.Errorf("job manager create failed")
	}
	defer jm.Close()
	if err != nil {
		t.Fatalf("Unable to create JobManager")
	}

	_, err = jm.CreateJob("alpinetest", "testns", "alpine", command, nil, notifier)
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	<-completechannel
	log.Println("Completed; attempting delete")
	err = jm.DeleteJob("alpinetest")
	if err != nil {
		t.Fatalf("Unable to delete job %v", err)
	}
	log.Println(("Deleted."))
	<-deletechannel
}

func getTestConfig(t *testing.T) *rest.Config {
	configpath := path.Join(gitRoot, "integration/k3s-c2.yaml")
	c, err := GetConfigOutofCluster(configpath)
	if err != nil {
		t.Fatalf("Couldn't get config %v", err)
	}
	return c
}

func TestCreateAndSucceed(t *testing.T) {
	t.Logf("TestCreateAndSucceed")
	// SkipCI(t)
	completechannel := make(chan struct{})
	deletechannel := make(chan struct{})
	notifier := (func(job *batchv1.Job, typee ResourseStateType) {
		log.Printf("Got job in outside world %v", typee)

		if completechannel != nil && typee == Update && job.Status.Active == 0 && job.Status.Failed > 0 {
			log.Printf("Error detected")
			close(completechannel)
			completechannel = nil // avoid double close
		}

		if typee == Delete {
			log.Printf("Deleted")
			close(deletechannel)
		}
	})
	command := []string{"error"}
	RunTestJob(completechannel, deletechannel, t, command, notifier)
}

func TestCreateAndFail(t *testing.T) {
	t.Logf("TestCreateAndFail")
	// SkipCI(t)
	client := fake.NewSimpleClientset()
	client.PrependWatchReactor("*", func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()
		watch, err := client.Tracker().Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}
		return false, watch, nil
	})

	jm, err := newjobmanagerwithclient(client, "testns")
	if err != nil {
		t.Errorf("error")
	}
	defer jm.Close()
	if err != nil {
		t.Fatalf("Unable to create JobManager")
	}
	completechannel := make(chan struct{})
	deletechannel := make(chan struct{})
	notifier := (func(job *batchv1.Job, typee ResourseStateType) {
		log.Printf("Got job in outside world %v Active %v Failed %v", typee, job.Status.Active, job.Status.Failed)

		if completechannel != nil && typee == Update && job.Status.Active == 0 && job.Status.Failed > 0 {
			log.Printf("Error detected")
			close(completechannel)
			completechannel = nil // avoid double close
		}

		if typee == Delete {
			log.Printf("Deleted")
			close(deletechannel)
		}
	})
	command := []string{"error"}
	_, err = jm.CreateJob("alpinetest", "", "alpine", command, nil, notifier)
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	<-completechannel
	log.Println("Completed; attempting delete")
	err = jm.DeleteJob("alpinetest")
	if err != nil {
		t.Fatalf("Unable to delete job %v", err)
	}
	log.Println(("Deleted."))
	<-deletechannel
	//TODO: [x] add delete function
	//TODO: Move logic into test for succeeded / failed job incl delete.. does it work with mock
	//TODO: ability to inject volumes
	//TODO: support for namespace
	//TODO:         verbose logging
	//TODO:             Conditional log statements
	//TODO:             Environment variable
	//TODO: flag for job to autodelete
	// TODO: test that verifies auto delete
	// TODO: Ensure error if job with same name already exists
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
