// Package jobmanager is an abstraction for scheduling and managing jobs
package jobmanager

import (
	"context"
	"fmt"
	"log"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	kubelayer "github.com/clarkezone/previewd/pkg/kubelayer"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

// ResourseStateType is used in the notification callback
type ResourseStateType int

const (
	// Create indicates a job was just created
	Create = 0
	// Update indicates a job was just updated
	Update
	// Delete indicates a job was just deleted
	Delete
)

type jobnotifier func(*batchv1.Job, ResourseStateType)

type jobdescriptor struct {
	name       string
	namespace  string
	image      string
	command    []string
	args       []string
	notifier   jobnotifier
	autoDelete bool
	mountlist  []kubelayer.PVClaimMountRef
}

type jobupdate struct {
	job   *batchv1.Job
	typee ResourseStateType
}

type jobxxx interface {
	CreateJob(name string, namespace string,
		image string, command []string, args []string, notifier jobnotifier,
		autoDelete bool, mountlist []kubelayer.PVClaimMountRef) (*batchv1.Job, error)
}

// Jobmanager enables scheduling and querying of jobs
type Jobmanager struct {
	currentConfig    *rest.Config
	currentClientset kubernetes.Interface
	ctx              context.Context
	cancel           context.CancelFunc
	jobnotifiers     map[string]jobnotifier
	addQueue         chan jobdescriptor
	monitorExit      chan bool
	monitorDone      chan bool
}

// Newjobmanager is a factory method to create a new instanace of a job manager
func Newjobmanager(config *rest.Config, namespace string, startwatchers bool) (*Jobmanager, error) {
	clarkezoneLog.Debugf("Newjobmanager called with incluster:%v, namespace:%v", config, namespace)
	if config == nil {
		return nil, fmt.Errorf("config supplied is nil")
	}
	jm := newjobmanagerinternal(config)

	clientset, err := kubernetes.NewForConfig(jm.currentConfig)
	if err != nil {
		clarkezoneLog.Errorf("unable to create new clientset for config:%v", err)
		return nil, err
	}
	jm.currentClientset = clientset

	if startwatchers {
		clarkezoneLog.Debugf("Starting watchers")
		created := jm.startWatchers(namespace)
		jm.startMonitor(nil)
		if created {
			clarkezoneLog.Debugf("watchers sarted correctly")
			return jm, nil
		}

		clarkezoneLog.Debugf("watchers failed to start correctly")
		return nil, fmt.Errorf("unable to create jobmanager; startwatchers failed")
	}
	return jm, nil
}

// nolint
func newjobmanagerwithclient(clientset kubernetes.Interface, namespace string) (*Jobmanager, error) {
	clarkezoneLog.Debugf("newjobmanagerwithclient called with clientset:%v, namespace:%v",
		clientset, namespace)
	jm := newjobmanagerinternal(nil)

	jm.currentClientset = clientset

	// TODO only if we want watchers
	created := jm.startWatchers(namespace)
	if created {
		return jm, nil
	}
	clarkezoneLog.Debugf("watchers failed to start correctly")
	return nil, fmt.Errorf("unable to create jobmanaer; startwatchers failed")
}

func newjobmanagerinternal(config *rest.Config) *Jobmanager {
	if config != nil {
		clarkezoneLog.Debugf("newjobmanagerinternal called with incluster:%v", config)
	} else {
		clarkezoneLog.Debugf("newjobmanagerinternal called with nil config")
	}
	jm := Jobmanager{}

	ctx, cancel := context.WithCancel(context.Background())
	jm.ctx = ctx
	jm.cancel = cancel
	jm.jobnotifiers = make(map[string]jobnotifier)

	jm.currentConfig = config
	jm.addQueue = make(chan jobdescriptor)
	return &jm
}

func (jm *Jobmanager) startMonitor(jobcontroller jobxxx) {
	// TODO ensure monitor isn't already running
	jobqueue := make([]jobdescriptor, 0)
	jm.monitorExit = make(chan bool)
	jm.monitorDone = make(chan bool)
	go func() {
		// define queue for structs
		// create channel to pass to notifiers
		jobnotifierchannel := make(chan *jobupdate)
		clarkezoneLog.Debugf("Starting job monitor")
		// TODO: replace with infinite loop
		defer func() {
			clarkezoneLog.Debugf("Loop exited")
			close(jm.monitorExit)
		}()
		for {
			select {
			case nextJob := <-jm.addQueue:
				// push onto queue
				if nextJob.name != "" {
					jobqueue = append(jobqueue, nextJob)
				}
			case update := <-jobnotifierchannel:
				// k8s job completed is jobcommpleted function
				readyNext := isCompleted(update)
				if readyNext {
					err := jm.DeleteJob(update.job.Name, update.job.Namespace)
					if err != nil {
						clarkezoneLog.Errorf("Unable to delete job %v due to error %v", update.job.Name, err)
					}
				}
			case <-jm.monitorDone:
				clarkezoneLog.Debugf("monitorDone signalled, exiting loop")
				return
			}
			// if queue contains jobs, schedule new job
			if len(jobqueue) > 0 {
				clarkezoneLog.Debugf("jobqueue contains > 1 jobs, scheduling")
				nextjob := jobqueue[0]
				jobqueue = jobqueue[1:]
				notifier := func(job *batchv1.Job, typee ResourseStateType) {
					clarkezoneLog.Debugf("Got job in outside world %v", typee)
					// signal to notifierchannel
					jobnotifierchannel <- &jobupdate{job, typee}
				}
				_, err := jobcontroller.CreateJob(nextjob.name, nextjob.namespace, nextjob.image, nextjob.command,
					nextjob.args, notifier, false, nextjob.mountlist)
				if err != nil {
					clarkezoneLog.Debugf("Error creating job %v", err)
				}
			}
		} // for
	}()
}

func (jm *Jobmanager) stopMonitor() {
	jm.monitorDone <- true
	<-jm.monitorExit
}

func (jm *Jobmanager) startWatchers(namespace string) bool {
	clarkezoneLog.Debugf("startWatchers called with incluster:%v", namespace)
	// We will create an informer that writes added pods to a channel.
	var info informers.SharedInformerFactory
	if namespace == "" {
		// when watching in global scope, we need clusterrole / clusterrolebinding not role / rolebinding in the rbac setup
		info = informers.NewSharedInformerFactory(jm.currentClientset, 0)
	} else {
		info = informers.NewSharedInformerFactoryWithOptions(jm.currentClientset, 0, informers.WithNamespace(namespace))
	}
	podInformer := info.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Printf("pod added: %s/%s", pod.Namespace, pod.Name)
			//	pods <- pod
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			log.Printf("pod deleted: %s/%s", pod.Namespace, pod.Name)
		},
	})

	jobInformer := info.Batch().V1().Jobs().Informer()

	jobInformer.AddEventHandler(jm.getJobEventHandlers())
	err := jobInformer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		// your code goes here
		clarkezoneLog.Errorf("watcher errorhandler caught error: %v", err.Error())
		jm.cancel()
	})
	if err != nil {
		clarkezoneLog.Errorf("Unable to set watcher error handler with %v, aborting", err)
		panic(err)
	}
	info.Start(jm.ctx.Done())

	// Ensuring that the informer goroutine have warmed up and called List before
	// we send any events to it.
	result := cache.WaitForCacheSync(jm.ctx.Done(), podInformer.HasSynced)
	result2 := cache.WaitForCacheSync(jm.ctx.Done(), jobInformer.HasSynced)
	if !result || !result2 {
		clarkezoneLog.Errorf("Waitforcachesync failed")
		return false
	}
	return true
}

