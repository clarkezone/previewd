package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/clarkezone/previewd/internal"
	"github.com/clarkezone/previewd/pkg/config"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

// TestMain initizlie all tests
func TestMain(m *testing.M) {
	internal.SetupGitRoot()
	clarkezoneLog.Init(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func Test_ExecuteVersion(t *testing.T) {
	config.VersionString = "1"
	config.VersionHash = "A"
	cmd := getVersionCommand()
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
	if string(out) != "previewd version:1 hash:A\n" {
		t.Fatalf("expected \"%s\" got \"%s\"", "hi", string(out))
	}
}
