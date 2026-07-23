package v1

import "time"

// The content write commands. Each is a Platform application service call a
// capability makes to change the object graph; the Platform validates,
// authenticates the Caller, authorises and commits (ADR 0016). A command
// carries only what a caller legitimately chooses — the Platform derives ids,
// timestamps and structural invariants.

// AddContentWorkCommand creates the root of a containment tree — a film, a
// series, an album, an artist, a collection (ADR 0013). A work roots its own
// tree, so its id and work id are the Platform's to set, not the caller's.
type AddContentWorkCommand struct {
	Caller    Caller
	MediaType MediaType
	Title     string
	// ExternalIDs and Attributes are optional JSON documents. Empty means an
	// empty object; the schema does not validate their contents (ADR 0013).
	ExternalIDs []byte
	Attributes  []byte
	// Artwork is the node's stored image URLs (ADR 0071). A materialising
	// capability fills it from the metadata it already fetched so a list surface
	// need not re-derive art per card; empty is "the source had none".
	Artwork Artwork
}

// AddContentWorkResult carries the committed work, whose MediaType is the
// canonical form (ADR 0015) and may differ from what was sent.
type AddContentWorkResult struct {
	Work Node
}

// AddContentChildCommand adds a container or item beneath an existing node.
// A child inherits its work id and media type from its parent, so neither is
// the caller's to choose (ADR 0013).
type AddContentChildCommand struct {
	Caller   Caller
	ParentID NodeID
	// Kind must be container or item. A work is never a child.
	Kind NodeKind
	// ContainerType applies when Kind is container, ItemType when it is item;
	// the other must be empty.
	ContainerType ContainerType
	ItemType      ItemType
	Title         string
	// NaturalOrder places the child among its siblings, a float so an
	// insertion does not renumber the rest (ADR 0013).
	NaturalOrder float64
	ExternalIDs  []byte
	Attributes   []byte
	// Artwork is the child's stored image URLs (ADR 0071) — for an episode, its
	// still. Empty is "the source had none".
	Artwork Artwork
}

// AddContentChildResult carries the committed child.
type AddContentChildResult struct {
	Node Node
}

// AttachContentPartCommand attaches playable bytes to an item node. A Part
// points at bytes and never contains them (ADR 0014), so the command carries
// a location, not media.
type AttachContentPartCommand struct {
	Caller Caller
	// NodeID must be an item — a work or container has nothing to play.
	NodeID       NodeID
	Role         PartRole
	EditionLabel string
	NaturalOrder float64
	Location     MediaLocation
	// Technical metadata is optional; the zero value means "not known".
	Container  string
	VideoCodec string
	AudioCodec string
	Width      int
	Height     int
	HDRFormat  string
	Duration   time.Duration
	BitrateBPS int64
	SizeBytes  int64
	Attributes []byte
}

// AttachContentPartResult carries the committed part.
type AttachContentPartResult struct {
	Part Part
}

// RelateContentCommand draws a typed, directed edge between two works
// (ADR 0013) — an adaptation, a sequel, a collection membership.
type RelateContentCommand struct {
	Caller     Caller
	FromNodeID NodeID
	ToNodeID   NodeID
	Type       RelationType
	// Confidence is between 0 and 1; Origin says where the assertion came
	// from, which is what makes a low confidence actionable.
	Confidence float64
	Origin     RelationOrigin
}

// RelateContentResult carries the committed edge.
type RelateContentResult struct {
	Relation Relation
}

// BindContentSourceCommand ties an external source to a node with an explicit
// confidence (ADR 0013). Status is confirmed for a strong match or
// pending_review to queue a weak one; a binding is never created rejected.
type BindContentSourceCommand struct {
	Caller          Caller
	NodeID          NodeID
	SourceProvider  string
	SourceRef       string
	MatchConfidence float64
	MatchMethod     MatchMethod
	Status          BindingStatus
}

// BindContentSourceResult carries the committed binding.
type BindContentSourceResult struct {
	Binding SourceBinding
}

// BindingResolution is the decision a reviewer makes about a pending binding.
type BindingResolution string

const (
	// ResolveConfirm settles a binding against its node — a merge.
	ResolveConfirm BindingResolution = "confirm"
	// ResolveReject declines the match, keeping the row so the same weak
	// match is not proposed again.
	ResolveReject BindingResolution = "reject"
)

// ResolveContentBindingCommand acts on a binding under review (ADR 0013). A
// merge is Confirm, a rejection is Reject, and a split is Confirm with
// MoveToNodeID set — the binding moves and the source is never re-resolved.
type ResolveContentBindingCommand struct {
	Caller     Caller
	BindingID  SourceBindingID
	Resolution BindingResolution
	// MoveToNodeID re-targets the binding before confirming — a split. It is
	// only valid with Confirm.
	MoveToNodeID NodeID
}

// ResolveContentBindingResult carries the updated binding.
type ResolveContentBindingResult struct {
	Binding SourceBinding
}
