package v1

import "context"

// The provider surface (ADR 0027). A module is not only an importer: a source
// exposes resources — metadata, search, catalogs, streams — and a module
// contributes them as typed provider roles the Platform inherits on enable and,
// by design, other modules resolve. Each role is a small interface a module
// implements only if it fills it, and declares in its Manifest.Provides.
//
// The read roles return the *virtual* content types below (SearchResult,
// CatalogItem, ContentMetadata) — transient projections, never object-graph
// nodes (ADR 0028). Only Import (see capability.go) writes nodes, and it does
// so from a ContentRef a read role produced. Nothing here takes a
// ContentService: reads do not write.

// Role names a provider role a module fills. It is the vocabulary of
// Manifest.Provides. A module declares the roles it implements; the Platform
// checks at composition that a declared role is backed by the matching
// interface below (a role named but not implemented is a composition-time
// error, not a runtime surprise), and resolves providers by role at runtime.
type Role string

const (
	// RoleMetadata is backed by MetadataProvider — the addon `meta` resource.
	RoleMetadata Role = "metadata"
	// RoleSearch is backed by SearchProvider — the addon `catalog/…/search`
	// resource.
	RoleSearch Role = "search"
	// RoleCatalog is backed by CatalogProvider — the addon `catalog` resource.
	RoleCatalog Role = "catalog"
	// RoleStream is backed by StreamProvider — the addon `stream` resource.
	RoleStream Role = "stream"
	// RoleSubtitles is backed by SubtitlesProvider — the addon `subtitles`
	// resource. It is a source role like the others: it resolves subtitle tracks
	// for an item. The consumer is a player, which does not exist yet — the role
	// is defined and filled ahead of it so the source is complete (ADR 0037), the
	// same way a stream location is snapshotted before anything resolves it.
	RoleSubtitles Role = "subtitles"
	// RolePlayback is backed by PlaybackProvider (playback.go) — the first
	// *consumer* role, and the one entry here that is not a source. Every role
	// above brings content in; this one acts on what materialising created,
	// resolving a Part to playable bytes (ADR 0045). It is what ADR 0036's
	// affordance gate keys on: with no consumer installed, the library is inert
	// and the surface is discovery-only.
	RolePlayback Role = "playback"
	// RoleArtwork is backed by ArtworkProvider — a source that supplies *only*
	// artwork for content something else already identified (ADR 0075).
	//
	// It is neither a source role nor a consumer role, which is why it is worth
	// reading twice. Every role above either brings content in or acts on what
	// materialising created; this one **enriches content that already exists**. A
	// dedicated artwork database has no titles, no overviews, no search and no
	// catalogs — there is no query that turns "Blade Runner" into a result, you
	// must already know which film you mean — so it can answer exactly one
	// question and none of the others.
	//
	// **It does not satisfy ADR 0035's metadata requirement, and must not be
	// used to.** A module that cannot name a film has no business counting
	// toward the guarantee that Mosaic can name films. That is the specific
	// mistake this role exists to prevent: declaring RoleMetadata to reach
	// ContentMetadata's image fields would pass the composition-root check with a
	// module that cannot describe anything, and the failure would be a
	// deployment that boots and cannot identify content — not a red test.
	RoleArtwork Role = "artwork"
	// RoleSettingsUI is backed by SettingsUIProvider — a module contributing its
	// own settings screen as SDUI (ADR 0038). Unlike the source roles it produces
	// no content: it renders the module's configuration UI, which the Platform
	// hosts in a bounded settings slot. It is the one place a module contributes a
	// screen rather than data — scoped to its own settings, never content.
	RoleSettingsUI Role = "settings_ui"
)

