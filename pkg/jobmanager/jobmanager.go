// Package jobmanager is an abstraction for scheduling and managing jobs
package jobmanager

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/rest"

	"github.com/clarkezone/previewd/internal"
	kubelayer "github.com/clarkezone/previewd/pkg/kubelayer"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

type jobdescriptor struct {
	name       string
	namespace  string
	image      string
	command    []string
	args       []string
	notifier   kubelayer.JobNotifier
	autoDelete bool
	mountlist  []kubelayer.PVClaimMountRef
}

type jobupdate struct {
	job   *batchv1.Job
	typee kubelayer.ResourseStateType
}

// Jobxxx is a provider interface for tracking job status
type Jobxxx interface {
	CreateJob(name string, namespace string,
		image string, command []string, args []string, notifier kubelayer.JobNotifier,
		autoDelete bool, mountlist []kubelayer.PVClaimMountRef) (*batchv1.Job, error)
	DeleteJob(name string, namespace string) error
	FailedJob(name string, namesapce string)
	// InProgress returns true if jobs we have scheduled are in progress
	InProgress() bool
}

// Jobmanager enables scheduling and querying of jobs
type Jobmanager struct {
	kubeSession   *kubelayer.KubeSession
	namespace     string
	addQueue      chan jobdescriptor
	monitorExit   chan bool
	monitorDone   chan bool
	haveFailedJob bool
	// TODO: is there a better way to avoid test need impacting API shape
	// JobProvider is public to enable integration tests to modify
	JobProvider Jobxxx
}

type kubeJobManager struct {
	kubeSession *kubelayer.KubeSession
	jobRefs     map[string]string
}

// Implement jobxxx interface begin
func (o *kubeJobManager) CreateJob(name string, namespace string,
	image string, command []string, args []string, notifier kubelayer.JobNotifier,
	autoDelete bool, mountlist []kubelayer.PVClaimMountRef) (*batchv1.Job, error) {
	clarkezoneLog.Debugf("CreateJob called with name:%v, namespace:%v, image:%v", name, namespace,
		image)
	o.jobRefs[name] = name
	return o.kubeSession.CreateJob(name, namespace,
		image, command, args, notifier, autoDelete, mountlist)
}

func (o *kubeJobManager) DeleteJob(name string, namespace string) error {
	clarkezoneLog.Debugf("DeleteJob called with name:%v, namespace:%v", name, namespace)
	delete(o.jobRefs, name)
	return o.kubeSession.DeleteJob(name, namespace)
}

func (o *kubeJobManager) FailedJob(name string, namespace string) {
	clarkezoneLog.Debugf("FailedJob called with name:%v, namespace:%v", name, namespace)
}

func (o *kubeJobManager) InProgress() bool {
	return len(o.jobRefs) > 0
}

// Implement jobxxx interface end

// Newjobmanager is a factory method to create a new instanace of a job manager
func Newjobmanager(config *rest.Config, namespace string, startwatchers bool,
	startnsWatcher bool) (*Jobmanager, error) {
	clarkezoneLog.Debugf("Newjobmanager called with incluster:%v, namespace:%v, startwatchers:%v",
		config, namespace, startwatchers)
	if config == nil {
		return nil, fmt.Errorf("config supplied is nil")
	}
	kubeProvider := kubeJobManager{}
	jm, err := newjobmanagerinternal(config, &kubeProvider, namespace)
	if err != nil {
		return nil, err
	}

	kubeProvider.kubeSession = jm.kubeSession
	kubeProvider.jobRefs = make(map[string]string)

	if startwatchers {
		err = jm.StartWatchers(startnsWatcher)
		if err != nil {
			return nil, err
		}
	}
	return jm, nil
}

func newjobmanagerinternal(config *rest.Config, provider Jobxxx, namespace string) (*Jobmanager, error) {
	if config != nil {
		clarkezoneLog.Debugf("newjobmanagerinternal called with incluster:%v", config)
	} else {
		clarkezoneLog.Debugf("newjobmanagerinternal called with nil config")
	}
	jm := Jobmanager{}
	var err error
	if config != nil {
		jm.kubeSession, err = kubelayer.Newkubesession(config)
		if err != nil {
			return nil, err
		}
	}

	jm.addQueue = make(chan jobdescriptor)
	jm.JobProvider = provider
	jm.namespace = namespace
	return &jm, nil
}

