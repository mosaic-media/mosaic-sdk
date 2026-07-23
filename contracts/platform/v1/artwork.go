package v1

// Artwork is the set of image URLs shown for a content node — the poster on a
// card, the backdrop behind a hero, the logo used as a title treatment.
//
// It is stored on the node at materialisation rather than re-derived from the
// provider on every read (ADR 0071). That is what lets a list surface such as
// the continue-watching rail render from a single node read instead of a
// metadata round-trip per card, and what makes a node's art something a user can
// later override — a choice possible only for artwork the library owns, not for
// a value re-derived on every view.
//
// The URLs are the provider's primary choice per type. A future candidate set
// and user selection grows this value without changing what it means. An empty
// field is "the source had none", which a renderer reads as "fall back", not as
// a blank image.
//
// The struct tags give the value a stable lower-case shape wherever it is
// serialised — the Platform stores it as a JSON document — so the storage form
// does not track Go's exported-field capitalisation.
type Artwork struct {
	// Poster is portrait key art — the image a card shows. For an episode node
	// it is the episode still.
	Poster string `json:"poster,omitempty"`
	// Backdrop is landscape art shown behind a hero.
	Backdrop string `json:"backdrop,omitempty"`
	// Logo is the clearlogo / title-treatment image, rendered as a hero's title.
	Logo string `json:"logo,omitempty"`
}

// Empty reports whether the value carries no artwork at all, so a caller can
// tell "this node has stored art" from "this node has none" without inspecting
// each field.
func (a Artwork) Empty() bool {
	return a.Poster == "" && a.Backdrop == "" && a.Logo == ""
}
