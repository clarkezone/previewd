package kubelayer

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestApi(t *testing.T) {
	t.Logf("TestApi")
	clientset := fake.NewSimpleClientset()
	PingAPI(clientset)
}

func TestCreateJobKubeLayer(t *testing.T) {
	t.Logf("TestCreateJobKubeLayer")
	clientset := fake.NewSimpleClientset()
	_, err := CreateJob(clientset, "testns", "", "", nil, nil, false, false, nil)
	if err != nil {
		t.Fatalf("Create job failed %v", err)
	}
}