// KubeSession returns the current active kubesession object
func (jm *Jobmanager) KubeSession() *kubelayer.KubeSession {
	// TODO: wrap it in an interface?
	return jm.kubeSession
}

// StartWatchers starts jobmonitoring infra for cases when these were not started in jobmanager creation
func (jm *Jobmanager) StartWatchers(watchNs bool) error {
	clarkezoneLog.Debugf("JobManager: Starting watchers with namespace %v", jm.namespace)
	err := jm.kubeSession.StartWatchers(jm.namespace, watchNs)

	if err != nil {
		return err
	}

	jm.startMonitor(jm.JobProvider)
	return nil
}

func (jm *Jobmanager) startMonitor(jobcontroller Jobxxx) {
	// TODO ensure monitor isn't already running
	jobqueue := make([]jobdescriptor, 0)
	jm.monitorExit = make(chan bool)
	jm.monitorDone = make(chan bool)
	go func() {
		// define queue for structs
		// create channel to pass to notifiers
		jobnotifierchannel := make(chan *jobupdate)
		clarkezoneLog.Debugf("startmonitor() starting job monitor")
		defer func() {
			clarkezoneLog.Debugf(" startMonitor: Loop exited")
			close(jm.monitorExit)
		}()
		for {
			select {
			case nextJob := <-jm.addQueue:
				clarkezoneLog.Debugf(" startMonitor(): received job from jm.addQueue channel")
				// push onto queue
				if nextJob.name != "" {
					clarkezoneLog.Debugf(" startMonitor(): nextJob name is not empty hence adding to jobqueue")
					jobqueue = append(jobqueue, nextJob)
				} else {
					clarkezoneLog.Debugf(" startMonitor(): nextJob name is empty hence not adding to jobqueue")
				}
			case update := <-jobnotifierchannel:
				clarkezoneLog.Debugf(" startMonitor(): received job notification from jobnotifierchannel")
				// k8s job completed is jobcommpleted function
				readyNext, failed := isCompleted(update)
				jm.haveFailedJob = failed
				switch {
				case readyNext && !failed:
					clarkezoneLog.Debugf(" startMonitor(): successfully completed job detected, deleting job")
					err := jobcontroller.DeleteJob(update.job.Name, update.job.Namespace)
					if err != nil {
						clarkezoneLog.Errorf("Unable to delete job %v due to error %v", update.job.Name, err)
					}
				case readyNext && failed:
					clarkezoneLog.Debugf(" startMonitor(): Failed completed job name:%v namespace:%v, cannot process further jobs",
						update.job.Name, update.job.Namespace)
					jobcontroller.FailedJob(update.job.Name, update.job.Namespace)
				default:
					clarkezoneLog.Debugf(" startMonitor(): Received non completed update")
				}
			case <-jm.monitorDone:
				clarkezoneLog.Debugf(" startMonitor(): jm.monitorDone channel signalled, exiting loop")
				return
			}
			// if queue contains jobs and no jobs in progress, schedule new job
			// signal to notifierchannel
			jm.scheduleIfPossible(&jobqueue, jobcontroller, jobnotifierchannel)
		} // for
	}()
}

func (jm *Jobmanager) scheduleIfPossible(jobqueue *[]jobdescriptor,
	jobcontroller Jobxxx, jobnotifierchannel chan *jobupdate) {
	jobQueueLength := len(*jobqueue)
	jobInProgress := jobcontroller.InProgress()
	clarkezoneLog.Debugf("scheduleIfPossible called jobqueue length:%v, jobcontroller.InProgress():%v",
		jobQueueLength, jobInProgress)
	if jobQueueLength > 0 && !jobInProgress {
		clarkezoneLog.Debugf(" scheduleIfPossible attempting to schedule")
		if jm.haveFailedJob {
			clarkezoneLog.Debugf(" scheduleIfPossible jobqueue contains > 1 jobs, but we have a failed job hence not scheduling")
		} else {
			clarkezoneLog.Debugf(" scheduleIfPossible jobqueue contains > 1 jobs, scheduling")
			nextjob := (*jobqueue)[0]
			*jobqueue = (*jobqueue)[1:]
			notifier := func(job *batchv1.Job, typee kubelayer.ResourseStateType) {
				clarkezoneLog.Debugf(" notifier called: Got job in outside world %v", typee)

				clarkezoneLog.Debugf(" notifier begin send job update to jobnotifierchannel")
				jobnotifierchannel <- &jobupdate{job, typee}
				clarkezoneLog.Debugf(" notifier end send job update to jobnotifierchannel")
			}
			_, err := jobcontroller.CreateJob(nextjob.name, nextjob.namespace, nextjob.image, nextjob.command,
				nextjob.args, notifier, false, nextjob.mountlist)
			if err != nil {
				clarkezoneLog.Debugf(" scheduleIfPossible Error creating job %v", err)
			}
		}
	} else {
		clarkezoneLog.Debugf(" scheduleIfPossible: nothing to schedule")
	}
}

