package v1_test

import (
	"context"
	"testing"

	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// stubPlayer is a consumer module: it fills RolePlayback and *no source role*.
// That is the point of testing it separately from stubProvider — a consumer is
// not a source that happens to also play, and nothing in the contract should
// require it to import content it did not source (ADR 0045).
type stubPlayer struct{}

func (stubPlayer) Manifest() v1.Manifest {
	return v1.Manifest{
		ID: "stub-playback", Version: "0.0.1", Name: "Stub Playback",
		Provides: []v1.Role{v1.RolePlayback},
	}
}

// Import is the base Capability's one write verb. A consumer materialises
// nothing, so it reports an empty result rather than pretending to import.
func (stubPlayer) Import(_ context.Context, _ v1.ContentService, _ v1.ImportRequest) (v1.ImportResult, error) {
	return v1.ImportResult{}, nil
}

func (stubPlayer) Resolve(_ context.Context, req v1.PlaybackRequest) (v1.PlaybackResolution, error) {
	return v1.PlaybackResolution{
		Kind:    v1.PlaybackDirect,
		URL:     req.Part.Location.Ref,
		Headers: map[string]string{"Authorization": "Bearer stub"},
	}, nil
}

// TestPlaybackProviderImplementableExternally checks that the consumer surface
// is satisfiable from outside the package with no Platform types, that a module
// can fill RolePlayback without filling a single source role, and that a
// resolution round-trips the part location it was handed.
func TestPlaybackProviderImplementableExternally(t *testing.T) {
	var (
		cap  v1.Capability       = stubPlayer{}
		play v1.PlaybackProvider = stubPlayer{}
	)

	provides := cap.Manifest().Provides
	if len(provides) != 1 || provides[0] != v1.RolePlayback {
		t.Fatalf("Manifest().Provides = %v, want exactly [%q]", provides, v1.RolePlayback)
	}

	part := v1.Part{
		Role:     v1.PartEdition,
		Location: v1.MediaLocation{Scheme: v1.RemoteLocation, Provider: "stremio-addons", Ref: "https://cdn.example/movie.mp4"},
	}

	res, err := play.Resolve(context.Background(), v1.PlaybackRequest{Part: part})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if res.Kind != v1.PlaybackDirect {
		t.Fatalf("Kind = %q, want %q", res.Kind, v1.PlaybackDirect)
	}
	if res.URL != part.Location.Ref {
		t.Fatalf("URL = %q, want %q", res.URL, part.Location.Ref)
	}
	if res.Headers["Authorization"] == "" {
		t.Fatal("Headers lost the authorization the provider set")
	}
}

// TestPlaybackRoleIsNotASourceRole guards the distinction the role vocabulary
// now carries: a consumer must not be resolvable as any source provider, or the
// registry's role check (and ADR 0036's gate) would read a player as a source.
func TestPlaybackRoleIsNotASourceRole(t *testing.T) {
	var c v1.Capability = stubPlayer{}

	if _, ok := c.(v1.StreamProvider); ok {
		t.Error("a playback consumer must not satisfy StreamProvider")
	}
	if _, ok := c.(v1.MetadataProvider); ok {
		t.Error("a playback consumer must not satisfy MetadataProvider")
	}
	if _, ok := c.(v1.SearchProvider); ok {
		t.Error("a playback consumer must not satisfy SearchProvider")
	}
	if _, ok := c.(v1.CatalogProvider); ok {
		t.Error("a playback consumer must not satisfy CatalogProvider")
	}
}
