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

// nolint
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
	clarkezoneLog.Debugf("== initial build with '%v'", a)
	p.Called(a)
	return nil
}

func (p *webhooklistenmockprovider) webhookListen() {
	clarkezoneLog.Debugf("webhookListen")
	p.Called()
}

func (p *webhooklistenmockprovider) WaitForInterupt() error {
	clarkezoneLog.Debugf("waitForInterupt")
	p.Called()
	return nil
}

func (*webhooklistenmockprovider) needInitialization() bool {
	return false
}

func Test_CmdBase(t *testing.T) {
	clarkezoneLog.Debugf("Test_cmdBase Start ============================================================== ")
	m := &webhooklistenmockprovider{}
	m.On("initialClone", "http://foo", "").Return(nil)
	m.On("initialBuild", "").Return(nil)
	m.On("webhookListen").Return()

	cmd := getRunWebhookServerCmd(m)
	cmd.SetArgs([]string{"--targetrepo", "http://foo",
		"--localdir", "/tmp", "--kubeconfigpath", internal.GetTestConfigPath(t)})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	m.AssertExpectations(t)
	clarkezoneLog.Debugf("Test_cmdBase END ============================================================== ")
}

func Test_CmdInitialRenderHookListen(t *testing.T) {
	clarkezoneLog.Debugf("Test_CmdInitialRenderHookListen Start ===============================================")
	mo := new(webhooklistenmockprovider)
	mo.On("initialBuild", "").Return(nil)
	mo.On("webhookListen")
	cmd := getRunWebhookServerCmd(mo)
	cmd.SetArgs([]string{"--targetrepo=http://bar",
		"--localdir=/tmp", "--initialclone=false"})

	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	mo.AssertExpectations(t)
	clarkezoneLog.Debugf("Test_CmdInitialRenderHookListen END ============================================ ")
}

func Test_CmdCloneOnly(t *testing.T) {
	clarkezoneLog.Debugf("Test_CmdCloneOnly Start ===============================================")
	m := &webhooklistenmockprovider{}
	m.On("initialClone", "http://foo", "")
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
	clarkezoneLog.Debugf("Test_CmdCloneOnly END ===============================================")
}

// This test runs on it's own but not a part of suite
// Some weird memory corruption or race condition that I can't figure out
// Disabling for now
// func Test_CmdBaseInClusterDefaultFail(t *testing.T) {
// 	clarkezoneLog.Debugf("Test_CmdBaseInClusterDefaultFail START ===============================================")
// 	m := &webhooklistenmockprovider{}
// 	m.On("initialClone", "http://baz", "").Return(nil)
// 	m.On("initialBuild", testNamespace).Return(nil)
// 	cmd := getRunWebhookServerCmd(m)
// 	cmd.SetArgs([]string{"--targetrepo=http://baz",
// 		"--localdir=/tmp", "--namespace=testns", "--kubeconfigpath="})
//
// 	// Simulate running in cluster
// 	internal.KubeConfigPath = ""
// 	err := cmd.Execute()
// 	if err == nil {
// 		t.Fatal("We should have an error for not running in cluster")
// 	}
// 	clarkezoneLog.Debugf("Test_CmdBaseInClusterDefaultFail END ===============================================")
// }

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