// ContentRef identifies a piece of source content that is not (necessarily) in
// the library yet — the stable handle a virtual result carries so the Platform
// can dedup it against the library and hand it back to the module to
// materialise (ADR 0028). It is produced by a read role and consumed by Import.
type ContentRef struct {
	// Provider is the id of the capability that produced this ref and can
	// materialise it — the Manifest.ID the Platform routes Import back to. A
	// search fans out to many providers, so each result must name its own.
	Provider string
	// NativeID is the source's own identifier for the content (an addon's
	// content id — an IMDB id in Stremio's case). The module owns its meaning;
	// the Platform never parses it.
	NativeID string
	// NativeType is the source's own content type ("movie", "series"), which the
	// module needs to route and to shape the tree on materialise. Module-owned,
	// like NativeID.
	NativeType string
	// MediaType is the Platform media type this maps to (ADR 0015), for display
	// and library filtering. The module derives it from NativeType.
	MediaType MediaType
	// ExternalScheme and ExternalID are the provider identity the Platform
	// dedups on: a virtual result whose ExternalID already resolves through
	// FindContentByExternalID is marked in-library rather than shown as new
	// (ADR 0028's union). For Stremio these are "imdb" and the IMDB id — the
	// same pair the module binds the Work under on materialise.
	ExternalScheme string
	ExternalID     string
}

// SearchResult is one candidate a SearchProvider returns — a virtual content
// item, not a node. It carries enough to render a row and to materialise later
// (its Ref). InLibrary and NodeID are filled by the Platform when it unions
// provider results with the library (ADR 0028); a provider leaves them zero.
type SearchResult struct {
	Ref    ContentRef
	Title  string
	Year   int
	Poster string
	// InLibrary is true when the Platform matched this result to a library node.
	InLibrary bool
	// NodeID is that node's id when InLibrary, so a client can open it; empty
	// for a purely virtual result.
	NodeID NodeID
}

// Catalog is one collection a CatalogProvider exposes — a *view* the source
// computes (Popular, Trending), addressed by its native id. It is not persisted
// and is not a container in the object graph; materialising a catalog copies its
// items into the library, it does not import the catalog as an object (ADR 0028).
type Catalog struct {
	// ID is the source-native catalog id, passed back to CatalogItems.
	ID string
	// NativeType is the source content type the catalog lists ("movie",
	// "series").
	NativeType string
	// Name is the human label ("Popular Movies").
	Name string
}

// CatalogItem is one entry of a catalog listing — a virtual content item, the
// same shape and rules as SearchResult (it is what the admin collection browser
// renders, ADR 0028). InLibrary and NodeID are Platform-filled.
type CatalogItem struct {
	Ref       ContentRef
	Title     string
	Year      int
	Poster    string
	InLibrary bool
	NodeID    NodeID
}

// Person is one cast or crew credit — a name, and an optional role (a character
// or a job). Sources that expose only names leave Role empty (ADR 0034).
type Person struct {
	Name string
	Role string
	// Photo is a headshot URL, empty when the source has none. Sources differ
	// sharply here: an addon carrying only names is common, and one proxying a
	// real database supplies both a character name and an image. Carrying it is
	// what lets a consumer render a cast rail rather than a list of names —
	// the field was absent, so a source that had photos had nowhere to put them.
	Photo string
}

// EpisodePreview is one episode in a series' descriptive preview (ADR 0034) —
// what a detail screen shows so a user can read the episode list *before*
// deciding to add the series, including for a virtual series that has no
// materialised tree. It is a read-only projection, never persisted and never the
// materialised Node: Import still builds the tree from the source's own
// structure. The Platform hands this list to the UI, which groups it by season.
type EpisodePreview struct {
	Season   int
	Episode  int
	Title    string
	Overview string
	// Thumbnail is a still image URL for the episode, empty when the source has
	// none.
	Thumbnail string
	// Released is the source's air/release date as it provides it (an ISO date or
	// a year), display-only — the Platform does not parse it.
	Released string
}

// RelatedItem is one title related to the one being described — a franchise
// sibling or a recommendation. It is a *virtual* item like a search result: it
// carries a Ref so a consumer can open or materialise it, and the Platform fills
// InLibrary and NodeID when it unions the list against the library (ADR 0028).
//
// It repeats SearchResult's shape rather than reusing it, deliberately. The two
// are produced by different roles and answer different questions, and collapsing
// them would mean a later field that belongs to one appearing on the other — the
// same reason CatalogItem is its own type despite being identical to
// SearchResult today.
type RelatedItem struct {
	Ref    ContentRef
	Title  string
	Year   int
	Poster string
	// InLibrary is true when the Platform matched this item to a library node.
	InLibrary bool
	// NodeID is that node's id when InLibrary, so a client can open it.
	NodeID NodeID
}

