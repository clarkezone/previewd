package cmd

import (
	"bytes"
	"io/ioutil"
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

	m.AssertExpectations(t)
	if err != nil {
		t.Fatal(err)
	}
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

func Test_CmdBaseInClusterMissingNsFail(t *testing.T) {
	m := &webhooklistenmockprovider{}
	cmd := getRunWebhookServerCmd(m)
	cmd.SetArgs([]string{"--targetrepo", "http://foo",
		"--localdir", "/tmp"})

	// TODO: should error out since missing namespace

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "" {
		t.Fatalf("expected \"%s\" got \"%s\"", "hi", string(out))
	}
}

func Test_CmdCloneOnly(t *testing.T) {
	// TODO: mock ensure no initial clone, no render, no webhook
	m := &webhooklistenmockprovider{}
	cmd := getRunWebhookServerCmd(m)
	cmd.SetArgs([]string{"--targetrepo", "http://foo",
		"--localdir", "/tmp", "--initialclone", "true",
		"--initialbuild", "false", "--webhooklisten", "false"})

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "" {
		t.Fatalf("expected \"%s\" got \"%s\"", "hi", string(out))
	}
}

func Test_CmdInitialRenderHookListen(t *testing.T) {
	// TODO: Mock ensure no initial clone, initial render, listen
	m := &webhooklistenmockprovider{}
	cmd := getRunWebhookServerCmd(m)
	cmd.SetArgs([]string{"--targetrepo", "http://foo",
		"--localdir", "/tmp", "--initialclone", "false"})

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	err := cmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "" {
		t.Fatalf("expected \"%s\" got \"%s\"", "hi", string(out))
	}
}

// GetTestConfig returns a local testing config for k8s
func GetTestConfig(t *testing.T) *rest.Config {
	p := internal.GetTestConfigPath(t)
	c, err := kubelayer.GetConfigOutofCluster(p)
	if err != nil {
		t.Fatalf("Couldn't get config %v", err)
	}
	return c
}
