package v1

// The content read queries. Like the commands, each authenticates its Caller
// and passes policy before reading; unlike them, a query opens no
// transaction.

// SearchContentQuery finds content by title, media type and kind. Every
// filter is optional; the empty query returns the first page of everything.
// This is "do I already have this?" — the read a capability makes before it
// sources anything.
type SearchContentQuery struct {
	Caller Caller
	// Title matches case-insensitively anywhere in the title.
	Title     string
	MediaType MediaType
	Kind      NodeKind
	// AttributesContain narrows to nodes whose Attributes document *contains*
	// this JSON document. Empty means no filter.
	//
	// # Why containment, and not a typed filter
	//
	// A node's Attributes are module-owned opaque JSON: the Platform stores them
	// and never interprets them (ADR 0013 assigns their correctness to the
	// writing capability). A typed filter would require the Platform to learn
	// what a module put there, which is the coupling the whole arrangement
	// exists to avoid. Containment is the one question that can be asked without
	// understanding the answer — *does this document contain this shape* — so it
	// is the only filter that respects the ownership.
	//
	// That makes it the counterpart of FindContentByExternalID, which has always
	// worked this way over the neighbouring external-ids document.
	//
	// # What a caller should expect
	//
	//	// works a module recorded as available on a given service
	//	AttributesContain: []byte(`{"watch":{"providers":["Netflix"]}}`)
	//
	// Matching is structural and nested: a document containing the key with a
	// superset of the listed array entries matches. It must be a valid JSON
	// document; anything else is an InvalidArgument rather than an empty result,
	// because a filter that silently matched nothing would read as "you own none
	// of these" rather than as a mistake.
	//
	// **It is a storage-engine obligation, not a Postgres detail.** Any
	// StorageAdapter implementation must support containment over both JSON
	// documents; the Platform's own does it with an indexed jsonb operator.
	//
	// **A module can see what another module wrote.** That is deliberate and
	// matches the rest of the read surface — providers are resolvable across
	// modules by design (ADR 0027) — but it means a module should treat its
	// attribute keys as a published shape rather than a private one.
	AttributesContain []byte
	// Limit is clamped to a Platform maximum and defaults when zero or
	// negative, so a search cannot ask for an unbounded scan.
	Limit int
}

// SearchContentResult carries the matching nodes.
type SearchContentResult struct {
	Nodes []Node
}

// FindContentByExternalIDQuery resolves content by a provider's own
// identifier — the strong form of "do I already have this", not depending on
// titles matching.
type FindContentByExternalIDQuery struct {
	Caller Caller
	Scheme string
	Value  string
}

// FindContentByExternalIDResult carries the matches. It is a list because an
// anime and its source manga can share a provider reference and remain two
// Works (ADR 0013).
type FindContentByExternalIDResult struct {
	Nodes []Node
}

// GetContentNodeQuery reads a single node, optionally with its direct
// children — one level, not a subtree, since variable depth means a caller
// descends deliberately (ADR 0013).
type GetContentNodeQuery struct {
	Caller       Caller
	NodeID       NodeID
	WithChildren bool
}

// GetContentNodeResult carries the node and, when asked, its direct children
// in order. Children is nil unless requested and empty for a childless node.
type GetContentNodeResult struct {
	Node     Node
	Children []Node
}

// ListContentPartsQuery reads the playable parts of an item node.
//
// It closes a hole the surface has carried since the content model landed:
// AttachContentPart could write a Part and nothing could read one back, so a
// capability could not see what it had itself created. ADR 0045 named it while
// building the first consumer; this is what finally needed it — a re-import has
// to know which releases are already stored before it can add the ones that are
// not.
type ListContentPartsQuery struct {
	Caller Caller
	// NodeID is the item whose parts to read. A work or container has none of
	// its own.
	NodeID NodeID
}

// ListContentPartsResult carries the node's parts in natural order, so the
// segments of a multi-disc edition come back in sequence and a candidate set
// comes back in the order its source ranked it.
type ListContentPartsResult struct {
	Parts []Part
}
