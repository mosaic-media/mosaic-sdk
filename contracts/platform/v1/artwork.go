package v1

// Artwork is the set of images shown for a content node — the poster on a card,
// the backdrop behind a hero, the logo used as a title treatment.
//
// It is stored on the node at materialisation rather than re-derived from the
// provider on every read (ADR 0071). That is what lets a list surface such as
// the continue-watching rail render from a single node read instead of a
// metadata round-trip per card, and what makes a node's art something a user can
// override — a choice possible only for artwork the library owns, not for a
// value re-derived on every view.
//
// # The flat slots are the selection; Candidates is what it was selected from
//
// The four flat fields hold one URL each and are what a renderer reads. They
// used to mean "the artwork the provider gave"; since ADR 0074 they mean "the
// artwork that was *chosen*", resolved once when the node is written.
// Candidates carries what it was chosen from — every image every source
// offered, with enough provenance to choose again differently.
//
// Keeping both is deliberate and is the whole shape of ADR 0074. A consumer
// never walks the candidate list to find out what to draw: there is exactly one
// answer to "what is this node's poster" and it is the field called Poster. The
// alternative — a list plus a selection index — would make every surface resolve
// the selection itself, and each would do it slightly differently.
//
// An empty field is "nothing was selected for this slot", which a renderer reads
// as "fall back", not as a blank image.
//
// The struct tags give the value a stable lower-case shape wherever it is
// serialised — the Platform stores it as a JSON document — so the storage form
// does not track Go's exported-field capitalisation. A document written before
// candidates existed decodes as a selection with nothing behind it, which is
// exactly what it was.
type Artwork struct {
	// Poster is portrait key art — the image a card shows. For an episode node
	// it is the episode still.
	Poster string `json:"poster,omitempty"`
	// Landscape is wide key art: the same title treated as a 16:9 card rather
	// than a portrait poster. It is distinct from Backdrop — a backdrop is
	// scenery to sit *behind* a hero, this is a composed card image to sit *in*
	// one, which is what a resume rail wants. Sources differ on whether they
	// have it: Cinemeta does not, an addon proxying a real artwork database
	// does, and it is empty rather than substituted when absent.
	Landscape string `json:"landscape,omitempty"`
	// Backdrop is landscape art shown behind a hero.
	Backdrop string `json:"backdrop,omitempty"`
	// Logo is the clearlogo / title-treatment image, rendered as a hero's title.
	Logo string `json:"logo,omitempty"`
	// Candidates is every image any source offered for this node, best-first
	// within each slot (ADR 0074).
	//
	// **Ordering is a write-time obligation, not a read-time one.** Whoever
	// assembles the set sorts it, so reading the best candidate for a slot is
	// taking the first one rather than re-running a policy on every render. That
	// is what keeps selection in one place and off the hot path.
	//
	// It holds slots the flat fields have no room for — a banner, clearart, disc
	// art — which is how the slot vocabulary grows without this struct growing a
	// field per art type and becoming a bag.
	Candidates []ArtworkCandidate `json:"candidates,omitempty"`
}

// ArtworkSlot names a kind of image — what the picture is *for*, not what it
// depicts. It is the axis a consumer selects on: a hero wants a backdrop, a card
// wants a poster, and the two are not interchangeable even for the same title.
//
// Open text with known values, like the media vocabularies (ADR 0015). A
// dedicated artwork source has types Mosaic has never heard of, and carrying an
// unrecognised one costs nothing while dropping it loses data a later Mosaic
// could use. A consumer that does not recognise a slot ignores it.
//
// Only the first four have a flat field on Artwork. The rest are reachable
// through Candidates, deliberately — see Artwork.Candidates.
type ArtworkSlot string

