package kubelayer

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	batchv1 "k8s.io/api/batch/v1"

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

// KubeSession is a session to a k8s cluster
type KubeSession struct {
	currentConfig    *rest.Config
	currentClientset kubernetes.Interface
	ctx              context.Context
	cancel           context.CancelFunc
	jobnotifiers     map[string]jobnotifier
}

// Newkubesession creates a new kubesession from a config
func Newkubesession(config *rest.Config) (*KubeSession, error) {
	clarkezoneLog.Debugf("Newjobmanager called with incluster:%v, namespace:%v", config)
	if config == nil {
		return nil, fmt.Errorf("config supplied is nil")
	}

	ks := KubeSession{currentConfig: config}

	clientset, err := kubernetes.NewForConfig(ks.currentConfig)
	if err != nil {
		clarkezoneLog.Errorf("unable to create new clientset for config:%v", err)
		return nil, err
	}
	ks.currentClientset = clientset
	ks.jobnotifiers = make(map[string]jobnotifier)
	return &ks, nil
}

// CreatePersistentVolumeClaim creates a new persistentvolumeclaim
func (ks *KubeSession) CreatePersistentVolumeClaim(name string, namespace string) error {
	clarkezoneLog.Debugf("CreateVolume() called with name:%v namespace:%v", name, namespace)
	_, err := CreatePersistentVolumeClaim(ks.currentClientset, name, namespace)
	return err
}

// CreateNamespace creates a new namespace
func (ks *KubeSession) CreateNamespace(namespace string) error {
	clarkezoneLog.Debugf("CreateNamespace() called with namespace:%v", namespace)
	_, err := CreateNamespace(ks.currentClientset, namespace)
	return err
}

// DeleteNamespace deletes a namespace
func (ks *KubeSession) DeleteNamespace(namespace string) error {
	clarkezoneLog.Debugf("DeleteNamespace() called with namespace:%v", namespace)
	err := DeleteNamespace(ks.currentClientset, namespace)
	return err
}

// CreateJob makes a new job
func (ks *KubeSession) CreateJob(name string, namespace string,
	image string, command []string, args []string, notifier jobnotifier,
	autoDelete bool, mountlist []PVClaimMountRef) (*batchv1.Job, error) {
	clarkezoneLog.Debugf("CreateJob() called with name %v, namespace:%v,"+
		"image:%v, command:%v, args:%v, notifier:%v, autodelete:%v, pvlist:%v",
		name, namespace, image, command, args, notifier, autoDelete, mountlist)
	// TODO: if job exists, delete it
	job, err := CreateJob(ks.currentClientset, name, namespace,
		image, command, args, true, autoDelete, mountlist)
	if err != nil {
		return nil, err
	}
	if notifier != nil {
		ks.jobnotifiers[job.Name] = notifier
	}
	return job, nil
}

// DeleteJob deletes a job
func (ks *KubeSession) DeleteJob(name string, namespace string) error {
	clarkezoneLog.Debugf("DeleteJob() called with name:%v namespace:%v", name, namespace)
	return DeleteJob(ks.currentClientset, name, namespace)
}

// StartWatchers starts a goroutine that causes notifications to fire
func (ks *KubeSession) StartWatchers(namespace string) bool {
	clarkezoneLog.Debugf("startWatchers called with incluster:%v", namespace)
	// We will create an informer that writes added pods to a channel.
	var info informers.SharedInformerFactory
	if namespace == "" {
		// when watching in global scope, we need clusterrole / clusterrolebinding not role / rolebinding in the rbac setup
		info = informers.NewSharedInformerFactory(ks.currentClientset, 0)
	} else {
		info = informers.NewSharedInformerFactoryWithOptions(ks.currentClientset, 0, informers.WithNamespace(namespace))
	}
	podInformer := info.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			clarkezoneLog.Debugf("pod added: %s/%s", pod.Namespace, pod.Name)
			//	pods <- pod
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			clarkezoneLog.Debugf("pod deleted: %s/%s", pod.Namespace, pod.Name)
		},
	})

	jobInformer := info.Batch().V1().Jobs().Informer()

	jobInformer.AddEventHandler(ks.getJobEventHandlers())
	err := jobInformer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		// your code goes here
		clarkezoneLog.Errorf("watcher errorhandler caught error: %v", err.Error())
		ks.cancel()
	})
	if err != nil {
		clarkezoneLog.Errorf("Unable to set watcher error handler with %v, aborting", err)
		panic(err)
	}
	info.Start(ks.ctx.Done())

	// Ensuring that the informer goroutine have warmed up and called List before
	// we send any events to it.
	result := cache.WaitForCacheSync(ks.ctx.Done(), podInformer.HasSynced)
	result2 := cache.WaitForCacheSync(ks.ctx.Done(), jobInformer.HasSynced)
	if !result || !result2 {
		clarkezoneLog.Errorf("Waitforcachesync failed")
		return false
	}
	return true
}

func (ks *KubeSession) getJobEventHandlers() *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			job := obj.(*batchv1.Job)
			clarkezoneLog.Infof("Job added: %s/%s uid:%v", job.Namespace, job.Name, job.UID)
			if val, ok := ks.jobnotifiers[job.Name]; ok {
				val(job, Create)
			}
		},
		DeleteFunc: func(obj interface{}) {
			job := obj.(*batchv1.Job)
			clarkezoneLog.Infof("Job deleted: %s/%s uid:%v", job.Namespace, job.Name, job.UID)
			if val, ok := ks.jobnotifiers[job.Name]; ok {
				val(job, Delete)
				delete(ks.jobnotifiers, job.Name)
			}
		},
		UpdateFunc: func(oldobj interface{}, newobj interface{}) {
			oldjob := oldobj.(*batchv1.Job)
			newjob := newobj.(*batchv1.Job)
			clarkezoneLog.Infof("Job updated: %s/%s status:%v uid:%v", oldjob.Namespace, oldjob.Name, newjob.Status, newjob.UID)

			if val, ok := ks.jobnotifiers[newjob.Name]; ok {
				val(newjob, Update)
			}
		},
	}
}

// FindpvClaimByName searches for a persistentvolumeclaim by name
func (ks *KubeSession) FindpvClaimByName(pvname string, namespace string) (string, error) {
	return FindpvClaimByName(ks.currentClientset, pvname, namespace)
}

// CreatePvCMountReference creates a reference based on name and mountpoint
func (ks *KubeSession) CreatePvCMountReference(claimname string,
	mountPath string, readOnly bool) PVClaimMountRef {
	claim := PVClaimMountRef{}
	claim.PVClaimName = claimname
	claim.MountPath = mountPath
	claim.ReadOnly = readOnly
	return claim
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
func (ks *KubeSession) Close() {
	ks.cancel()
}
