package host

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// manifestEmitter is a module whose whole purpose is to be run with the manifest
// flag: a small binary this test builds and runs, so it exercises the real
// os.Args path and the real os.Exit rather than a refactored inner function.
type manifestEmitter struct{}

func (manifestEmitter) Manifest() v1.Manifest {
	return v1.Manifest{
		ID:       "emitter",
		Version:  "v1.2.3",
		Name:     "Emitter",
		Provides: []v1.Role{v1.RoleSearch, v1.RoleMetadata},
	}
}

func (manifestEmitter) Import(_ context.Context, _ v1.ContentService, _ v1.ImportRequest) (v1.ImportResult, error) {
	return v1.ImportResult{}, nil
}

// This test process, re-executed with a sentinel env var, becomes the module:
// it serves the emitter, so `--mosaic-manifest` prints its manifest and exits.
// Running the real binary is the point — the flag is parsed from os.Args and the
// process exits, neither of which a direct function call reaches.
func TestMain(m *testing.M) {
	if os.Getenv("HOST_TEST_EMIT") == "1" {
		Serve(manifestEmitter{})
		return
	}
	os.Exit(m.Run())
}

func TestManifestFlagPrintsAndExits(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("locating test binary: %v", err)
	}
	cmd := exec.Command(exe, ManifestFlag)
	cmd.Env = append(os.Environ(), "HOST_TEST_EMIT=1")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("running with %s: %v", ManifestFlag, err)
	}

	var doc struct {
		ID       string   `json:"id"`
		Version  string   `json:"version"`
		Name     string   `json:"name"`
		Provides []string `json:"provides"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("manifest output is not JSON: %v\n%s", err, out)
	}
	if doc.ID != "emitter" || doc.Version != "v1.2.3" || doc.Name != "Emitter" {
		t.Errorf("identity: got %+v", doc)
	}
	if strings.Join(doc.Provides, ",") != "search,metadata" {
		t.Errorf("provides: got %v, want [search metadata]", doc.Provides)
	}
}
