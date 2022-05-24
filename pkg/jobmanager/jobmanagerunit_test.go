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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testNamespace = "testns"
)

type MockJobManager struct {
	mock.Mock
	notifier                jobnotifier
	done                    chan bool
	jobFail                 bool
	scheduledByMeinProgress int
}

// Implement jobxxx interface begin
func (o *MockJobManager) CreateJob(name string, namespace string,
	image string, command []string, args []string, notifier jobnotifier,
	autoDelete bool, mountlist []kubelayer.PVClaimMountRef) (*batchv1.Job, error) {
	// schedule callbacks to mimic kube
	o.notifier = notifier
	o.Called(name, namespace, image, command, args, notifier, autoDelete, mountlist)
	o.launchSuccess(name, namespace)
	// TODO: track jobs i've scheduled and do more accurate refcount
	o.scheduledByMeinProgress++
	return nil, nil
}

func (o *MockJobManager) DeleteJob(name string, namespace string) error {
	o.Called(name, namespace)
	// TODO: track jobs i've scheduled and do more accurate refcount
	o.scheduledByMeinProgress--
	o.done <- true
	return nil
}

func (o *MockJobManager) FailedJob(name string, namespace string) {
	o.Called(name, namespace)
	o.done <- true
}

func (o *MockJobManager) InProgress() bool {
	return o.scheduledByMeinProgress > 0
}

// Implement jobxxx interface end

func (o *MockJobManager) launchSuccess(name string, namespace string) {
	go func() {
		j := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		j.Status = batchv1.JobStatus{Active: 1}
		o.notifier(j, Create)

		if o.jobFail {
			j = &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			}
			j.Status = batchv1.JobStatus{Failed: 1}
			o.notifier(j, Update)
		} else {
			j = &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			}
			j.Status = batchv1.JobStatus{Succeeded: 1}
			o.notifier(j, Update)
		}
	}()
}

func (o *MockJobManager) WaitDone(t *testing.T, numjobs int) {
	for i := 0; i < numjobs; i++ {
		select {
		case <-o.done:
		case <-time.After(10 * time.Second):
			t.Fatalf("No done before 10 second timeout")
		}
	}
}

func (o *MockJobManager) SetJobFail() {
	o.jobFail = true
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

func TestSingleJobSucess(t *testing.T) {
	jm, mjm := getJobManagerMockedMonitor(t)
	mjm.On("CreateJob", "alpinetest", "testns",
		"alpine", mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
		mock.AnythingOfType("jobmanager.jobnotifier"), false,
		[]kubelayer.PVClaimMountRef{}).Return(&batchv1.Job{}, nil)
	mjm.On("DeleteJob", "alpinetest", "testns")
	err := jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	// This wait will be completed when delete is called on the mockjobmanager
	mjm.WaitDone(t, 1)
	mjm.AssertExpectations(t)
	jm.stopMonitor()
}

func TestMultiJobSuccess(t *testing.T) {
	jm, mjm := getJobManagerMockedMonitor(t)

	mjm.On("CreateJob", "alpinetest", "testns",
		"alpine", mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
		mock.AnythingOfType("jobmanager.jobnotifier"), false,
		[]kubelayer.PVClaimMountRef{}).Return(&batchv1.Job{}, nil)
	mjm.On("DeleteJob", "alpinetest", "testns")

	mjm.On("CreateJob", "alpinetest2", "testns",
		"alpine", mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
		mock.AnythingOfType("jobmanager.jobnotifier"), false,
		[]kubelayer.PVClaimMountRef{}).Return(&batchv1.Job{}, nil)
	mjm.On("DeleteJob", "alpinetest2", "testns")

	err := jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}

	err = jm.AddJobtoQueue("alpinetest2", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	// This wait will be completed when delete is called on the mockjobmanager
	mjm.WaitDone(t, 2)
	mjm.AssertExpectations(t)
	jm.stopMonitor()
}

func TestSingleJobFail(t *testing.T) {
	jm, mjm := getJobManagerMockedMonitor(t)

	mjm.On("CreateJob", "alpinetest", "testns",
		"alpine", mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
		mock.AnythingOfType("jobmanager.jobnotifier"), false,
		[]kubelayer.PVClaimMountRef{}).Run(func(args mock.Arguments) {
		mjm.SetJobFail()
	})

	mjm.On("FailedJob", "alpinetest", "testns")

	err := jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}

	// This wait will be completed when delete is called on the mockjobmanager
	mjm.WaitDone(t, 1)
	mjm.AssertExpectations(t)
	jm.stopMonitor()
}
func TestMultiJobFail(t *testing.T) {
	jm, mjm := getJobManagerMockedMonitor(t)

	mjm.On("CreateJob", "alpinetest", "testns",
		"alpine", mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string"),
		mock.AnythingOfType("jobmanager.jobnotifier"), false,
		[]kubelayer.PVClaimMountRef{}).Run(func(args mock.Arguments) {
		mjm.SetJobFail()
	})

	mjm.On("FailedJob", "alpinetest", "testns")
	err := jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}

	err = jm.AddJobtoQueue("alpinetest2", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}
	// This wait will be completed when delete is called on the mockjobmanager
	mjm.WaitDone(t, 1)
	mjm.AssertExpectations(t)
	jm.stopMonitor()
}

// TestMain initizlie all tests
func TestMain(m *testing.M) {
	clarkezoneLog.Init(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}
