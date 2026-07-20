package v1

import "context"

// Capability is what an optional Module implements so the Platform can invoke
// it. ADR 0008 always reserved a capability and registration surface for the
// SDK; ADR 0016 populated only the content services, and this is the first
// addition of the capability side.
//
// The dependency direction is the whole point: a Module depends only on this
// SDK, and the Platform holds the Capability, registers it and routes to it.
// The Module never imports the Platform. A Capability sources content from
// somewhere the Platform does not know about — a provider, an addon, a
// scanner — and reflects it into the object graph through ContentService,
// acting as the Caller it is handed (ADR 0017). It owns no schema (ADR 0012):
// everything it does to the graph goes through the published services.
type Capability interface {
	// Manifest is the Module's self-declaration — the stable identity the
	// Platform registers it under and routes imports to.
	Manifest() Manifest

	// Import sources the content named by query and reflects it into the
	// Platform through svc, acting as caller throughout. It returns what it
	// did so the invoker (or a test) can see the shape without re-reading the
	// graph. Errors carry a Platform error category, since every service call
	// it makes does.
	Import(ctx context.Context, svc ContentService, caller Caller, query string) (ImportResult, error)
}

// Manifest is a Module's declaration of identity. It is deliberately minimal —
// an id the Platform registers and routes by, a version, and a human-readable
// name. It grows as the module system needs it (the permissions a module
// declares, the media types it sources); those are named future additions,
// not omissions to fix now.
type Manifest struct {
	// ID is the stable key the Platform registers the capability under and a
	// caller names to invoke it. It must be unique across registered modules.
	ID string
	// Version is the module's own version, carried for diagnostics and future
	// compatibility checks. It is the module's version, not the SDK's.
	Version string
	// Name is a human-readable label for the module.
	Name string
}

// ImportResult reports what an Import did. Containers, Items and Parts count
// the nodes and parts the import created; AlreadyKnown is true when the
// content was already present, in which case the import was a no-op that
// returned the existing work's id in WorkID.
type ImportResult struct {
	WorkID       NodeID
	AlreadyKnown bool
	Containers   int
	Items        int
	Parts        int
}
