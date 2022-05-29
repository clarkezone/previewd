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

// JobNotifier is a function prototype for notifications for job state changes
type JobNotifier func(*batchv1.Job, ResourseStateType)

// NamespaceNotifier is a function prototype for notifications for job state changes
type NamespaceNotifier func(*corev1.Namespace, ResourseStateType)

// KubeSession is a session to a k8s cluster
type KubeSession struct {
	currentConfig      *rest.Config
	currentClientset   kubernetes.Interface
	ctx                context.Context
	cancel             context.CancelFunc
	jobnotifiers       map[string]JobNotifier
	namespacenotifiers map[string]NamespaceNotifier
}

// Newkubesession creates a new kubesession from a config
func Newkubesession(config *rest.Config) (*KubeSession, error) {
	clarkezoneLog.Debugf("KubeSession: Newkubesession() called with incluster:%v, namespace:%v", config)
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
	ctx, cancel := context.WithCancel(context.Background())
	ks.ctx = ctx
	ks.cancel = cancel
	ks.jobnotifiers = make(map[string]JobNotifier)
	ks.namespacenotifiers = make(map[string]NamespaceNotifier)
	return &ks, nil
}

// CreatePersistentVolumeClaim creates a new persistentvolumeclaim
func (ks *KubeSession) CreatePersistentVolumeClaim(name string, namespace string) error {
	clarkezoneLog.Debugf("KubeSession: CreateVolume() called with name:%v namespace:%v", name, namespace)
	_, err := CreatePersistentVolumeClaim(ks.currentClientset, name, namespace)
	return err
}

// CreateNamespace creates a new namespace
func (ks *KubeSession) CreateNamespace(namespace string, notifier NamespaceNotifier) error {
	clarkezoneLog.Debugf("KubeSession: CreateNamespace() called with namespace:%v", namespace)
	if notifier != nil {
		ks.namespacenotifiers[namespace] = notifier
	}
	_, err := CreateNamespace(ks.currentClientset, namespace)
	return err
}

// GetNamespace creates a new namespace
func (ks *KubeSession) GetNamespace(namespace string) (*corev1.Namespace, error) {
	clarkezoneLog.Debugf("KubeSession: GetNamespace() called with namespace:%v", namespace)
	ns, err := GetNamespace(ks.currentClientset, namespace)
	return ns, err
}

// DeleteNamespace deletes a namespace
func (ks *KubeSession) DeleteNamespace(namespace string, notifier NamespaceNotifier) error {
	clarkezoneLog.Debugf("KubeSession: DeleteNamespace() called with namespace:%v", namespace)
	if notifier != nil {
		ks.namespacenotifiers[namespace] = notifier
	}
	err := DeleteNamespace(ks.currentClientset, namespace)
	return err
}

// CreateJob makes a new job
func (ks *KubeSession) CreateJob(name string, namespace string,
	image string, command []string, args []string, notifier JobNotifier,
	autoDelete bool, mountlist []PVClaimMountRef) (*batchv1.Job, error) {
	clarkezoneLog.Debugf("KubeSession: CreateJob() called with name %v, namespace:%v,"+
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
	clarkezoneLog.Debugf("KubeSession: DeleteJob() called with name:%v namespace:%v", name, namespace)
	return DeleteJob(ks.currentClientset, name, namespace)
}

// StartWatchers starts a goroutine that causes notifications to fire
func (ks *KubeSession) StartWatchers(namespace string, enablenamespacewatcher bool) error {
	clarkezoneLog.Debugf("Kubesession: startWatchers called with namesapce:%v", namespace)
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

	var namespaceInformer cache.SharedIndexInformer
	if enablenamespacewatcher {
		namespaceInformer = info.Core().V1().Namespaces().Informer()
		namespaceInformer.AddEventHandler(ks.getNamespaceHandlers())
	}

	// Handle errors
	err := jobInformer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		clarkezoneLog.Errorf(" KubeSession: StartWatchers() watcher errorhandler caught error: %v", err.Error())
		ks.cancel()
	})
	if err != nil {
		clarkezoneLog.Errorf(" KubeSession: StartWatchers() Unable to set watcher error handler with %v, aborting", err)
		return err
	}

	// Start informers
	info.Start(ks.ctx.Done())

	// Ensuring that the informer goroutine have warmed up and called List before
	// we send any events to it.
	result := cache.WaitForCacheSync(ks.ctx.Done(), podInformer.HasSynced)
	result2 := cache.WaitForCacheSync(ks.ctx.Done(), jobInformer.HasSynced)
	var result3 bool
	if enablenamespacewatcher {
		result3 = cache.WaitForCacheSync(ks.ctx.Done(), namespaceInformer.HasSynced)
	}
	if !result || !result2 || !result3 {
		err := fmt.Errorf(" kubesession: waitforcachesync failed")
		clarkezoneLog.Errorf(" kubesession: startWatchers: failed: %v", err)
		return err
	}
	return nil
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

func (ks *KubeSession) getNamespaceHandlers() *cache.ResourceEventHandlerFuncs {
	return &cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ns := obj.(*corev1.Namespace)
			clarkezoneLog.Debugf("Namespace added: %s", ns.Name)
			if val, ok := ks.namespacenotifiers[ns.Name]; ok {
				val(ns, Update)
			}
		},
		DeleteFunc: func(obj interface{}) {
			ns := obj.(*corev1.Namespace)
			clarkezoneLog.Debugf("Namespace deleted: %s", ns.Name)
			if val, ok := ks.namespacenotifiers[ns.Name]; ok {
				val(ns, Update)
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
	clarkezoneLog.Debugf("Kubesession: GetConfigIncluster() called with incluster")
	var config *rest.Config
	var err error
	config, err = rest.InClusterConfig()
	if err != nil {
		clarkezoneLog.Errorf("Kubesession: InClusterConfig() returned error %v", err)
	}
	return config, err
}

// GetConfigOutofCluster returns a config loaded from the supplied path
func GetConfigOutofCluster(kubepath string) (*rest.Config, error) {
	clarkezoneLog.Debugf("Kubesession: GetConfigOutofCluster() called with kubepath:%v", kubepath)
	var config *rest.Config
	var err error
	var kubeconfig = &kubepath
	// use the current context in kubeconfig
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		clarkezoneLog.Errorf("Kubesession: BuildConfigFromFlags() failed with %v", err)
	}
	return config, err
}

// Close cancels all jobmanager go routines
func (ks *KubeSession) Close() {
	ks.cancel()
}
