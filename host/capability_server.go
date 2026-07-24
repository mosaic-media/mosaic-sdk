package host

import (
	"context"
	"sync"

	goplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	modulev1 "github.com/mosaic-media/contracts/gen/mosaic/module/v1"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// capabilityServer runs in the module process and turns the wire back into
// calls on the author's plain Go [v1.Capability].
//
// Every role method type-asserts the implementation to the matching provider
// interface and reports Unimplemented when the module does not fill it. That is
// a backstop rather than the mechanism: the handshake already refuses a module
// whose manifest declares a role it does not serve, so reaching one of these
// returns means the Platform called a role the manifest never claimed.
type capabilityServer struct {
	modulev1.UnimplementedCapabilityServiceServer

	impl   v1.Capability
	broker *goplugin.GRPCBroker

	// The callback connection is dialled once, on first use, and shared. Dialling
	// per call would open a stream per ContentService write, which for a tree
	// import is one per node.
	dialOnce  sync.Once
	callbacks *grpc.ClientConn
	dialErr   error
}

// platform returns the Platform-side services, dialling the callback stream on
// first use. The returned ContentService and Telemetry are what the module's
// own code sees through the SDK interfaces.
func (s *capabilityServer) platform() (v1.ContentService, v1.Telemetry, error) {
	s.dialOnce.Do(func() {
		s.callbacks, s.dialErr = s.broker.Dial(callbackBrokerID)
	})
	if s.dialErr != nil {
		return nil, nil, s.dialErr
	}
	return &contentClient{client: modulev1.NewContentServiceClient(s.callbacks)},
		&telemetryClient{client: modulev1.NewTelemetryServiceClient(s.callbacks)},
		nil
}

func unimplemented(role string) error {
	return status.Errorf(codes.Unimplemented, "module does not fill the %q role", role)
}

func (s *capabilityServer) GetManifest(_ context.Context, _ *modulev1.ManifestRequest) (*modulev1.ManifestResponse, error) {
	return &modulev1.ManifestResponse{Manifest: manifestToWire(s.impl.Manifest())}, nil
}

// Import is the one write verb, and the only method whose implementation is
// handed the Platform's ContentService: a read role never writes (ADR 0027), so
// none of the others take one.
func (s *capabilityServer) Import(ctx context.Context, req *modulev1.ImportRequest) (*modulev1.ImportResponse, error) {
	content, telemetry, err := s.platform()
	if err != nil {
		return nil, errorToWire(err, nil)
	}

	// The module reaches Telemetry ambiently, off the context the Platform
	// handed it (ADR 0059) — so the context carries it here exactly as it would
	// in process.
	ctx = v1.WithTelemetry(ctx, telemetry)

	result, err := s.impl.Import(ctx, content, v1.ImportRequest{
		Caller:   callerFromWire(req.GetCaller()),
		Ref:      refFromWire(req.GetRef()),
		Settings: req.GetSettings(),
	})
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	return &modulev1.ImportResponse{
		WorkId:       string(result.WorkID),
		AlreadyKnown: result.AlreadyKnown,
		Containers:   int32(result.Containers),
		Items:        int32(result.Items),
		Parts:        int32(result.Parts),
	}, nil
}

func (s *capabilityServer) withTelemetry(ctx context.Context) context.Context {
	_, telemetry, err := s.platform()
	if err != nil {
		return ctx
	}
	return v1.WithTelemetry(ctx, telemetry)
}

func (s *capabilityServer) Metadata(ctx context.Context, req *modulev1.MetadataRequest) (*modulev1.MetadataResponse, error) {
	p, ok := s.impl.(v1.MetadataProvider)
	if !ok {
		return nil, unimplemented(string(v1.RoleMetadata))
	}
	out, err := p.Metadata(s.withTelemetry(ctx), v1.MetadataRequest{
		Caller:   callerFromWire(req.GetCaller()),
		Settings: req.GetSettings(),
		Ref:      refFromWire(req.GetRef()),
	})
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	return &modulev1.MetadataResponse{Metadata: metadataToWire(out)}, nil
}

func (s *capabilityServer) Search(ctx context.Context, req *modulev1.SearchRequest) (*modulev1.SearchResponse, error) {
	p, ok := s.impl.(v1.SearchProvider)
	if !ok {
		return nil, unimplemented(string(v1.RoleSearch))
	}
	out, err := p.Search(s.withTelemetry(ctx), v1.SearchRequest{
		Caller:    callerFromWire(req.GetCaller()),
		Settings:  req.GetSettings(),
		Text:      req.GetText(),
		MediaType: v1.MediaType(req.GetMediaType()),
		Limit:     int(req.GetLimit()),
	})
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	resp := &modulev1.SearchResponse{}
	for _, r := range out.Results {
		resp.Results = append(resp.Results, searchResultToWire(r))
	}
	return resp, nil
}

func (s *capabilityServer) Catalogs(ctx context.Context, req *modulev1.CatalogsRequest) (*modulev1.CatalogsResponse, error) {
	p, ok := s.impl.(v1.CatalogProvider)
	if !ok {
		return nil, unimplemented(string(v1.RoleCatalog))
	}
	out, err := p.Catalogs(s.withTelemetry(ctx), v1.CatalogsRequest{
		Caller:   callerFromWire(req.GetCaller()),
		Settings: req.GetSettings(),
	})
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	resp := &modulev1.CatalogsResponse{}
	for _, c := range out.Catalogs {
		resp.Catalogs = append(resp.Catalogs, catalogToWire(c))
	}
	return resp, nil
}

func (s *capabilityServer) CatalogItems(ctx context.Context, req *modulev1.CatalogItemsRequest) (*modulev1.CatalogItemsResponse, error) {
	p, ok := s.impl.(v1.CatalogProvider)
	if !ok {
		return nil, unimplemented(string(v1.RoleCatalog))
	}
	out, err := p.CatalogItems(s.withTelemetry(ctx), v1.CatalogItemsRequest{
		Caller:     callerFromWire(req.GetCaller()),
		Settings:   req.GetSettings(),
		CatalogID:  req.GetCatalogId(),
		NativeType: req.GetNativeType(),
		Skip:       int(req.GetSkip()),
	})
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	resp := &modulev1.CatalogItemsResponse{}
	for _, i := range out.Items {
		resp.Items = append(resp.Items, catalogItemToWire(i))
	}
	return resp, nil
}

func (s *capabilityServer) Streams(ctx context.Context, req *modulev1.StreamsRequest) (*modulev1.StreamsResponse, error) {
	p, ok := s.impl.(v1.StreamProvider)
	if !ok {
		return nil, unimplemented(string(v1.RoleStream))
	}
	out, err := p.Streams(s.withTelemetry(ctx), v1.StreamRequest{
		Caller:   callerFromWire(req.GetCaller()),
		Settings: req.GetSettings(),
		Ref:      refFromWire(req.GetRef()),
		Season:   int(req.GetSeason()),
		Episode:  int(req.GetEpisode()),
	})
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	resp := &modulev1.StreamsResponse{}
	for _, l := range out.Streams {
		resp.Streams = append(resp.Streams, streamLinkToWire(l))
	}
	return resp, nil
}

func (s *capabilityServer) Subtitles(ctx context.Context, req *modulev1.SubtitlesRequest) (*modulev1.SubtitlesResponse, error) {
	p, ok := s.impl.(v1.SubtitlesProvider)
	if !ok {
		return nil, unimplemented(string(v1.RoleSubtitles))
	}
	out, err := p.Subtitles(s.withTelemetry(ctx), v1.SubtitlesRequest{
		Caller:   callerFromWire(req.GetCaller()),
		Settings: req.GetSettings(),
		Ref:      refFromWire(req.GetRef()),
	})
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	resp := &modulev1.SubtitlesResponse{}
	for _, sub := range out.Subtitles {
		resp.Subtitles = append(resp.Subtitles, subtitleToWire(sub))
	}
	return resp, nil
}

func (s *capabilityServer) Artwork(ctx context.Context, req *modulev1.ArtworkRequest) (*modulev1.ArtworkResponse, error) {
	p, ok := s.impl.(v1.ArtworkProvider)
	if !ok {
		return nil, unimplemented(string(v1.RoleArtwork))
	}
	in := v1.ArtworkRequest{
		Caller:    callerFromWire(req.GetCaller()),
		Settings:  req.GetSettings(),
		MediaType: v1.MediaType(req.GetMediaType()),
		Season:    int(req.GetSeason()),
	}
	for _, id := range req.GetIdentities() {
		in.Identities = append(in.Identities, identityFromWire(id))
	}
	out, err := p.Artwork(s.withTelemetry(ctx), in)
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	resp := &modulev1.ArtworkResponse{}
	for _, c := range out.Candidates {
		resp.Candidates = append(resp.Candidates, artworkCandidateToWire(c))
	}
	return resp, nil
}

func (s *capabilityServer) Playback(ctx context.Context, req *modulev1.PlaybackRequest) (*modulev1.PlaybackResponse, error) {
	p, ok := s.impl.(v1.PlaybackProvider)
	if !ok {
		return nil, unimplemented(string(v1.RolePlayback))
	}
	out, err := p.Resolve(s.withTelemetry(ctx), v1.PlaybackRequest{
		Caller:   callerFromWire(req.GetCaller()),
		Settings: req.GetSettings(),
		Part:     partFromWire(req.GetPart()),
	})
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	return &modulev1.PlaybackResponse{
		Kind:    playbackKindToWire(out.Kind),
		Url:     out.URL,
		Headers: out.Headers,
	}, nil
}

func (s *capabilityServer) SettingsUI(ctx context.Context, req *modulev1.SettingsUIRequest) (*modulev1.SettingsUIResponse, error) {
	p, ok := s.impl.(v1.SettingsUIProvider)
	if !ok {
		return nil, unimplemented(string(v1.RoleSettingsUI))
	}
	out, err := p.SettingsUI(s.withTelemetry(ctx), v1.SettingsUIRequest{
		Caller:   callerFromWire(req.GetCaller()),
		Settings: req.GetSettings(),
	})
	if err != nil {
		return nil, errorToWire(err, nil)
	}
	return &modulev1.SettingsUIResponse{Ui: out.UI}, nil
}
