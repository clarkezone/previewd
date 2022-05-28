package cmd

import (
	"testing"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/kubelayer"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/rest"
)

func init() {
	clarkezoneLog.Init(logrus.DebugLevel)
}

const (
	testNamespace = "testns"
)

type webhooklistenmockprovider struct {
	mock.Mock
}

func (p *webhooklistenmockprovider) initialClone(a string, b string) error {
	p.Called(a, b)
	return nil
}

func (p *webhooklistenmockprovider) initialBuild(a string) error {
	p.Called(a)
	return nil
}

func (p *webhooklistenmockprovider) webhookListen() {
	p.Called()
}

func (*webhooklistenmockprovider) needInitialization() bool {
	return false
}

func Test_CmdBase(t *testing.T) {
	m := &webhooklistenmockprovider{}
	m.On("initialClone", "http://foo", "main")
	m.On("initialBuild", "testns")
	m.On("webhookListen")

	cmd := getRunWebhookServerCmd(m)
	cmd.SetArgs([]string{"--targetrepo", "http://foo",
		"--localdir", "/tmp", "--kubeconfigpath", internal.GetTestConfigPath(t), "--namespace", testNamespace})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	m.AssertExpectations(t)
}

func Test_CmdBaseInClusterDefaultFail(t *testing.T) {
	m := &webhooklistenmockprovider{}
	cmd := getRunWebhookServerCmd(m)
	cmd.SetArgs([]string{"--targetrepo", "http://foo",
		"--localdir", "/tmp", "--namespace", "testns"})

	// Simulate running in cluster
	internal.KubeConfigPath = ""
	err := cmd.Execute()
	if err == nil {
		t.Fatal("We should have an error for not running in cluster")
	}
}

func Test_CmdCloneOnly(t *testing.T) {
	m := &webhooklistenmockprovider{}
	m.On("initialClone", "http://foo", "main")
	cmd := getRunWebhookServerCmd(m)

	// Note that for bool flags they cannot be passed in without =
	// or as separate strings for key and value
	cmd.SetArgs([]string{"--targetrepo=http://foo",
		"--localdir=/tmp", "--initialclone=true",
		"--initialbuild=false", "--webhooklisten=false"})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	m.AssertExpectations(t)
}

func Test_CmdInitialRenderHookListen(t *testing.T) {
	m := &webhooklistenmockprovider{}
	m.On("initialBuild", "")
	m.On("webhookListen")
	cmd := getRunWebhookServerCmd(m)
	cmd.SetArgs([]string{"--targetrepo=http://foo",
		"--localdir=/tmp", "--initialclone=false"})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	m.AssertExpectations(t)
}

// TODO: confirm no args error

// TODO: what are minimal commandline args for this to work usefully?

// GetTestConfig returns a local testing config for k8s
func GetTestConfig(t *testing.T) *rest.Config {
	p := internal.GetTestConfigPath(t)
	c, err := kubelayer.GetConfigOutofCluster(p)
	if err != nil {
		t.Fatalf("Couldn't get config %v", err)
	}
	return c
}
