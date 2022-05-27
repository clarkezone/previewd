package cmd

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/kubelayer"
	"k8s.io/client-go/rest"
)

const (
	testNamespace = "testns"
)

type webhooklistenmockprovider struct {
}

func (p webhooklistenmockprovider) initialClone(string, string) error {
	return nil
}

func (p webhooklistenmockprovider) initialBuild(string) error {
	return nil
}

func (p webhooklistenmockprovider) webhookListen() {

}

func Test_CmdBase(t *testing.T) {
	// ensure clone, render, webhook
	m := &webhooklistenmockprovider{}
	cmd := getRunWebhookServerCmd(m)
	cmd.SetArgs([]string{"--targetrepo", "http://foo",
		"--localdir", "/tmp"})

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
	// ensure no initial clone, no render, no webhook
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
	// ensure no initial clone, initial render, listen
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
