package host

import (
	"context"
	"errors"
	"testing"
	"time"

	goplugin "github.com/hashicorp/go-plugin"

	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// These tests run the real wire: a real gRPC connection, real protobuf
// serialization, and go-plugin's real broker for the callback direction. What
// they do not do is spawn a process — that is the probe module's job (ADR 0064's
// build order puts it after this).
//
// The distinction matters when reading a green result here: this proves the
// conversions, the dispatch and the bidirectional call graph. It does not prove
// the handshake or the Unix socket.

// ─── Test doubles ───────────────────────────────────────────────────────────

// stubCapability is a module. It fills one read role, which is deliberate:
// ADR 0064's step 1 is "a trivial in-repo module implementing one role", and a
// module that filled everything would not exercise the not-implemented path.
type stubCapability struct {
	manifest v1.Manifest

	// importFn lets a test decide what Import does, including failing.
	importFn func(ctx context.Context, svc v1.ContentService, req v1.ImportRequest) (v1.ImportResult, error)
}

func (s *stubCapability) Manifest() v1.Manifest { return s.manifest }

func (s *stubCapability) Import(ctx context.Context, svc v1.ContentService, req v1.ImportRequest) (v1.ImportResult, error) {
	return s.importFn(ctx, svc, req)
}

func (s *stubCapability) Search(_ context.Context, req v1.SearchRequest) (v1.SearchResponse, error) {
	return v1.SearchResponse{Results: []v1.SearchResult{{
		Ref:    v1.ContentRef{Provider: "stub", NativeID: req.Text, MediaType: v1.MediaMovie},
		Title:  "Result for " + req.Text,
		Year:   1982,
		Poster: "https://example.invalid/p.jpg",
	}}}, nil
}

// stubContent is the Platform's ContentService. Only the methods the tests
// exercise do anything; the rest satisfy the interface.
type stubContent struct {
	gotWork  v1.AddContentWorkCommand
	workErr  error
	callerIn string
}

func (s *stubContent) AddContentWork(_ context.Context, cmd v1.AddContentWorkCommand) (v1.AddContentWorkResult, error) {
	s.gotWork = cmd
	s.callerIn = cmd.Caller.Session
	if s.workErr != nil {
		return v1.AddContentWorkResult{}, s.workErr
	}
	return v1.AddContentWorkResult{Work: v1.Node{
		ID:        "node-1",
		WorkID:    "node-1",
		Kind:      v1.NodeWork,
		MediaType: cmd.MediaType,
		Title:     cmd.Title,
		Status:    v1.NodeActive,
		CreatedAt: time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC),
	}}, nil
}

func (s *stubContent) AddContentChild(context.Context, v1.AddContentChildCommand) (v1.AddContentChildResult, error) {
	return v1.AddContentChildResult{}, nil
}
func (s *stubContent) AttachContentPart(context.Context, v1.AttachContentPartCommand) (v1.AttachContentPartResult, error) {
	return v1.AttachContentPartResult{}, nil
}
func (s *stubContent) SetContentArtwork(context.Context, v1.SetContentArtworkCommand) (v1.SetContentArtworkResult, error) {
	return v1.SetContentArtworkResult{}, nil
}
func (s *stubContent) RelateContent(context.Context, v1.RelateContentCommand) (v1.RelateContentResult, error) {
	return v1.RelateContentResult{}, nil
}
func (s *stubContent) BindContentSource(context.Context, v1.BindContentSourceCommand) (v1.BindContentSourceResult, error) {
	return v1.BindContentSourceResult{}, nil
}
func (s *stubContent) ResolveContentBinding(context.Context, v1.ResolveContentBindingCommand) (v1.ResolveContentBindingResult, error) {
	return v1.ResolveContentBindingResult{}, nil
}
func (s *stubContent) SearchContent(context.Context, v1.SearchContentQuery) (v1.SearchContentResult, error) {
	return v1.SearchContentResult{}, nil
}
func (s *stubContent) FindContentByExternalID(context.Context, v1.FindContentByExternalIDQuery) (v1.FindContentByExternalIDResult, error) {
	return v1.FindContentByExternalIDResult{}, nil
}
func (s *stubContent) GetContentNode(context.Context, v1.GetContentNodeQuery) (v1.GetContentNodeResult, error) {
	return v1.GetContentNodeResult{}, nil
}
func (s *stubContent) ListContentParts(context.Context, v1.ListContentPartsQuery) (v1.ListContentPartsResult, error) {
	return v1.ListContentPartsResult{}, nil
}
func (s *stubContent) RecordPlaybackProgress(context.Context, v1.RecordPlaybackProgressCommand) (v1.RecordPlaybackProgressResult, error) {
	return v1.RecordPlaybackProgressResult{}, nil
}
func (s *stubContent) SetPlaybackFinished(context.Context, v1.SetPlaybackFinishedCommand) (v1.SetPlaybackFinishedResult, error) {
	return v1.SetPlaybackFinishedResult{}, nil
}
func (s *stubContent) GetPlaybackState(context.Context, v1.GetPlaybackStateQuery) (v1.GetPlaybackStateResult, error) {
	return v1.GetPlaybackStateResult{}, nil
}
func (s *stubContent) ListPlaybackStates(context.Context, v1.ListPlaybackStatesQuery) (v1.ListPlaybackStatesResult, error) {
	return v1.ListPlaybackStatesResult{}, nil
}
func (s *stubContent) ListInProgress(context.Context, v1.ListInProgressQuery) (v1.ListInProgressResult, error) {
	return v1.ListInProgressResult{}, nil
}

