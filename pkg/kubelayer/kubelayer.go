// Package kubelayer contains helpers for calling kube client
package kubelayer

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/batch/v1"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

const (
	//	jobttlsecondsafterfinished int32 = 1
	volumeName = "vol"
)

// PVClaimMountRef is a reference used to identify PVCs
type PVClaimMountRef struct {
	PVClaimName string
	MountPath   string
	ReadOnly    bool
}

// PingAPI tests if server is working
func PingAPI(clientset kubernetes.Interface) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	clarkezoneLog.Infof("There are %d pods in the cluster\n", len(pods.Items))
}

// CreateJob creates a new job resource
func CreateJob(clientset kubernetes.Interface,
	name string,
	namespace string, image string, command []string,
	args []string, always bool, autoDelete bool, mountlist []PVClaimMountRef) (*batchv1.Job, error) {
	clarkezoneLog.Debugf("CreateJob called with name %v namespace %v image %v command %v args %v always %v",
		name, namespace, image, command, args, always)

	var jobsClient v1.JobInterface
	if namespace == "" {
		jobsClient = clientset.BatchV1().Jobs(apiv1.NamespaceDefault)
	} else {
		jobsClient = clientset.BatchV1().Jobs(namespace)
	}

	// TODO hook up pull policy
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			// TODO: parameterize
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(1),
			// TTLSecondsAfterFinished: int32Ptr(jobttlsecondsafterfinished),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},

				Spec: apiv1.PodSpec{
					Volumes:       getVolumes(mountlist),
					Containers:    getContainers(name, image, command, args, mountlist),
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
		},
	}
	if command != nil {
		job.Spec.Template.Spec.Containers[0].Command = command
	}
	if args != nil {
		job.Spec.Template.Spec.Containers[0].Args = args
	}
	result, err := jobsClient.Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		clarkezoneLog.Errorf("CreateJob: jobsClient.Create failed %v", err)
		return nil, err
	}
	clarkezoneLog.Infof("Created job %v.\n", result.GetObjectMeta().GetName())
	return job, nil
}

// FindpvClaimByName searches for a PersistentVolumeClaim
func FindpvClaimByName(clientset kubernetes.Interface, pvname string, namespace string) (string, error) {
	var found string
	pvclient := clientset.CoreV1().PersistentVolumeClaims(namespace)
	pvlist, err := pvclient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, item := range pvlist.Items {
		if strings.Contains(item.ObjectMeta.Name, pvname) {
			found = item.ObjectMeta.Name
			break
		}
	}
	return found, nil
}

// getContainers returns containers based on name, command etc
func getContainers(name string, image string, command []string,
	args []string, mountlist []PVClaimMountRef) []apiv1.Container {
	containerList := []apiv1.Container{}
	volumeMountList := []apiv1.VolumeMount{}

	for i, mountitem := range mountlist {
		volumeMountList = append(volumeMountList, apiv1.VolumeMount{
			Name:      fmt.Sprintf("%v%v", volumeName, i),
			ReadOnly:  mountitem.ReadOnly,
			MountPath: mountitem.MountPath,
		},
		)
	}
	container := apiv1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: "Always",
		VolumeMounts:    volumeMountList,
	}

	if command != nil {
		container.Command = command
	}

	if args != nil {
		container.Args = args
	}

	containerList = append(containerList, container)

	return containerList
}

// getVolumes returns volumes based on mount refs
func getVolumes(mountlist []PVClaimMountRef) []apiv1.Volume {
	volumelist := []apiv1.Volume{}

	for i, mountitem := range mountlist {
		volumelist = append(volumelist, apiv1.Volume{
			Name: fmt.Sprintf("%v%v", volumeName, i),
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: mountitem.PVClaimName,
					ReadOnly:  mountitem.ReadOnly,
				},
			},
		})
	}

	return volumelist
}

// DeleteJob deletes an existing job resource
func DeleteJob(clientset kubernetes.Interface, name string, namespace string) error {
	var jobsClient v1.JobInterface
	if namespace == "" {
		jobsClient = clientset.BatchV1().Jobs(apiv1.NamespaceDefault)
	} else {
		jobsClient = clientset.BatchV1().Jobs(namespace)
	}
	meta := metav1.DeleteOptions{
		TypeMeta:           metav1.TypeMeta{},
		GracePeriodSeconds: new(int64),
		Preconditions:      &metav1.Preconditions{},
	}
	return jobsClient.Delete(context.TODO(), name, meta)
}

func int32Ptr(i int32) *int32 { return &i }
