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
}

func (o *MockJobManager) CreateJob(name string, namespace string,
	image string, command []string, args []string, notifier jobnotifier,
	autoDelete bool, mountlist []kubelayer.PVClaimMountRef) (*batchv1.Job, error) {
	// schedule callbacks to mimic kube
	return nil, nil
}

func newMockJobManager() *MockJobManager {
	mjm := MockJobManager{}
	mjm.On("CreateJjob", "", []string{}, mock.AnythingOfType("jobmanager.jobnotifier"), false,
		[]kubelayer.PVClaimMountRef{})
	return &mjm
}

// nolint
func getJobManagerMockedMonitor(t *testing.T) *Jobmanager {
	jm := newjobmanagerinternal(nil)
	mjm := newMockJobManager()
	jm.startMonitor(mjm)
	return jm
}

func TestStartMonitor(t *testing.T) {
	jm := getJobManagerMockedMonitor(t)

	jm.stopMonitor()
}

func TestSingleJobAdded(t *testing.T) {
	waitchan := make(chan bool)
	jm := getJobManagerMockedMonitor(t)
	err := jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	select {
	case <-waitchan:
	case <-time.After(10 * time.Second):
		fmt.Println("timeout 10")
	}

	// TODO verify createjob called on mock
	// mock dispatches messages to job causing exit
	// TODO verify queue is back in steady state
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
