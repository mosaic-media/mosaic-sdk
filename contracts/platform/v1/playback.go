package v1

import "context"

// The consumer surface (ADR 0045). Every role in provider.go is a *source*
// role: it brings content in and populates the virtual plane. This is the first
// role on the other side — one that acts on the materialised library rather
// than producing virtual results (ADR 0036's consumer capability, made
// concrete).
//
// The asymmetry that shapes it: a source role exchanges DTOs, but a consumer
// produces *bytes*, and bytes need a transport the SDK does not have and should
// not grow (ADR 0003 — the Platform owns transports; ADR 0019 — a module
// contributing a route would break the SDK-only boundary). So a provider
// resolves and never serves: it says where the bytes are, or undertakes to
// produce them, and the Platform hosts the origin either way.

// PlaybackKind discriminates how a resolution's bytes are reached. It is a
// closed set that grows deliberately: PlaybackServed (the module produces the
// bytes itself through an io.ReadSeekCloser the Platform serves) is the named
// next variant, for the torrent engine, and a transcoded variant is where
// transcoding would land. Neither exists yet, and neither is declared ahead of
// the slice that fills it.
type PlaybackKind string

const (
	// PlaybackDirect names bytes reachable at a URL — an addon's plain `url`
	// stream, or a link a debrid service has already resolved. The Platform
	// fetches that URL and relays it to the client rather than handing it over
	// (ADR 0045): the viewer's IP never reaches the origin, and a link carrying a
	// credential never leaves the server.
	PlaybackDirect PlaybackKind = "direct"
)

// PlaybackProvider resolves a Part to playable bytes. A module fills
// RolePlayback by implementing it.
//
// It is deliberately one method. ADR 0045 specifies a second — Open, serving an
// io.ReadSeekCloser for a PlaybackServed resolution — and it is not here,
// because nothing produces a Served resolution until the torrent engine does.
// Declaring it now would be surface with no consumer, which is the discipline
// ADR 0036 sets and ADR 0037 departed from only by explicit exception. The SDK
// is pre-1.0 precisely so the method can arrive with the slice that needs it.
type PlaybackProvider interface {
	// Resolve turns a Part's location into something playable. It is called at
	// play time, once per playback, never at import: what a source offered when
	// the item was materialised may be gone by now, and a resolution is specific
	// to the moment it is made.
	//
	// Its error carries a Platform error category like every other role's —
	// Unavailable for an unreachable source, NotFound for a location that no
	// longer resolves, InvalidArgument for a Part it cannot interpret.
	Resolve(ctx context.Context, req PlaybackRequest) (PlaybackResolution, error)
}

// PlaybackRequest is what the Platform hands a playback provider. It carries the
// Caller and Settings every module request carries, plus the Part to play.
//
// The Part is passed in rather than read by the module, which is the point: the
// Platform has already resolved the node, checked the caller's authority and
// loaded the part, so a provider stays a pure function of what it is given and
// needs no graph read to do its job.
type PlaybackRequest struct {
	// Caller is the principal the provider acts as (ADR 0017).
	Caller Caller
	// Settings is the module's user-managed configuration document (ADR 0021),
	// an empty object ({}) when the user has set nothing.
	Settings []byte
	// Part is the item's playable part, with its MediaLocation — the direct URL
	// or magnet the source snapshotted (ADR 0028).
	Part Part
}

// PlaybackResolution is where the bytes are. Read Kind first: the fields that
// carry meaning depend on it, and a field belonging to another variant is zero.
type PlaybackResolution struct {
	// Kind discriminates the variant.
	Kind PlaybackKind

	// URL is the location to fetch, for PlaybackDirect. It is what the Platform
	// relays from; it is not handed to the client.
	URL string
	// Headers are request headers the URL's origin requires — an authorization
	// header a debrid service expects, or the proxy headers a Stremio addon
	// declares. Nil when the URL can be fetched bare. They are applied by the
	// Platform when it fetches, and never forwarded to the client.
	Headers map[string]string
}
