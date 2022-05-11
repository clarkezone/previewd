// Package kubelayer contains helpers for calling kube client
package kubelayer

import (
	"context"
	"strings"

	"k8s.io/client-go/kubernetes"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

const (
	jobttlsecondsafterfinished int32 = 1
)

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
	args []string, always bool) (*batchv1.Job, error) {
	// TODO use default namespace if empty
	// TODO switch tests to call with empty
	// FIX

	clarkezoneLog.Debugf("CreateJob called with name %v namespace %v image %v command %v args %v always %v",
		name, namespace, image, command, args, always)

	jobsClient := clientset.BatchV1().Jobs(namespace)

	sourcename, rendername, err := findpvnames(clientset, namespace)
	clarkezoneLog.Debugf("Got volume names sourcename %v rendername %v", sourcename, rendername)

	if err != nil {
		clarkezoneLog.Errorf("CreateJob: findpvnames failed with %v", err)
		return nil, err
	}

	// TODO hook up pull policy
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			// TODO: parameterize
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            int32Ptr(1),
			TTLSecondsAfterFinished: int32Ptr(jobttlsecondsafterfinished),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},

				Spec: apiv1.PodSpec{
					//Volumes:       getVolumes(sourcename, rendername),
					Containers:    getContainers(name, image),
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

func getVolumes(sourcename string, rendername string) []apiv1.Volume {
	return []apiv1.Volume{
		{
			Name: "blogsource",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: sourcename,
					ReadOnly:  true,
				},
			},
		},
		{
			Name: "blogrender",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: rendername,
				},
			},
		},
	}
}

func getContainers(name string, image string) []apiv1.Container {
	return []apiv1.Container{
		{
			Name:            name,
			Image:           image,
			ImagePullPolicy: "Always",
			//TODO: command and args optional
			//Command:         command,
			//Args:            args,
			//			VolumeMounts: []apiv1.VolumeMount{
			//				{
			//					Name:      "blogsource",
			//					ReadOnly:  true,
			//					MountPath: "/src",
			//				},
			//				{
			//					Name:      "blogrender",
			//					ReadOnly:  false,
			//					MountPath: "/site",
			//				},
			//			},
		},
	}
}

func findpvnames(clientset kubernetes.Interface, namespace string) (string, string, error) {
	var sourcename string
	var rendername string

	pvclient := clientset.CoreV1().PersistentVolumeClaims(namespace)
	pvlist, err := pvclient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", "", err
	}
	for _, item := range pvlist.Items {
		if strings.Contains(item.ObjectMeta.Name, "render") {
			rendername = item.ObjectMeta.Name
		}
		if strings.Contains(item.ObjectMeta.Name, "source") {
			sourcename = item.ObjectMeta.Name
		}
	}
	return sourcename, rendername, nil
}

// DeleteJob deletes an existing job resource
func DeleteJob(clientset kubernetes.Interface, name string) error {
	jobsClient := clientset.BatchV1().Jobs(apiv1.NamespaceDefault)
	meta := metav1.DeleteOptions{
		TypeMeta:           metav1.TypeMeta{},
		GracePeriodSeconds: new(int64),
		Preconditions:      &metav1.Preconditions{},
	}
	return jobsClient.Delete(context.TODO(), name, meta)
}

func int32Ptr(i int32) *int32 { return &i }
