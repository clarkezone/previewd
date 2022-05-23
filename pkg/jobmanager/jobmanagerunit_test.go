package jobmanager

import (
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
	o.Called(name, namespace, image, command, args, notifier, autoDelete, mountlist)
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

func (o *MockJobManager) WaitDone(t *testing.T) {
	select {
	case <-o.done:
	case <-time.After(10 * time.Second):
		t.Fatalf("No done before 10 second timeout")
	}
}

func (o *MockJobManager) ConfirmSuccess(t *testing.T) {
	if !o.AssertCalled(t, "CreateJob") {
		t.Fatalf("CreateJob not called")
	}
}

func newMockJobManager() *MockJobManager {
	mjm := MockJobManager{}
	mjm.done = make(chan bool)
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
	mjm.On("CreateJob", "alpinetest", "testns",
		"alpine", mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
		mock.AnythingOfType("jobmanager.jobnotifier"), false,
		[]kubelayer.PVClaimMountRef{}).Return(&batchv1.Job{}, nil)
	err := jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	// This wait will be completed when delete is called on the mockjobmanager
	mjm.WaitDone(t)
	mjm.AssertExpectations(t)
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
