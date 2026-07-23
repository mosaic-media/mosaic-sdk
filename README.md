# Mosaic SDK

The public contract language between the Mosaic Platform and the Modules that
extend it ([ADR 0008](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0008-sdk-as-public-contract-language.md)).
A Module compiles against this module and nothing else of the Platform's.

It is deliberately small. Today it carries the **content surface**
(`contracts/platform/v1`): the object-graph models (`Node`, `Part`,
`Relation`, `SourceBinding` and their vocabularies), the content command,
query and result types, the `ContentService` interface a capability calls, the
**provider roles** a module declares in `Manifest.Provides` — source roles that
populate the virtual plane (metadata, search, catalog, stream, subtitles) and
the consumer role that acts on the materialised library (playback) — and
the opaque `Caller` a capability forwards from its invocation context
([ADR 0016](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0016-published-contract-surface.md),
[ADR 0017](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0017-how-a-capability-acts.md)).

It holds **no** storage contracts, no transaction type, no identity or
configuration models, and no Platform implementation — a capability calls
application services, never stores. It depends only on the Go standard
library.

```go
import v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
```

## Hand-written Go, not generated

Mosaic has two published contract repositories built in opposite ways, and it
is worth stating which is which. **This one is hand-written Go**; there are no
`.proto` files, no codegen and no build step. [`sdui`](https://github.com/mosaic-media/sdui)
is the protobuf one, with Go and TypeScript generated from `proto/`
([ADR 0044](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0044-contracts-protobuf-workspace.md),
which is scoped to the SDUI and session contracts).

The split follows what each contract *is*. This SDK is Go interfaces with
behaviour — `Capability`, `ContentService`, the provider roles, `Telemetry` —
which a module implements in its own process, and which protobuf cannot
express. `sdui` is a wire format consumed by several client languages, where
codegen is exactly right.

## Build and test

**Everything runs in a container; nothing is built or tested on the host.**

```bash
docker compose -f docker-compose.test.yml run --rm test
```

That runs gofmt, `go build`, `go vet` and `go test` against a pinned toolchain.
There is no hidden dependency and `go build ./...` really is the whole thing —
the container is here because **this is the surface a third party compiles
against**, and the only claim worth making about it is that it builds under a
pinned toolchain rather than under whatever a given machine happens to have.

## Status

Extracted from `platform` into a standalone module and published. The Platform
and modules build against it as an ordinary tagged dependency, with no
`replace`.

`v0.1.0` carried the content surface; `v0.2.0` added the `Capability`
interface; `v0.3.0` added the `ImportRequest` that hands a module its settings;
`v0.4.0` added the source provider roles and the virtual content DTOs;
`v0.5.0` grew `ContentMetadata` into the rich detail surface; `v0.7.0` added
the subtitles role and richer `StreamLink`, and `v0.8.0` the settings-UI role.

**`v0.9.0` opens the consumer side** — `RolePlayback` and `PlaybackProvider`
([ADR 0045](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0045-playback-consumer-and-media-origin.md)),
the first role that *acts on* the materialised library rather than sourcing
into it. It is one method: a provider resolves a `Part` to a playable location
and never serves bytes, because the Platform owns transports. The `Served`
variant (a module producing bytes for the Platform to serve) and its `Open`
method arrive with the torrent engine that needs them.

**`v0.10.0`** adds `ListContentParts` — the read side of `AttachContentPart`,
missing since the content model landed. A capability could write a Part and
never read one back, so it could not see what it had itself created. A
re-import needing to know which releases were already stored is what finally
forced it.

**`v0.15.0` stores artwork on the node** — an `Artwork` value (poster,
backdrop, logo) on `Node` and on `AddContentWorkCommand` / `AddContentChildCommand`
([ADR 0071](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0071-content-artwork-is-stored-on-the-node.md)).
Descriptive metadata is otherwise re-derived live from the provider (ADR 0034);
artwork alone is written at materialisation and read back, because it is
rendered in bulk on list surfaces like the continue-watching rail — one node
read instead of a metadata round-trip per card — and because it is the one piece
of art a user may later want to override, which is possible only for something
the library owns.

**`v0.14.0` opens the per-user tier** — playback state
([ADR 0046](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0046-playback-state-is-platform-owned.md)):
`RecordPlaybackProgress`, `SetPlaybackFinished`, and the reads behind resume, a
watched mark and a continue-watching list. Every other method on this surface
operates on an install-global graph; a position belongs to a person, and this is
the first thing here that differs between two users of one Mosaic. It is keyed
by node rather than Part on purpose — a viewer resumes an episode, not the
1080p release of an episode — and it lives on the Platform rather than in a
playback module so it survives swapping one, and so anything that has no
business resolving bytes can still read progress.

**`v0.13.0` gives modules a voice** — `Telemetry`, reached with
`TelemetryFrom(ctx)`, and the redaction-classed `Field` that crosses with it
([ADR 0059](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0059-modules-observe-through-the-sdk.md)).
Before it, a module could return an error or print to the Platform's stdout,
and neither could be filtered, correlated or classified. The Platform owns the
observability plane and implements the interface; a module emits and configures
nothing. `TelemetryFrom` never returns nil, so a module records nothing and
works unchanged on a Platform that provides none.

Pre-1.0 on purpose: the surface still changes as modules find its gaps.

## License

Apache License, Version 2.0 (see [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE)).
The SDK is deliberately permissive: it is the contract a Module builds against,
so a Module author may use it under any license. This is independent of the
Platform's license (AGPL-3.0 with a Module Linking Exception).
