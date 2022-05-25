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

package kubelayer

import (
	"log"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apimachinery "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"

	"github.com/clarkezone/previewd/internal"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

const (
	testNamespace = "testns"
)

func GetKubeSession(t *testing.T) *KubeSession {
	c := getTestConfig(t)
	ks, err := Newkubesession(c)
	if err != nil {
		t.Fatalf("failed to get kubesession %v", err)
	}
	err = ks.StartWatchers(testNamespace)
	if err != nil {
		t.Fatalf("failed to start watchers %v", err)
	}

	return ks
}

func createNSIfMissing(ks *KubeSession, recreate bool, t *testing.T) {
	ns, err := ks.GetNamespace(testNamespace)
	if err != nil && !apimachinery.IsNotFound(err) {
		t.Fatalf("failed to get namespace %v", err)
	}

	if recreate && ns != nil {
		wait := make(chan bool)
		err = ks.DeleteNamespace(testNamespace, func(ns *corev1.Namespace, rt ResourseStateType) {
			if rt == Delete {
				wait <- true
			}
		})
		if err != nil {
			t.Fatalf("CreateNamespace failed %v", err)
		}
		<-wait
		ns = nil
	}

	if ns == nil {
		wait := make(chan bool)
		err = ks.CreateNamespace(testNamespace, func(ns *corev1.Namespace, rt ResourseStateType) {
			if rt == Create {
				wait <- true
			}
		})
		if err != nil {
			t.Fatalf("CreateNamespace failed %v", err)
		}
		<-wait
	}
}

func RunTestJob(ks *KubeSession, testNamespace string, completechannel chan batchv1.Job, deletechannel chan batchv1.Job,
	t *testing.T, command []string, notifier func(*batchv1.Job, ResourseStateType),
	mountlist []PVClaimMountRef) batchv1.Job {
	// SkipCI(t)
	defer ks.Close()

	_, err := ks.CreateJob("alpinetest", testNamespace, "alpine", command, nil, notifier, false, mountlist)
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

func TestCreateAndSucceed(t *testing.T) {
	t.Logf("TestCreateAndSucceed")
	completechannel, deletechannel, notifier := getNotifier()
	ks := GetKubeSession(t)
	createNSIfMissing(ks, false, t)
	outputjob := RunTestJob(ks, testNamespace, completechannel, deletechannel, t, nil, notifier, nil)
	if outputjob.Status.Succeeded != 1 {
		t.Fatalf("Jobs didn't succeed")
	}
}

func TestCreateAndErrorWork(t *testing.T) {
	t.Logf("TestCreateAndSucceed")
	completechannel, deletechannel, notifier := getNotifier()
	command := []string{"error"}
	ks := GetKubeSession(t)
	createNSIfMissing(ks, true, t)
	outputjob := RunTestJob(ks, testNamespace, completechannel, deletechannel, t, command, notifier, nil)
	if outputjob.Status.Failed != 1 {
		t.Fatalf("Jobs didn't fail")
	}
}

func TestCreatePersistentVolumeClaim(t *testing.T) {
	ks := GetKubeSession(t)
	createNSIfMissing(ks, true, t)
	err := ks.CreatePersistentVolumeClaim("source", testNamespace)
	if err != nil {
		t.Fatalf("unable to creates persistent volume claim %v", err)
	}

	err = ks.DeleteNamespace(testNamespace, nil)
	if err != nil {
		t.Fatalf("unable to delete namespace %v", err)
	}
}

func TestFindVolumeSuccess(t *testing.T) {
	const name = "render"
	ks := GetKubeSession(t)
	createNSIfMissing(ks, true, t)

	err := ks.CreatePersistentVolumeClaim(name, testNamespace)
	if err != nil {
		t.Fatalf("unable to creates persistent volume claim %v", err)
	}
	render, err := ks.FindpvClaimByName(name, testNamespace)
	if err != nil {
		t.Fatalf("can't find pvcalim render %v", err)
	}
	if render == "" {
		t.Fatalf("didn't find render volume %v in namespace %v", name, testNamespace)
	}
}

func TestFindVolumeFail(t *testing.T) {
	const name = "notexists"
	ks := GetKubeSession(t)
	createNSIfMissing(ks, false, t)
	render, err := ks.FindpvClaimByName(name, testNamespace)
	if err != nil {
		t.Fatalf("error finding pvcalim %v: %v", name, err)
	}
	if render != "" {
		t.Fatalf("render should be nil")
	}
}

func TestCreateJobwithVolumes(t *testing.T) {
	// TODO: this test takes longer than 30 seconds and hits timeout
	t.Logf("TestCreateJobwithVolumes")
	const rendername = "render"
	const sourcename = "source"
	completechannel, deletechannel, notifier := getNotifier()
	// find render vol by name
	ks := GetKubeSession(t)
	createNSIfMissing(ks, true, t)

	err := ks.CreatePersistentVolumeClaim(sourcename, testNamespace)
	if err != nil {
		t.Fatalf("unable to create persistent volume claim %v", err)
	}

	err = ks.CreatePersistentVolumeClaim(rendername, testNamespace)
	if err != nil {
		t.Fatalf("unable to create persistent volume claim %v", err)
	}
	render, err := ks.FindpvClaimByName(rendername, testNamespace)
	if err != nil {
		t.Fatalf("can't find pvcalim render %v", err)
	}
	if render == "" {
		t.Fatalf("render name empty")
	}
	source, err := ks.FindpvClaimByName(sourcename, testNamespace)
	if err != nil {
		t.Fatalf("can't find pvcalim source %v", err)
	}
	if source == "" {
		t.Fatalf("source name empty")
	}
	renderref := ks.CreatePvCMountReference(render, "/site", false)
	srcref := ks.CreatePvCMountReference(source, "/src", true)
	refs := []PVClaimMountRef{renderref, srcref}

	outputjob := RunTestJob(ks, testNamespace, completechannel, deletechannel, t, nil, notifier, refs)
	if outputjob.Status.Succeeded != 1 {
		t.Fatalf("Jobs didn't succeed")
	}
}

func TestCreateDeletetestNamespace(t *testing.T) {
	ks := GetKubeSession(t)
	createNSIfMissing(ks, true, t)
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

func getTestConfig(t *testing.T) *rest.Config {
	configPath := internal.GetTestConfigPath(t)
	c, err := GetConfigOutofCluster(configPath)
	if err != nil {
		t.Fatalf("Couldn't get config %v", err)
	}
	return c
}
