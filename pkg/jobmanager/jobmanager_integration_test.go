//go:build integration
// +build integration

// open settings json or remote settings json
// {
//"go.buildFlags": [
//    "-tags=unit,integration"
//],
//"go.buildTags": "-tags=unit,integration",
//"go.testTags": "-tags=unit,integration"
// }

package jobmanager

import (
	"testing"

	"github.com/clarkezone/previewd/internal"
	kubelayer "github.com/clarkezone/previewd/pkg/kubelayer"
)

func TestCreateJobE2E(t *testing.T) {
	path := internal.GetTestConfigPath(t)
	config, err := kubelayer.GetConfigOutofCluster(path)
	if err != nil {
		t.Fatalf("Can't get config %v", err)
	}
	jm, err := Newjobmanager(config, "testns", true)
	if err != nil {
		t.Fatalf("Can't create jobmanager %v", err)
	}
	err = jm.AddJobtoQueue("alpinetest", testNamespace, "alpine", nil, nil,
		[]kubelayer.PVClaimMountRef{})
	if err != nil {
		t.Fatalf("Unable to create job %v", err)
	}

	panic("How do we know if job is complete")
}
