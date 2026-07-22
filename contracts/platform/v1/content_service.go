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
}
