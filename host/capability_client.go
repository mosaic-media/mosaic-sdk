package host

import (
	"context"

	modulev1 "github.com/mosaic-media/contracts/gen/mosaic/module/v1"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// capabilityClient runs in the Platform process and implements [v1.Capability]
// — and every provider role — by calling the module process.
//
// The `CapabilityRegistry` holds it as a [v1.Capability] and cannot tell it from
// a compiled-in module. That is the property ADR 0064 is arranged around, and it
// is why no Platform code above the registry changes.
//
// **It implements every role interface unconditionally**, which matters because
// Go's type assertions are how the Platform discovers roles in process
// (`cap.(v1.SearchProvider)`). Against a proxy that assertion always succeeds,
// so it is not a usable test of what the module actually fills. Use
// [Roles] instead, which reads the manifest — trustworthy because the handshake
// refuses a module whose manifest declares a role it does not serve.
type capabilityClient struct {
	client modulev1.CapabilityServiceClient
}

var (
	_ v1.Capability         = (*capabilityClient)(nil)
	_ v1.MetadataProvider   = (*capabilityClient)(nil)
	_ v1.SearchProvider     = (*capabilityClient)(nil)
	_ v1.CatalogProvider    = (*capabilityClient)(nil)
	_ v1.StreamProvider     = (*capabilityClient)(nil)
	_ v1.SubtitlesProvider  = (*capabilityClient)(nil)
	_ v1.ArtworkProvider    = (*capabilityClient)(nil)
	_ v1.PlaybackProvider   = (*capabilityClient)(nil)
	_ v1.SettingsUIProvider = (*capabilityClient)(nil)
)

// Roles reports the roles a module's manifest declares. It is the correct way
// for the Platform to discover what an out-of-process module fills, because a
// type assertion against the proxy always succeeds.
func Roles(c v1.Capability) []v1.Role {
	return c.Manifest().Provides
}

// Manifest is cached by the Platform after the handshake in practice, but this
// makes the call each time rather than memoising: a manifest is read rarely, and
// a cache here would be a second place for the Platform's own registration state
// to disagree with the module.
func (c *capabilityClient) Manifest() v1.Manifest {
	resp, err := c.client.GetManifest(context.Background(), &modulev1.ManifestRequest{})
	if err != nil {
		// The interface cannot report an error. A zero manifest is the honest
		// answer for a module that is not answering, and the Platform's health
		// checking is what notices — not a silently plausible manifest.
		return v1.Manifest{}
	}
	return manifestFromWire(resp.GetManifest())
}

// Import ignores the svc argument, and that is not a bug. In process the
// Platform hands a capability the ContentService to write through; across the
// boundary the module reaches its own client of that same service over the
// callback stream, because a Go interface value cannot be serialized. The
// service the module ends up calling is the one the Platform passed to
// [ClientPluginMap], so the authority and the destination are identical — only
// the delivery differs.
func (c *capabilityClient) Import(ctx context.Context, _ v1.ContentService, req v1.ImportRequest) (v1.ImportResult, error) {
	resp, err := c.client.Import(ctx, &modulev1.ImportRequest{
		Caller:   callerToWire(req.Caller),
		Ref:      refToWire(req.Ref),
		Settings: req.Settings,
	})
	if err != nil {
		return v1.ImportResult{}, errorFromWire(err)
	}
	return v1.ImportResult{
		WorkID:       v1.NodeID(resp.GetWorkId()),
		AlreadyKnown: resp.GetAlreadyKnown(),
		Containers:   int(resp.GetContainers()),
		Items:        int(resp.GetItems()),
		Parts:        int(resp.GetParts()),
	}, nil
}

func (c *capabilityClient) Metadata(ctx context.Context, req v1.MetadataRequest) (v1.ContentMetadata, error) {
	resp, err := c.client.Metadata(ctx, &modulev1.MetadataRequest{
		Caller:   callerToWire(req.Caller),
		Settings: req.Settings,
		Ref:      refToWire(req.Ref),
	})
	if err != nil {
		return v1.ContentMetadata{}, errorFromWire(err)
	}
	return metadataFromWire(resp.GetMetadata()), nil
}

func (c *capabilityClient) Search(ctx context.Context, req v1.SearchRequest) (v1.SearchResponse, error) {
	resp, err := c.client.Search(ctx, &modulev1.SearchRequest{
		Caller:    callerToWire(req.Caller),
		Settings:  req.Settings,
		Text:      req.Text,
		MediaType: string(req.MediaType),
		Limit:     int32(req.Limit),
	})
	if err != nil {
		return v1.SearchResponse{}, errorFromWire(err)
	}
	out := v1.SearchResponse{}
	for _, r := range resp.GetResults() {
		out.Results = append(out.Results, searchResultFromWire(r))
	}
	return out, nil
}

func (c *capabilityClient) Catalogs(ctx context.Context, req v1.CatalogsRequest) (v1.CatalogsResponse, error) {
	resp, err := c.client.Catalogs(ctx, &modulev1.CatalogsRequest{
		Caller:   callerToWire(req.Caller),
		Settings: req.Settings,
	})
	if err != nil {
		return v1.CatalogsResponse{}, errorFromWire(err)
	}
	out := v1.CatalogsResponse{}
	for _, cat := range resp.GetCatalogs() {
		out.Catalogs = append(out.Catalogs, catalogFromWire(cat))
	}
	return out, nil
}

func (c *capabilityClient) CatalogItems(ctx context.Context, req v1.CatalogItemsRequest) (v1.CatalogItemsResponse, error) {
	resp, err := c.client.CatalogItems(ctx, &modulev1.CatalogItemsRequest{
		Caller:     callerToWire(req.Caller),
		Settings:   req.Settings,
		CatalogId:  req.CatalogID,
		NativeType: req.NativeType,
		Skip:       int32(req.Skip),
	})
	if err != nil {
		return v1.CatalogItemsResponse{}, errorFromWire(err)
	}
	out := v1.CatalogItemsResponse{}
	for _, i := range resp.GetItems() {
		out.Items = append(out.Items, catalogItemFromWire(i))
	}
	return out, nil
}

func (c *capabilityClient) Streams(ctx context.Context, req v1.StreamRequest) (v1.StreamResponse, error) {
	resp, err := c.client.Streams(ctx, &modulev1.StreamsRequest{
		Caller:   callerToWire(req.Caller),
		Settings: req.Settings,
		Ref:      refToWire(req.Ref),
		Season:   int32(req.Season),
		Episode:  int32(req.Episode),
	})
	if err != nil {
		return v1.StreamResponse{}, errorFromWire(err)
	}
	out := v1.StreamResponse{}
	for _, l := range resp.GetStreams() {
		out.Streams = append(out.Streams, streamLinkFromWire(l))
	}
	return out, nil
}

func (c *capabilityClient) Subtitles(ctx context.Context, req v1.SubtitlesRequest) (v1.SubtitlesResponse, error) {
	resp, err := c.client.Subtitles(ctx, &modulev1.SubtitlesRequest{
		Caller:   callerToWire(req.Caller),
		Settings: req.Settings,
		Ref:      refToWire(req.Ref),
	})
	if err != nil {
		return v1.SubtitlesResponse{}, errorFromWire(err)
	}
	out := v1.SubtitlesResponse{}
	for _, s := range resp.GetSubtitles() {
		out.Subtitles = append(out.Subtitles, subtitleFromWire(s))
	}
	return out, nil
}

func (c *capabilityClient) Artwork(ctx context.Context, req v1.ArtworkRequest) (v1.ArtworkResponse, error) {
	wire := &modulev1.ArtworkRequest{
		Caller:    callerToWire(req.Caller),
		Settings:  req.Settings,
		MediaType: string(req.MediaType),
		Season:    int32(req.Season),
	}
	for _, id := range req.Identities {
		wire.Identities = append(wire.Identities, identityToWire(id))
	}
	resp, err := c.client.Artwork(ctx, wire)
	if err != nil {
		return v1.ArtworkResponse{}, errorFromWire(err)
	}
	out := v1.ArtworkResponse{}
	for _, cand := range resp.GetCandidates() {
		out.Candidates = append(out.Candidates, artworkCandidateFromWire(cand))
	}
	return out, nil
}

func (c *capabilityClient) Resolve(ctx context.Context, req v1.PlaybackRequest) (v1.PlaybackResolution, error) {
	resp, err := c.client.Playback(ctx, &modulev1.PlaybackRequest{
		Caller:   callerToWire(req.Caller),
		Settings: req.Settings,
		Part:     partToWire(req.Part),
	})
	if err != nil {
		return v1.PlaybackResolution{}, errorFromWire(err)
	}
	return v1.PlaybackResolution{
		Kind:    playbackKindFromWire(resp.GetKind()),
		URL:     resp.GetUrl(),
		Headers: resp.GetHeaders(),
	}, nil
}

func (c *capabilityClient) SettingsUI(ctx context.Context, req v1.SettingsUIRequest) (v1.SettingsUIResponse, error) {
	resp, err := c.client.SettingsUI(ctx, &modulev1.SettingsUIRequest{
		Caller:   callerToWire(req.Caller),
		Settings: req.Settings,
	})
	if err != nil {
		return v1.SettingsUIResponse{}, errorFromWire(err)
	}
	return v1.SettingsUIResponse{UI: resp.GetUi()}, nil
}
