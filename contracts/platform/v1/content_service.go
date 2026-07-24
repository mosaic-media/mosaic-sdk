package v1

import "context"

// ContentService is the Platform's content application surface — everything a
// capability does to the object graph, and the interface it is handed rather
// than a concrete Platform type (ADR 0016). The Platform's application service
// implements it; a capability holds it.
//
// Every method authenticates the command's Caller, authorises it and — for
// the writes — commits state and an outbox event in one transaction. Errors
// carry a Platform error category (InvalidArgument, NotFound, Conflict,
// PermissionDenied, Unauthenticated, Unavailable, Internal).
type ContentService interface {
	// AddContentWork creates the root of a containment tree.
	AddContentWork(ctx context.Context, cmd AddContentWorkCommand) (AddContentWorkResult, error)
	// AddContentChild inserts a container or item beneath an existing node.
	AddContentChild(ctx context.Context, cmd AddContentChildCommand) (AddContentChildResult, error)
	// AttachContentPart attaches playable bytes to an item.
	AttachContentPart(ctx context.Context, cmd AttachContentPartCommand) (AttachContentPartResult, error)
	// SetContentArtwork replaces a node's stored artwork. It is the one write
	// here that updates an existing node rather than creating one.
	SetContentArtwork(ctx context.Context, cmd SetContentArtworkCommand) (SetContentArtworkResult, error)
	// RelateContent draws one edge of the association graph.
	RelateContent(ctx context.Context, cmd RelateContentCommand) (RelateContentResult, error)
	// BindContentSource records that a source resolves to a node.
	BindContentSource(ctx context.Context, cmd BindContentSourceCommand) (BindContentSourceResult, error)
	// ResolveContentBinding settles one entry in the review queue.
	ResolveContentBinding(ctx context.Context, cmd ResolveContentBindingCommand) (ResolveContentBindingResult, error)

	// SearchContent finds content by title, media type and kind.
	SearchContent(ctx context.Context, query SearchContentQuery) (SearchContentResult, error)
	// FindContentByExternalID resolves content by a provider identifier.
	FindContentByExternalID(ctx context.Context, query FindContentByExternalIDQuery) (FindContentByExternalIDResult, error)
	// GetContentNode reads one node, optionally with its direct children.
	GetContentNode(ctx context.Context, query GetContentNodeQuery) (GetContentNodeResult, error)
	// ListContentParts reads an item's playable parts — the read side of
	// AttachContentPart, absent until a re-import needed to know what it had
	// already stored.
	ListContentParts(ctx context.Context, query ListContentPartsQuery) (ListContentPartsResult, error)

	// Playback state (ADR 0046). These are the first per-user methods on this
	// surface: everything above operates on an install-global graph, and a
	// position belongs to a person. A consumer records progress as its invoking
	// user, so a module can never write another user's position.

	// RecordPlaybackProgress reports where a viewer has got to.
	RecordPlaybackProgress(ctx context.Context, cmd RecordPlaybackProgressCommand) (RecordPlaybackProgressResult, error)
	// SetPlaybackFinished marks an item watched or unwatched explicitly,
	// overriding the derived threshold in either direction.
	SetPlaybackFinished(ctx context.Context, cmd SetPlaybackFinishedCommand) (SetPlaybackFinishedResult, error)
	// GetPlaybackState reads one viewer's position in one item.
	GetPlaybackState(ctx context.Context, query GetPlaybackStateQuery) (GetPlaybackStateResult, error)
	// ListPlaybackStates reads state for several items at once, so a season of
	// episodes costs one query rather than one per row.
	ListPlaybackStates(ctx context.Context, query ListPlaybackStatesQuery) (ListPlaybackStatesResult, error)
	// ListInProgress reads what a viewer has started and not finished, most
	// recent first — the continue-watching list, as a query rather than a
	// client-side fold so every client gets it identically.
	ListInProgress(ctx context.Context, query ListInProgressQuery) (ListInProgressResult, error)
}