func (jm *Jobmanager) stopMonitor() {
	clarkezoneLog.Debugf("stopMonitor begin")
	clarkezoneLog.Debugf(" stopMonitor begin send true to monitorDone channel")
	jm.monitorDone <- true
	clarkezoneLog.Debugf(" stopMonitor end send true to monitorDone channel")
	clarkezoneLog.Debugf(" stopMonitor begin wait for monitor exit")
	<-jm.monitorExit
	clarkezoneLog.Debugf(" stopMonitor end wait for monitor exit")
	clarkezoneLog.Debugf("stopMonitor end")
}

func isCompleted(ju *jobupdate) (bool, bool) {
	clarkezoneLog.Debugf("isCompleted() type:%v name:%v namespace:%v", ju.typee, ju.job.Name, ju.job.Namespace)

	if ju.typee == kubelayer.Update && ju.job.Status.Failed > 0 {
		clarkezoneLog.Debugf(" isCompleted Job failed")
		return true, true
	}

	if ju.typee == kubelayer.Update && ju.job.Status.Succeeded > 0 {
		clarkezoneLog.Debugf(" isCompleted Job succeeded")
		return true, false
	}

	clarkezoneLog.Debugf(" isCompleted Job not complete")
	return false, false
}

// AddJobtoQueue adds a job to the processing queue
func (jm *Jobmanager) AddJobtoQueue(name string, namespace string,
	image string, command []string, args []string,
	mountlist []kubelayer.PVClaimMountRef) error {
	clarkezoneLog.Debugf("AddJobtoQueue() called with name %v, namespace:%v,"+
		"image:%v, command:%v, args:%v, pvlist:%v",
		name, namespace, image, command, args, mountlist)
	// TODO do we need to deep copy command array?
	clarkezoneLog.Debugf(" addjobtoqueue: begin add job descriptor to jm.addQueue channel")
	jm.addQueue <- jobdescriptor{name: name, namespace: namespace, image: image, command: command,
		args: args, notifier: nil, autoDelete: false, mountlist: mountlist}
	clarkezoneLog.Debugf(" addjobtoqueue: end add job descriptor to jm.addQueue channel")
	return nil
}

// CreateJekyllJob creates a render job using Jekyll
func CreateJekyllJob(ns string, ks *kubelayer.KubeSession, jm *Jobmanager) error {
	const rendername = "render"
	const sourcename = "source"
	render, err := ks.FindpvClaimByName(rendername, ns)
	if err != nil {
		clarkezoneLog.Errorf("CreateJekyllJob can't find pvcalim render %v", err)
	}
	if render == "" {
		clarkezoneLog.Errorf("CreateJekyllJob render name empty")
	}
	source, err := ks.FindpvClaimByName(sourcename, ns)
	if err != nil {
		clarkezoneLog.Errorf("CreateJekyllJob can't find pvcalim source %v", err)
	}
	if source == "" {
		clarkezoneLog.Errorf("CreateJekyllJob source name empty")
	}
	renderref := ks.CreatePvCMountReference(render, "/site", false)
	srcref := ks.CreatePvCMountReference(source, "/src", false)
	refs := []kubelayer.PVClaimMountRef{renderref, srcref}
	imagePath := internal.GetJekyllImage()

	command, params := internal.GetJekyllCommands()
	err = jm.AddJobtoQueue("jekyll-render-container", ns, imagePath, command, params, refs)
	if err != nil {
		clarkezoneLog.Errorf("Failed to create job: %v\n", err.Error())
	}
	return err
}