func (jm *Jobmanager) getJobEventHandlers() *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			job := obj.(*batchv1.Job)
			clarkezoneLog.Infof("Job added: %s/%s uid:%v", job.Namespace, job.Name, job.UID)
			if val, ok := jm.jobnotifiers[job.Name]; ok {
				val(job, Create)
			}
		},
		DeleteFunc: func(obj interface{}) {
			job := obj.(*batchv1.Job)
			clarkezoneLog.Infof("Job deleted: %s/%s uid:%v", job.Namespace, job.Name, job.UID)
			if val, ok := jm.jobnotifiers[job.Name]; ok {
				val(job, Delete)
				delete(jm.jobnotifiers, job.Name)
			}
		},
		UpdateFunc: func(oldobj interface{}, newobj interface{}) {
			oldjob := oldobj.(*batchv1.Job)
			newjob := newobj.(*batchv1.Job)
			clarkezoneLog.Infof("Job updated: %s/%s status:%v uid:%v", oldjob.Namespace, oldjob.Name, newjob.Status, newjob.UID)

			if val, ok := jm.jobnotifiers[newjob.Name]; ok {
				val(newjob, Update)
			}
		},
	}
}

func isCompleted(ju *jobupdate) bool {
	clarkezoneLog.Debugf("isCompleted() type:%v name:%v namespace:%v", ju.typee, ju.job.Name, ju.job.Namespace)

	if ju.typee == Update && ju.job.Status.Failed > 0 {
		clarkezoneLog.Debugf("Job failed")
		return true
	}

	if ju.typee == Update && ju.job.Status.Succeeded > 0 {
		clarkezoneLog.Debugf("Job succeeded")
		return true
	}
	return false
}

