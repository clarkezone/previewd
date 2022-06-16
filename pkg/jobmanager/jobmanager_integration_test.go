//go:build integration
// +build integration

// open settings json or remote settings json
// {
//"go.buildFlags": [
//    "-tags=unit,integration"
//],
//"go.buildTags": "-tags=unit,integration",
//"go.testTags": "-tags=unit,integration"
// }

package jobmanager

import (
	"testing"
	"time"

	"github.com/clarkezone/previewd/internal"
	kubelayer "github.com/clarkezone/previewd/pkg/kubelayer"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	batchv1 "k8s.io/api/batch/v1"
)

type CompletionTrackingJobManager struct {
	wrappedJob      Jobxxx
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

func newCompletionTrackingJobManager(towrap Jobxxx, timeout int) *CompletionTrackingJobManager {
	wrapped := &CompletionTrackingJobManager{wrappedJob: towrap}
	wrapped.timeoutinterval = timeout
	wrapped.done = make(chan bool)
	return wrapped
}

func TestCreateJobE2E(t *testing.T) {
	// TODO: ensure namespace exists
	internal.SetupGitRoot()
	path := internal.GetTestConfigPath(t)
	config, err := kubelayer.GetConfigOutofCluster(path)
	if err != nil {
		t.Fatalf("Can't get config %v", err)
	}
	jm, err := Newjobmanager(config, "testns", false, true)
	if err != nil {
		t.Fatalf("Can't create jobmanager %v", err)
	}
	wrappedProvider := newCompletionTrackingJobManager(jm.JobProvider, 20)

	jm.JobProvider = wrappedProvider

	// Watchers must be started after the provider has been wrapped
	jm.StartWatchers(true)
	err = jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}

	wrappedProvider.WaitDone(t, 3)
}

func TestCompletionTrackingMultiple(t *testing.T) {
	internal.SetupGitRoot()
	path := internal.GetTestConfigPath(t)
	config, err := kubelayer.GetConfigOutofCluster(path)
	if err != nil {
		t.Fatalf("Can't get config %v", err)
	}
	jm, err := Newjobmanager(config, "testns", false, true)
	if err != nil {
		t.Fatalf("Can't create jobmanager %v", err)
	}
	wrappedProvider := newCompletionTrackingJobManager(jm.JobProvider, 20)

	jm.JobProvider = wrappedProvider

	// Watchers must be started after the provider has been wrapped
	jm.StartWatchers(true)
	err = jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}

	wrappedProvider.WaitDone(t, 3)
	err = jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}

	wrappedProvider.WaitDone(t, 3)
}
