package jobmanager

import (
	"testing"

	kubelayer "github.com/clarkezone/previewd/pkg/kubelayer"
	"github.com/stretchr/testify/mock"
	batchv1 "k8s.io/api/batch/v1"
)

type MockJobManager struct {
	mock.Mock
}

func (o *MockJobManager) CreateJob(name string, namespace string,
	image string, command []string, args []string, notifier jobnotifier,
	autoDelete bool, mountlist []kubelayer.PVClaimMountRef) (*batchv1.Job, error) {
	return nil, nil
}

func newMockJobManager() *MockJobManager {
	mjm := MockJobManager{}
	mjm.On("CreateJjob", "", []string{}, mock.AnythingOfType("jobmanager.jobnotifier"), false,
		[]kubelayer.PVClaimMountRef{})
	return &mjm
}

func getJobManagerMockedMonitor(t *testing.T) *Jobmanager {
	jm, _ := GetJobManager(t, testNamespace, false)
	mjm := newMockJobManager()
	jm.startMonitor(mjm)
	return jm
}

func TestStartMonitor(t *testing.T) {
	//ch := make(chan bool)
	jm := getJobManagerMockedMonitor(t)

	//<-ch
	jm.stopMonitor()
}

func TestSingleJobAdded(t *testing.T) {
	jm := getJobManagerMockedMonitor(t)
	err := jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}

	// TODO verify createjob called on mock
	// mock dispatches messages to job causing exit
	// TODO verify queue is back in steady state
	jm.stopMonitor()
}

func TestMultiJobAdded(t *testing.T) {
}