// FindpvClaimByName searches for a persistentvolumeclaim by name
func (jm *Jobmanager) FindpvClaimByName(pvname string, namespace string) (string, error) {
	return kubelayer.FindpvClaimByName(jm.currentClientset, pvname, namespace)
}

// CreatePvCMountReference creates a reference based on name and mountpoint
func (jm *Jobmanager) CreatePvCMountReference(claimname string,
	mountPath string, readOnly bool) kubelayer.PVClaimMountRef {
	claim := kubelayer.PVClaimMountRef{}
	claim.PVClaimName = claimname
	claim.MountPath = mountPath
	claim.ReadOnly = readOnly
	return claim
}

// CreateJob makes a new job
func (jm *Jobmanager) CreateJob(name string, namespace string,
	image string, command []string, args []string, notifier jobnotifier,
	autoDelete bool, mountlist []kubelayer.PVClaimMountRef) (*batchv1.Job, error) {
	clarkezoneLog.Debugf("CreateJob() called with name %v, namespace:%v,"+
		"image:%v, command:%v, args:%v, notifier:%v, autodelete:%v, pvlist:%v",
		name, namespace, image, command, args, notifier, autoDelete, mountlist)
	// TODO: if job exists, delete it
	job, err := kubelayer.CreateJob(jm.currentClientset, name, namespace,
		image, command, args, true, autoDelete, mountlist)
	if err != nil {
		return nil, err
	}
	if notifier != nil {
		jm.jobnotifiers[job.Name] = notifier
	}
	return job, nil
}

// AddJobtoQueue adds a job to the processing queue
func (jm *Jobmanager) AddJobtoQueue(name string, namespace string,
	image string, command []string, args []string,
	mountlist []kubelayer.PVClaimMountRef) error {
	// TODO do we need to deep copy command array?
	jm.addQueue <- jobdescriptor{name: name, namespace: namespace, image: image, command: command,
		args: args, notifier: nil, autoDelete: false, mountlist: mountlist}
	return nil
}

// DeleteJob deletes a job
func (jm *Jobmanager) DeleteJob(name string, namespace string) error {
	clarkezoneLog.Debugf("DeleteJob() called with name:%v namespace:%v", name, namespace)
	return kubelayer.DeleteJob(jm.currentClientset, name, namespace)
}

// GetConfigIncluster returns a config that will work when caller is running in a k8s cluster
func GetConfigIncluster() (*rest.Config, error) {
	clarkezoneLog.Debugf("GetConfigIncluster() called with incluster")
	var config *rest.Config
	var err error
	config, err = rest.InClusterConfig()
	if err != nil {
		clarkezoneLog.Errorf("InClusterConfig() returned error %v", err)
	}
	return config, err
}

// GetConfigOutofCluster returns a config loaded from the supplied path
func GetConfigOutofCluster(kubepath string) (*rest.Config, error) {
	clarkezoneLog.Debugf("GetConfigOutofCluster() called with kubepath:%v", kubepath)
	var config *rest.Config
	var err error
	var kubeconfig = &kubepath
	// use the current context in kubeconfig
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		clarkezoneLog.Errorf("BuildConfigFromFlags() failed with %v", err)
	}
	return config, err
}

// Close cancels all jobmanager go routines
func (jm *Jobmanager) Close() {
	jm.cancel()
}