// Collection is the franchise a work belongs to — "The Matrix Collection", "the
// other Avatar films" ([ADR 0034](0034) recorded its absence as a gap Cinemeta
// could not fill and a TMDB-class source could).
//
// It is a *descriptive projection*, not the object graph's collection. The graph
// expresses membership as a RelationCollectionMember edge between Works, written
// by Import; this is what a detail screen renders for a title whether or not
// anything has been materialised, which is the case the edge cannot serve — an
// edge needs two Works to exist, and a virtual item is neither.
type Collection struct {
	// Name is the franchise's own label.
	Name string
	// Overview is the franchise's description, empty when the source has none.
	Overview string
	// Poster and Backdrop are the franchise's own artwork, distinct from any
	// member's.
	Poster   string
	Backdrop string
	// Items are the franchise's members, in the source's order — usually
	// chronological. It includes the work being described, so a consumer that
	// wants "the others" filters on the Ref it already holds rather than trusting
	// the source to have excluded it.
	Items []RelatedItem
}

// WatchOfferType is how a provider makes a title available — the distinction
// between "included in what you already pay for" and "costs money now", which is
// the first thing a viewer wants to know.
//
// Open text with known values, like the media vocabularies (ADR 0015): a source
// may report a kind Mosaic has never heard of, and a consumer that does not
// recognise one shows it as-is rather than dropping the offer.
type WatchOfferType string

const (
	// WatchSubscription is included with a subscription the viewer may hold.
	WatchSubscription WatchOfferType = "subscription"
	// WatchRent is a time-limited paid rental.
	WatchRent WatchOfferType = "rent"
	// WatchBuy is a paid purchase.
	WatchBuy WatchOfferType = "buy"
	// WatchFree is free and without advertising.
	WatchFree WatchOfferType = "free"
	// WatchAds is free but advertising-supported.
	WatchAds WatchOfferType = "ads"
)

// WatchOffer is one service a title can be watched on, and on what terms.
type WatchOffer struct {
	// Provider is the service's name as the source gives it — "Netflix",
	// "BBC iPlayer".
	Provider string
	// Logo is the service's logo URL, empty when the source has none.
	Logo string
	// Type is how the title is available on that service.
	Type WatchOfferType
}

// WatchAvailability is where a title can be watched **outside Mosaic**, in one
// region.
//
// # This is not a source and must never be rendered as one
//
// Every other read role answers "what can Mosaic get you"; this one answers
// "where else does this exist". It is not a StreamProvider result, it does not
// become a Part, and nothing in it is playable through the Platform. A client
// that renders an offer as a play control is making a promise the Platform
// cannot keep — the correct affordance is informational, and Link is where it
// sends the viewer.
//
// # It is regional, and empty is not "unavailable"
//
// Availability differs entirely by country, so a value describes exactly one
// Region and a source asked about the wrong one answers nothing. Empty Offers
// means "the source knows of none *here*", which is not the same as the title
// being unavailable — coverage varies by region and by how recently the source
// last looked.
type WatchAvailability struct {
	// Region is the ISO 3166-1 country these offers apply to. A value with no
	// Region is meaningless and a consumer should ignore it.
	Region string
	// Link is the source's own page for this title's availability. It is the
	// right destination for an informational control, and where a source whose
	// terms require sending traffic back expects it to go.
	Link string
	// Attribution names who compiled the availability data, when the source's
	// terms require it to be shown. A consumer that displays offers must display
	// this alongside them; it is carried in the contract precisely so the
	// Platform does not have to know which upstream imposed it.
	Attribution string
	// Offers are the services the title is on, subscription first as the source
	// ranks them.
	Offers []WatchOffer
}

// Trailer is one promotional video a source knows about.
//
// It carries the hosting site and that site's own key rather than a URL, and
// that is the contract: building a watch or embed URL is a client concern, and a
// Platform that assembled one would be choosing an embed policy on the client's
// behalf. A consumer that does not recognise Site ignores the entry.
type Trailer struct {
	// Name is the video's title ("Official Trailer").
	Name string
	// Site is the host as the source names it — "YouTube", "Vimeo".
	Site string
	// Key is the site's own identifier for the video.
	Key string
	// Official is true when the source marks the video as published by the
	// rights holder rather than by a channel.
	Official bool
}