var _ v1.ContentService = (*stubContent)(nil)

// connect wires both sides over a real gRPC connection with a live broker.
func connect(t *testing.T, impl v1.Capability, content v1.ContentService, categoryOf CategoryFunc) v1.Capability {
	t.Helper()

	p := &Plugin{Impl: impl, Content: content, CategoryOf: categoryOf}
	client, server := goplugin.TestPluginGRPCConn(t, false, map[string]goplugin.Plugin{
		PluginName: p,
	})
	t.Cleanup(func() {
		_ = client.Close()
		server.Stop()
	})

	raw, err := client.Dispense(PluginName)
	if err != nil {
		t.Fatalf("dispense: %v", err)
	}
	cap, ok := raw.(v1.Capability)
	if !ok {
		t.Fatalf("dispensed %T, which does not implement v1.Capability", raw)
	}
	return cap
}

// ─── Tests ──────────────────────────────────────────────────────────────────

func TestManifestCrossesTheBoundary(t *testing.T) {
	want := v1.Manifest{
		ID:       "stub",
		Version:  "v1.2.3",
		Name:     "Stub Module",
		Provides: []v1.Role{v1.RoleSearch},
	}
	c := connect(t, &stubCapability{manifest: want}, &stubContent{}, nil)

	got := c.Manifest()
	if got.ID != want.ID || got.Version != want.Version || got.Name != want.Name {
		t.Errorf("manifest identity: got %+v, want %+v", got, want)
	}
	if len(got.Provides) != 1 || got.Provides[0] != v1.RoleSearch {
		t.Errorf("provides: got %v, want %v", got.Provides, want.Provides)
	}
}

// The registry holds a v1.Capability and cannot tell a proxy from a local
// struct — the property ADR 0064 is arranged around.
func TestProxySatisfiesTheCapabilityInterface(t *testing.T) {
	c := connect(t, &stubCapability{manifest: v1.Manifest{ID: "stub"}}, &stubContent{}, nil)
	if _, ok := c.(v1.Capability); !ok {
		t.Fatal("proxy does not satisfy v1.Capability")
	}
	// And every role, unconditionally — which is exactly why Roles() reads the
	// manifest instead of type-asserting.
	if _, ok := c.(v1.SearchProvider); !ok {
		t.Fatal("proxy does not satisfy v1.SearchProvider")
	}
}

func TestRolesComeFromTheManifestNotTypeAssertion(t *testing.T) {
	// The stub fills only Search. A type assertion for a role it does not fill
	// still succeeds against the proxy, so the manifest is the only honest
	// answer.
	c := connect(t, &stubCapability{
		manifest: v1.Manifest{ID: "stub", Provides: []v1.Role{v1.RoleSearch}},
	}, &stubContent{}, nil)

	if _, ok := c.(v1.PlaybackProvider); !ok {
		t.Fatal("precondition: the proxy is expected to satisfy every role interface")
	}
	roles := Roles(c)
	if len(roles) != 1 || roles[0] != v1.RoleSearch {
		t.Fatalf("Roles: got %v, want [search]", roles)
	}
}

