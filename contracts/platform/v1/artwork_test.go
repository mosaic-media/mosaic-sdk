package v1_test

import (
	"encoding/json"
	"testing"

	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// TestArtworkDecodesADocumentWrittenBeforeCandidates is the compatibility claim
// ADR 0074 rests on, made executable.
//
// Artwork is stored as a jsonb document and migration 0019 predicted that a
// candidate set could be added "without a second migration". That is only true
// if a document written before candidates existed still decodes to the same
// selection it always was. Nothing else in the build would notice if it stopped
// being true — the column would keep accepting writes and old rows would quietly
// lose their art.
func TestArtworkDecodesADocumentWrittenBeforeCandidates(t *testing.T) {
	// Exactly what module-cinemeta and module-tmdb have been writing.
	stored := []byte(`{"poster":"https://cdn/p.jpg","backdrop":"https://cdn/b.jpg","logo":"https://cdn/l.png"}`)

	var art v1.Artwork
	if err := json.Unmarshal(stored, &art); err != nil {
		t.Fatalf("unmarshal a pre-candidate document: %v", err)
	}

	if art.Poster != "https://cdn/p.jpg" {
		t.Errorf("Poster = %q, want the stored poster", art.Poster)
	}
	if art.Backdrop != "https://cdn/b.jpg" {
		t.Errorf("Backdrop = %q, want the stored backdrop", art.Backdrop)
	}
	if art.Logo != "https://cdn/l.png" {
		t.Errorf("Logo = %q, want the stored logo", art.Logo)
	}
	if len(art.Candidates) != 0 {
		t.Errorf("Candidates = %v, want none — the row predates them", art.Candidates)
	}
	if art.Empty() {
		t.Error("Empty() = true for a row with three images; a pre-candidate row still has art")
	}

	// And the selection is still readable through the accessor a consumer uses,
	// so a node written before this change renders identically after it.
	if got := art.Slot(v1.ArtworkPoster); got != "https://cdn/p.jpg" {
		t.Errorf("Slot(poster) = %q, want the stored poster", got)
	}
}

// TestArtworkRoundTripsCandidates covers the other direction: a value written
// now must survive storage, since the flat slots and the candidate set are
// serialised into one column and read back as one value.
func TestArtworkRoundTripsCandidates(t *testing.T) {
	want := v1.Artwork{
		Poster: "https://cdn/selected.jpg",
		Candidates: []v1.ArtworkCandidate{
			{Slot: v1.ArtworkPoster, URL: "https://cdn/selected.jpg", Source: "fanart-tv", Language: "en", Rank: 42},
			{Slot: v1.ArtworkBanner, URL: "https://cdn/banner.jpg", Source: "fanart-tv"},
		},
	}

	encoded, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got v1.Artwork
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Poster != want.Poster {
		t.Errorf("Poster = %q, want %q", got.Poster, want.Poster)
	}
	if len(got.Candidates) != 2 {
		t.Fatalf("Candidates = %d, want 2", len(got.Candidates))
	}
	if got.Candidates[0].Rank != 42 {
		t.Errorf("Rank = %v, want 42", got.Candidates[0].Rank)
	}
	// A candidate whose language is absent is textless, and that has to survive
	// the round trip as absent rather than becoming a language.
	if got.Candidates[1].Language != "" {
		t.Errorf("Language = %q, want empty — the banner is textless", got.Candidates[1].Language)
	}
}

// TestArtworkSlotFallsBackToCandidates covers the accessor's whole reason to
// exist: four slots have a flat field and the rest do not, and a consumer asks
// for all of them the same way.
func TestArtworkSlotFallsBackToCandidates(t *testing.T) {
	art := v1.Artwork{
		Poster: "https://cdn/selected.jpg",
		Candidates: []v1.ArtworkCandidate{
			{Slot: v1.ArtworkPoster, URL: "https://cdn/other.jpg"},
			{Slot: v1.ArtworkClearArt, URL: "https://cdn/first-clearart.png"},
			{Slot: v1.ArtworkClearArt, URL: "https://cdn/second-clearart.png"},
		},
	}

	// A slot with a selection returns the selection, not the first candidate.
	// Getting this backwards would silently ignore a user's chosen poster.
	if got := art.Slot(v1.ArtworkPoster); got != "https://cdn/selected.jpg" {
		t.Errorf("Slot(poster) = %q, want the selection to win over a candidate", got)
	}
	// A slot with no flat field resolves to its best candidate, which is the
	// first — candidates are stored best-first.
	if got := art.Slot(v1.ArtworkClearArt); got != "https://cdn/first-clearart.png" {
		t.Errorf("Slot(clearart) = %q, want the first candidate", got)
	}
	// A slot nothing supplied is empty, which a renderer reads as "fall back".
	if got := art.Slot(v1.ArtworkDisc); got != "" {
		t.Errorf("Slot(disc) = %q, want empty", got)
	}
	// An unrecognised slot is not an error. The vocabulary is open (ADR 0015),
	// so a source with a type this build has never heard of must not panic a
	// consumer that asks about it.
	if got := art.Slot(v1.ArtworkSlot("holographic")); got != "" {
		t.Errorf("Slot(unknown) = %q, want empty", got)
	}
}

// TestArtworkCandidatesFor covers what a picker screen reads.
func TestArtworkCandidatesFor(t *testing.T) {
	art := v1.Artwork{
		Candidates: []v1.ArtworkCandidate{
			{Slot: v1.ArtworkPoster, URL: "a", Source: "fanart-tv"},
			{Slot: v1.ArtworkBackdrop, URL: "b", Source: "tmdb"},
			{Slot: v1.ArtworkPoster, URL: "c", Source: "tmdb"},
		},
	}

	posters := art.CandidatesFor(v1.ArtworkPoster)
	if len(posters) != 2 {
		t.Fatalf("CandidatesFor(poster) = %d, want 2", len(posters))
	}
	// Order is preserved, because the set is stored best-first and a picker
	// showing them in a different order than selection used would be confusing.
	if posters[0].URL != "a" || posters[1].URL != "c" {
		t.Errorf("CandidatesFor(poster) = %v, want the stored order preserved", posters)
	}
	// Provenance survives the filter — it is what a "from fanart.tv" label reads.
	if posters[0].Source != "fanart-tv" {
		t.Errorf("Source = %q, want fanart-tv", posters[0].Source)
	}
	if got := art.CandidatesFor(v1.ArtworkDisc); got != nil {
		t.Errorf("CandidatesFor(disc) = %v, want nil", got)
	}
}

// TestArtworkEmpty covers the distinction ADR 0074 draws between a node with no
// art and a node whose art failed to resolve — the second has candidates and no
// selection, and reporting it as empty would hide a resolution bug.
func TestArtworkEmpty(t *testing.T) {
	if !(v1.Artwork{}).Empty() {
		t.Error("the zero value should be empty")
	}
	unresolved := v1.Artwork{
		Candidates: []v1.ArtworkCandidate{{Slot: v1.ArtworkPoster, URL: "https://cdn/p.jpg"}},
	}
	if unresolved.Empty() {
		t.Error("candidates with no selection is not empty — it has art that failed to resolve")
	}
	if (v1.Artwork{Logo: "https://cdn/l.png"}).Empty() {
		t.Error("a selection with no candidates is not empty")
	}
}