// ContentMetadata is the descriptive detail a MetadataProvider resolves for a
// ContentRef (the `meta` resource) — used to enrich an existing node, to back a
// detail screen (for a virtual item and, re-derived, an in-library one — ADR
// 0034), and as the detail Import draws on when it materialises. It is the
// *descriptive* surface: the flat fields plus, for a series, a read-only episode
// *preview*. It is still not the materialised tree — a work's children are built
// inside Import where the source's structure is known — but it is no longer
// artificially thin (ADR 0034 refines ADR 0027's "flat enrichment DTO").
type ContentMetadata struct {
	Ref      ContentRef
	Title    string
	Year     int
	Overview string
	Poster   string
	Backdrop string
	Genres   []string
	// Logo is the clearlogo/title-treatment image URL, empty when the source has
	// none. It renders as a detail hero's title (ADR 0034).
	Logo string
	// Cast is the top billed cast, best-first as the source ranks it.
	Cast []Person
	// Rating is the source's rating on its own scale (Stremio/IMDB is out of 10),
	// 0 when unknown.
	Rating float64
	// Runtime is a display runtime string ("120 min", "2h 5m") as the source
	// provides it — display-only, since the format varies (ADR 0034).
	Runtime string
	// Episodes is the series episode preview (ADR 0034), empty for a movie or a
	// meta-only source. A read projection the UI groups by season; not the tree.
	Episodes []EpisodePreview
	// Keywords are the source's own descriptive tags ("dystopia", "time loop") —
	// finer-grained than Genres and a better basis for "more like this". Empty
	// when the source has no such vocabulary, which most addons do not.
	Keywords []string
	// Certification is the age rating as the source labels it for the region the
	// source was asked about — "15", "PG-13", "TV-MA". Display-only text, like
	// Runtime: the scales are national and not comparable, so the Platform stores
	// and shows the label rather than mapping it to an invented common scale.
	// Empty when unknown, which must not be read as "suitable for everyone".
	Certification string
	// Similar are titles a viewer of this one is likely to want, best-first as
	// the source ranks them. Empty for a source with no such notion.
	Similar []RelatedItem
	// Collection is the franchise this work belongs to, nil when it belongs to
	// none or the source does not model franchises.
	Collection *Collection
	// Trailers are promotional videos the source knows about, best-first. Empty
	// is the common case; a source that has them supplies a site and a key rather
	// than a URL.
	Trailers []Trailer
	// Watch is where this title can be watched **outside Mosaic**, for one
	// region. Nil when the source has no such data or was not told which region
	// to answer for.
	//
	// It is the one field here that does not describe the title itself, and the
	// one a consumer can most easily render wrongly: these offers are not
	// playable through the Platform and must not appear as play controls. See
	// WatchAvailability.
	Watch *WatchAvailability
}

// StreamLink is one playable location a StreamProvider resolves for a
// materialised item's ContentRef (the `stream` resource). Location is ready to
// attach as a Part (ADR 0014): Stremio yields RemoteLocation refs (a direct URL
// or a magnet), which the Platform snapshots onto the curated item (ADR 0028).
// The descriptive fields (ADR 0037) let a future source-picker rank and display
// candidates; they are best-effort — a source that does not report a field
// leaves it zero.
type StreamLink struct {
	// Label is a short human name for the stream ("Torrentio").
	Label string
	// Title is the source's full descriptive title (the release name), empty when
	// it only gave a Label.
	Title string
	// Quality is a display quality label the source exposes ("1080p", "2160p"),
	// empty when unknown.
	Quality string
	// SizeBytes is the stream's size when the source reports it, 0 otherwise.
	SizeBytes int64
	// Seeders is swarm health for a torrent stream, 0 when not a torrent or
	// unreported.
	Seeders  int
	Location MediaLocation
}

// Subtitle is one subtitle track a SubtitlesProvider resolves for an item (ADR
// 0037). It is a source projection like a stream location — the Platform does
// not fetch or store the file; a player consumes it.
type Subtitle struct {
	// Language is the track's language as the source labels it (an ISO 639 code
	// or a display name).
	Language string
	// URL is the subtitle file location.
	URL string
	// ID is the source's own id for the track, empty when it has none.
	ID string
}