// The bidirectional case: the Platform calls Import, and the module calls back
// into ContentService over the broker, within that invocation.
func TestImportCallsBackIntoContentService(t *testing.T) {
	content := &stubContent{}
	impl := &stubCapability{
		manifest: v1.Manifest{ID: "stub"},
		importFn: func(ctx context.Context, svc v1.ContentService, req v1.ImportRequest) (v1.ImportResult, error) {
			out, err := svc.AddContentWork(ctx, v1.AddContentWorkCommand{
				Caller:    req.Caller,
				MediaType: req.Ref.MediaType,
				Title:     "Blade Runner",
			})
			if err != nil {
				return v1.ImportResult{}, err
			}
			return v1.ImportResult{WorkID: out.Work.ID, Items: 1}, nil
		},
	}
	c := connect(t, impl, content, nil)

	got, err := c.Import(context.Background(), nil, v1.ImportRequest{
		Caller:   v1.CallerFromSession("handle-abc"),
		Ref:      v1.ContentRef{Provider: "stub", NativeID: "tt0083658", MediaType: v1.MediaMovie},
		Settings: []byte(`{"key":"value"}`),
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if got.WorkID != "node-1" || got.Items != 1 {
		t.Errorf("result: got %+v, want WorkID=node-1 Items=1", got)
	}

	// The callback arrived, carrying the Caller handle the Platform minted.
	if content.callerIn != "handle-abc" {
		t.Errorf("caller did not survive the callback: got %q, want %q", content.callerIn, "handle-abc")
	}
	if content.gotWork.Title != "Blade Runner" {
		t.Errorf("callback payload: got title %q", content.gotWork.Title)
	}
	if content.gotWork.MediaType != v1.MediaMovie {
		t.Errorf("media type did not survive: got %q", content.gotWork.MediaType)
	}
}

func TestSettingsCrossUninterpreted(t *testing.T) {
	var gotSettings []byte
	impl := &stubCapability{
		manifest: v1.Manifest{ID: "stub"},
		importFn: func(_ context.Context, _ v1.ContentService, req v1.ImportRequest) (v1.ImportResult, error) {
			gotSettings = req.Settings
			return v1.ImportResult{}, nil
		},
	}
	c := connect(t, impl, &stubContent{}, nil)

	// ADR 0021 stores module settings as opaque JSON, and ADR 0064 notes this
	// crossing unchanged is a small vindication of that.
	want := `{"addons":["https://example.invalid/manifest.json"],"nested":{"n":1}}`
	if _, err := c.Import(context.Background(), nil, v1.ImportRequest{
		Caller:   v1.CallerFromSession("h"),
		Settings: []byte(want),
	}); err != nil {
		t.Fatalf("import: %v", err)
	}
	if string(gotSettings) != want {
		t.Errorf("settings: got %q, want %q", gotSettings, want)
	}
}

// A module's error reaches the Platform with its message intact and no
// category — exactly as in process, where CategoryOf maps an uncategorised
// error to Internal.
func TestModuleErrorCrossesWithoutACategory(t *testing.T) {
	impl := &stubCapability{
		manifest: v1.Manifest{ID: "stub"},
		importFn: func(context.Context, v1.ContentService, v1.ImportRequest) (v1.ImportResult, error) {
			return v1.ImportResult{}, errors.New("upstream returned 502")
		},
	}
	c := connect(t, impl, &stubContent{}, nil)

	_, err := c.Import(context.Background(), nil, v1.ImportRequest{Caller: v1.CallerFromSession("h")})
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := err.Error(); got == "" {
		t.Fatal("error message was lost")
	}
	if cat := CategoryOfWireError(err); cat != "" {
		t.Errorf("a module error should carry no category, got %q", cat)
	}
}

// A Platform error reaching the module keeps its category, supplied by the
// injected CategoryFunc since this package cannot read the Platform's own
// vocabulary.
func TestPlatformErrorCategorySurvivesTheCallback(t *testing.T) {
	content := &stubContent{workErr: errors.New("node already exists")}
	var seen error
	impl := &stubCapability{
		manifest: v1.Manifest{ID: "stub"},
		importFn: func(ctx context.Context, svc v1.ContentService, req v1.ImportRequest) (v1.ImportResult, error) {
			_, err := svc.AddContentWork(ctx, v1.AddContentWorkCommand{Caller: req.Caller})
			seen = err
			return v1.ImportResult{}, nil
		},
	}
	categoryOf := func(error) string { return "conflict" }
	c := connect(t, impl, content, categoryOf)

	if _, err := c.Import(context.Background(), nil, v1.ImportRequest{Caller: v1.CallerFromSession("h")}); err != nil {
		t.Fatalf("import: %v", err)
	}
	if seen == nil {
		t.Fatal("the module saw no error from the callback")
	}
	if got := CategoryOfWireError(seen); got != "conflict" {
		t.Errorf("category: got %q, want conflict", got)
	}
	if seen.Error() != "node already exists" {
		t.Errorf("message: got %q", seen.Error())
	}
}

func TestSearchRoleRoundTrips(t *testing.T) {
	c := connect(t, &stubCapability{
		manifest: v1.Manifest{ID: "stub", Provides: []v1.Role{v1.RoleSearch}},
	}, &stubContent{}, nil)

	sp, ok := c.(v1.SearchProvider)
	if !ok {
		t.Fatal("proxy is not a SearchProvider")
	}
	out, err := sp.Search(context.Background(), v1.SearchRequest{
		Caller: v1.CallerFromSession("h"),
		Text:   "blade runner",
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(out.Results) != 1 {
		t.Fatalf("results: got %d, want 1", len(out.Results))
	}
	got := out.Results[0]
	if got.Title != "Result for blade runner" || got.Year != 1982 {
		t.Errorf("result: got %+v", got)
	}
	if got.Ref.MediaType != v1.MediaMovie || got.Ref.NativeID != "blade runner" {
		t.Errorf("ref did not survive: got %+v", got.Ref)
	}
}

// A role the module does not fill is refused rather than silently returning
// nothing, so a Platform bug shows up as an error instead of an empty result.
func TestUnfilledRoleIsRefused(t *testing.T) {
	c := connect(t, &stubCapability{manifest: v1.Manifest{ID: "stub"}}, &stubContent{}, nil)

	ap, ok := c.(v1.ArtworkProvider)
	if !ok {
		t.Fatal("proxy is not an ArtworkProvider")
	}
	if _, err := ap.Artwork(context.Background(), v1.ArtworkRequest{
		Caller: v1.CallerFromSession("h"),
	}); err == nil {
		t.Fatal("expected an error for a role the module does not fill")
	}
}
