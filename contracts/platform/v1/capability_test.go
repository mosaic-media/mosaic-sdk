package v1_test

import (
	"context"
	"testing"

	v1 "github.com/mosaic-media/mosaic-sdk/contracts/platform/v1"
)

// stubCapability is the shape a Module implements: it holds only the SDK and
// drives ContentService with the Caller it is handed. It exists to prove the
// interface is satisfiable from outside the package with no Platform types.
type stubCapability struct {
	id       string
	imported string
}

func (c *stubCapability) Manifest() v1.Manifest {
	return v1.Manifest{ID: c.id, Version: "0.0.1", Name: "Stub"}
}

func (c *stubCapability) Import(ctx context.Context, svc v1.ContentService, caller v1.Caller, query string) (v1.ImportResult, error) {
	c.imported = query
	// A real capability would call svc here; the stub only records that it was
	// handed the surface and a caller, which is all the contract promises.
	_ = svc
	_ = caller
	return v1.ImportResult{WorkID: v1.NodeID("work-" + query), Items: 1}, nil
}

// TestCapabilityIsImplementableExternally checks the capability surface holds
// together: a value implementing it satisfies v1.Capability, and Import
// returns the result type by value.
func TestCapabilityIsImplementableExternally(t *testing.T) {
	var cap v1.Capability = &stubCapability{id: "stub"}

	if got := cap.Manifest().ID; got != "stub" {
		t.Fatalf("Manifest().ID = %q, want %q", got, "stub")
	}

	res, err := cap.Import(context.Background(), nil, v1.CallerFromSession("session-1"), "tt1254207")
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}
	if res.WorkID != v1.NodeID("work-tt1254207") {
		t.Fatalf("Import WorkID = %q, want %q", res.WorkID, "work-tt1254207")
	}
	if res.Items != 1 {
		t.Fatalf("Import Items = %d, want 1", res.Items)
	}
}