// SubtitlesProvider resolves subtitle tracks for a materialised item's
// ContentRef (the `subtitles` resource). A module fills RoleSubtitles by
// implementing it. Like StreamProvider it is a source role; the consumer is a
// player (ADR 0036's deferred playback capability), so a provider may exist
// before anything consumes it.
type SubtitlesProvider interface {
	Subtitles(ctx context.Context, req SubtitlesRequest) (SubtitlesResponse, error)
}

// SubtitlesRequest names the item to resolve subtitles for by its ContentRef
// (for a series episode, the ref's NativeID is the episode id the source uses,
// exactly as StreamRequest does).
type SubtitlesRequest struct {
	Caller   Caller
	Settings []byte
	Ref      ContentRef
}

// SubtitlesResponse carries the resolved tracks, best-first as the source ranks
// them.
type SubtitlesResponse struct {
	Subtitles []Subtitle
}

// ExternalIdentity is one shared identifier for a piece of content — a scheme
// and the id under it ("imdb"/"tt0083658", "tvdb"/"73739").
//
// It is the *neutral* addressing a provider is handed when it did not source the
// content it is being asked about (ADR 0073, ADR 0075). A native id is the
// producing module's own business and means nothing to anyone else; a shared
// identity is the one thing two modules that have never heard of each other can
// both recognise.
type ExternalIdentity struct {
	// Scheme names the identifier space — "imdb", "tvdb", "tmdb".
	Scheme string
	// ID is the identifier within that scheme.
	ID string
}

// ArtworkProvider resolves artwork candidates for content that already exists.
// A module fills RoleArtwork by implementing it.
//
// It is called only about content it did not source, which is the whole point:
// the Platform runs it as a best-effort enrichment pass over a materialised work
// (ADR 0075), the same shape ADR 0073 established for stream providers. A
// provider that recognises none of the identities it was handed returns an empty
// response and **no error** — being asked about content it does not know is
// normal, not a failure, and an error there would make an artwork source being
// down lose a user the work they just added.
type ArtworkProvider interface {
	Artwork(ctx context.Context, req ArtworkRequest) (ArtworkResponse, error)
}

// ArtworkRequest names the content to find artwork for by its shared external
// identities, rather than by a ContentRef.
//
// **The whole identity set is handed over at once, not one at a time**, and that
// differs from how StreamRequest is invoked. Which identifier a source prefers
// is the source's business: a real artwork database keys films by one scheme and
// television by another, and letting the module pick from what is available
// keeps that dialect where ADR 0051 says it belongs. Looping identity-by-identity
// Platform-side would also turn one decline into several.
type ArtworkRequest struct {
	Caller   Caller
	Settings []byte
	// Identities are the shared external ids the content is known by, in stable
	// scheme order. A provider takes the first it speaks and ignores the rest.
	Identities []ExternalIdentity
	// MediaType is the Platform media type, which is what tells a provider
	// whether it is being asked about a film or a series — the two are
	// frequently different endpoints keyed by different schemes.
	MediaType MediaType
	// Season is the season number when the request is about one season of a
	// series, and 0 when it is about the work itself. Season artwork is a real
	// distinction rather than a filter: a source has per-season posters, and a
	// series that shows one poster four times is the thing having them fixes.
	Season int
}

// ArtworkResponse carries the candidates the source offers, best-first as it
// ranks them.
//
// A provider fills each candidate's Slot, URL, Language and Rank. It may leave
// Source empty: the Platform stamps its module id, so provenance is recorded by
// the one party that cannot get it wrong.
type ArtworkResponse struct {
	Candidates []ArtworkCandidate
}

// SettingsUIProvider lets a module contribute its own settings screen as SDUI
// (ADR 0038). The SDK stays SDUI-agnostic: the screen is returned as a
// serialised UINode tree (JSON bytes), not a typed SDUI value, so the contract
// does not depend on the SDUI package. The module builds it with the published
// mosaic-sdui producer binding; the Platform validates the bytes and renders
// them in a bounded settings slot. A module fills RoleSettingsUI by implementing
// it.
type SettingsUIProvider interface {
	SettingsUI(ctx context.Context, req SettingsUIRequest) (SettingsUIResponse, error)
}

// SettingsUIRequest carries the caller and the module's current settings, so the
// module can render the live configuration (e.g. the installed addons) into the
// screen.
type SettingsUIRequest struct {
	Caller   Caller
	Settings []byte
}