const (
	// ArtworkPoster is portrait key art, the image a card shows.
	ArtworkPoster ArtworkSlot = "poster"
	// ArtworkLandscape is wide composed key art, to sit *in* a card.
	ArtworkLandscape ArtworkSlot = "landscape"
	// ArtworkBackdrop is scenery to sit *behind* a hero.
	ArtworkBackdrop ArtworkSlot = "backdrop"
	// ArtworkLogo is the clearlogo / title treatment.
	ArtworkLogo ArtworkSlot = "logo"
	// ArtworkClearArt is a transparent cut-out treatment of the title's key art,
	// to lay over a backdrop. Cinemeta could never supply it and ADR 0034
	// recorded its absence as a gap waiting on a dedicated artwork source.
	ArtworkClearArt ArtworkSlot = "clearart"
	// ArtworkBanner is wide, short title art — a shape neither a poster nor a
	// backdrop fills. Recorded alongside clearart in ADR 0034's gap.
	ArtworkBanner ArtworkSlot = "banner"
	// ArtworkDisc is disc or label art for a physical edition.
	ArtworkDisc ArtworkSlot = "disc"
	// ArtworkCharacterArt is a cut-out render of a character from the title.
	ArtworkCharacterArt ArtworkSlot = "characterart"
)

// ArtworkCandidate is one image one source offers for one slot.
//
// It carries provenance rather than only a URL, because the point of a candidate
// set is choosing between entries, and every field here is something a choice
// gets made on.
type ArtworkCandidate struct {
	// Slot is what this image is for.
	Slot ArtworkSlot `json:"slot"`
	// URL is where the image lives, at the source's own CDN.
	//
	// It is not a Platform artwork-proxy URL. Proxying and signing happen when a
	// screen is emitted (ADR 0030), and a signed URL stored here would outlive
	// the process-scoped key that signed it — coming back after a restart as a
	// page that looks right and is broken.
	URL string `json:"url"`
	// Source is the module id that supplied this candidate, so a set assembled
	// from several providers stays attributable and a later per-provider
	// preference has something to key on.
	Source string `json:"source,omitempty"`
	// Language is the ISO 639 code of any text burned into the image, empty when
	// there is none.
	//
	// **Empty means textless, and textless is frequently the best answer** — a
	// backdrop with a title burned into it is wrong behind a hero that draws its
	// own clearlogo on top. So an empty Language here is a positive property
	// rather than missing data, which is the opposite of how most empty fields in
	// this package read.
	Language string `json:"language,omitempty"`
	// Rank is the source's own ordering of its candidates — vote counts,
	// popularity, or simply the order it listed them.
	//
	// **It is not normalised and must not be compared across sources.** One
	// source's likes and another's vote average measure different things over
	// different populations; a blended score would read as authoritative while
	// being invented, and the store would then persist the invention. Rank orders
	// candidates *within* a source; choosing between sources is a separate,
	// stated preference.
	Rank float64 `json:"rank,omitempty"`
}

// Empty reports whether the value carries no artwork at all, so a caller can
// tell "this node has stored art" from "this node has none" without inspecting
// each field.
//
// A node with candidates but no selection is not empty: it has art, and
// something failed to resolve it.
func (a Artwork) Empty() bool {
	return a.Poster == "" && a.Landscape == "" && a.Backdrop == "" &&
		a.Logo == "" && len(a.Candidates) == 0
}

// Slot is the resolved URL for one slot — the uniform read accessor.
//
// For the four slots with a flat field it returns that field, which is the
// selection. For every other slot, and for a selection that was never resolved,
// it falls back to the best candidate. That fallback is what lets a consumer ask
// for a banner or clearart in the same breath as a poster, without having to
// know which slots happen to have a field.
//
// Candidates are stored best-first, so "best" is "first" and this stays a scan
// rather than a sort.
func (a Artwork) Slot(slot ArtworkSlot) string {
	switch slot {
	case ArtworkPoster:
		if a.Poster != "" {
			return a.Poster
		}
	case ArtworkLandscape:
		if a.Landscape != "" {
			return a.Landscape
		}
	case ArtworkBackdrop:
		if a.Backdrop != "" {
			return a.Backdrop
		}
	case ArtworkLogo:
		if a.Logo != "" {
			return a.Logo
		}
	}
	for _, candidate := range a.Candidates {
		if candidate.Slot == slot && candidate.URL != "" {
			return candidate.URL
		}
	}
	return ""
}

// CandidatesFor returns every candidate for one slot, best-first, so a caller
// building a picker renders the alternatives without filtering the set itself.
//
// It returns nil rather than an empty slice when there are none, matching the
// zero value a caller gets from a node with no stored art.
func (a Artwork) CandidatesFor(slot ArtworkSlot) []ArtworkCandidate {
	var out []ArtworkCandidate
	for _, candidate := range a.Candidates {
		if candidate.Slot == slot {
			out = append(out, candidate)
		}
	}
	return out
}
