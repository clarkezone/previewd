package jobmanager

import (
	"fmt"
	"os"
	"testing"
	"time"

	kubelayer "github.com/clarkezone/previewd/pkg/kubelayer"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	batchv1 "k8s.io/api/batch/v1"
)

const (
	testNamespace = "testns"
)

type MockJobManager struct {
	mock.Mock
	notifier jobnotifier
	done     chan bool
}

func (o *MockJobManager) CreateJob(name string, namespace string,
	image string, command []string, args []string, notifier jobnotifier,
	autoDelete bool, mountlist []kubelayer.PVClaimMountRef) (*batchv1.Job, error) {
	// schedule callbacks to mimic kube
	o.notifier = notifier
	o.launchSuccess()
	return nil, nil
}

func (o *MockJobManager) DeleteJob(name string, namespace string) error {
	o.done <- true
	return nil
}

func (o *MockJobManager) launchSuccess() {
	go func() {
		j := &batchv1.Job{}
		j.Status = batchv1.JobStatus{Active: 1}
		o.notifier(j, Create)

		j = &batchv1.Job{}
		j.Status = batchv1.JobStatus{Succeeded: 1}
		o.notifier(j, Update)
	}()
}

func (o *MockJobManager) WaitDone() {
	select {
	case <-o.done:
	case <-time.After(10 * time.Second):
		fmt.Println("timeout 10")
	}
}

func newMockJobManager() *MockJobManager {
	mjm := MockJobManager{}
	mjm.done = make(chan bool)
	mjm.On("CreateJjob", "", []string{}, mock.AnythingOfType("jobmanager.jobnotifier"), false,
		[]kubelayer.PVClaimMountRef{})
	return &mjm
}

// nolint
func getJobManagerMockedMonitor(t *testing.T) (*Jobmanager, *MockJobManager) {
	jm := newjobmanagerinternal(nil)
	mjm := newMockJobManager()
	jm.startMonitor(mjm)
	return jm, mjm
}

func TestStartMonitor(t *testing.T) {
	jm, _ := getJobManagerMockedMonitor(t)

	jm.stopMonitor()
}

func TestSingleJobAdded(t *testing.T) {
	jm, mjm := getJobManagerMockedMonitor(t)
	err := jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	// This will be triggered when delete is called on the mockjobmanager
	mjm.WaitDone()
	// TODO verify createjob called on mock
	jm.stopMonitor()
}

func TestMultiJobAdded(t *testing.T) {
}

// TestMain initizlie all tests
func TestMain(m *testing.M) {
	clarkezoneLog.Init(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}