// SettingsUIResponse carries the settings screen as a serialised UINode tree.
// Empty UI is a valid "no settings screen" response.
type SettingsUIResponse struct {
	UI []byte
}

// MetadataProvider resolves full descriptive metadata for a ContentRef. It
// backs enrichment (update a node's metadata) and supplies the detail Import
// draws on. A module fills RoleMetadata by implementing it.
type MetadataProvider interface {
	Metadata(ctx context.Context, req MetadataRequest) (ContentMetadata, error)
}

// MetadataRequest is what the Platform hands MetadataProvider. Every provider
// request carries the Caller it acts as (ADR 0017) and the module's Settings
// (ADR 0021) — the same two the Platform hands Import — plus the role's params.
type MetadataRequest struct {
	Caller   Caller
	Settings []byte
	Ref      ContentRef
}

// SearchProvider answers free-text search over the source, returning virtual
// candidates with no raw id needed. A module fills RoleSearch by implementing it.
type SearchProvider interface {
	Search(ctx context.Context, req SearchRequest) (SearchResponse, error)
}

// SearchRequest carries the query text and an optional media-type filter. Limit
// is a hint; a provider may return fewer.
type SearchRequest struct {
	Caller    Caller
	Settings  []byte
	Text      string
	MediaType MediaType
	Limit     int
}

// SearchResponse is a slice wrapper so the surface can grow (paging) without
// breaking the interface.
type SearchResponse struct {
	Results []SearchResult
}

// CatalogProvider enumerates the source's collections and their items. A module
// fills RoleCatalog by implementing it.
type CatalogProvider interface {
	// Catalogs lists the collections the source exposes.
	Catalogs(ctx context.Context, req CatalogsRequest) (CatalogsResponse, error)
	// CatalogItems lists one catalog's entries, addressed by its native id.
	CatalogItems(ctx context.Context, req CatalogItemsRequest) (CatalogItemsResponse, error)
}

// CatalogsRequest asks for the source's catalog list; it needs only the caller
// and settings.
type CatalogsRequest struct {
	Caller   Caller
	Settings []byte
}

// CatalogsResponse carries the catalog descriptors.
type CatalogsResponse struct {
	Catalogs []Catalog
}

// CatalogItemsRequest addresses one catalog by its native id and type, with a
// Skip offset for paging through a large collection.
type CatalogItemsRequest struct {
	Caller     Caller
	Settings   []byte
	CatalogID  string
	NativeType string
	Skip       int
}

// CatalogItemsResponse carries one page of a catalog's items.
type CatalogItemsResponse struct {
	Items []CatalogItem
}

// StreamProvider resolves playable locations for a materialised item.
//
// It is called in two situations that look the same and are not. In the first,
// the module itself materialised the item and the Ref is its own — the case
// since ADR 0027. In the second, **another module sourced the metadata and this
// provider is being asked to supply the streams for it** (ADR 0073): the Ref
// then carries a *shared* external identity (`imdb`, `tvdb`) with no native id
// at all, and the provider derives its own addressing from that plus the
// request's Season and Episode.
//
// A provider that cannot recognise the identity it was handed returns an empty
// response and no error. Being asked about content it does not know is normal
// in the second case, not a failure, and an error there would fail a user's
// import over a title some other source happened to describe.
type StreamProvider interface {
	Streams(ctx context.Context, req StreamRequest) (StreamResponse, error)
}

// StreamRequest names the item to resolve streams for.
type StreamRequest struct {
	Caller   Caller
	Settings []byte
	Ref      ContentRef
	// Season and Episode locate an episode within a series. Both are zero for a
	// film, and for a series they are 1-based as the source numbers them —
	// season 0 being the specials a source may or may not have.
	//
	// They exist because a stream provider is now asked about content it did not
	// source ([ADR 0073](0073)), and in that case the Ref carries a *shared*
	// external identity rather than the provider's own id. A provider whose
	// episode addressing is derived — "the series' id, a colon, the season, a
	// colon, the episode" is one real example — composes it from these, because
	// that format is the provider's business and the Platform must never learn
	// it.
	//
	// A provider that was handed its own native id ignores these and uses the
	// Ref, exactly as before.
	Season  int
	Episode int
}

// StreamResponse carries the resolved locations, best-first as the source ranks
// them.
type StreamResponse struct {
	Streams []StreamLink
}
