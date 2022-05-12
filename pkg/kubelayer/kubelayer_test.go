package kubelayer

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes/fake"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

// TestMain initizlie all tests
func TestMain(m *testing.M) {
	clarkezoneLog.Init(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func TestApi(t *testing.T) {
	t.Logf("TestApi")
	clientset := fake.NewSimpleClientset()
	PingAPI(clientset)
}

func TestCreateJobKubeLayer(t *testing.T) {
	t.Logf("TestCreateJobKubeLayer")
	clientset := fake.NewSimpleClientset()
	_, err := CreateJob(clientset, "testns", "", "", nil, nil, false, false)
	if err != nil {
		t.Fatalf("Create job failed %v", err)
	}
}
