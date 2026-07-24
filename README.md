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

**`v0.21.0` turns artwork into a candidate set and adds the eighth role** —
`ArtworkCandidate`, `ArtworkSlot`, `Artwork.Candidates`, `RoleArtwork`,
`ArtworkProvider` and `ExternalIdentity`
([ADR 0074](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0074-artwork-is-a-candidate-set.md),
[ADR 0075](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0075-the-artwork-provider-role.md)).

`Artwork` held one URL per slot, so a source that returns forty posters — which
is what a dedicated artwork database *is* — had thirty-nine discarded, and a
user could never be offered a choice because the alternatives never left the
module. The four flat slots stay and keep their type; what changes is their
**meaning**, from "the artwork the provider gave" to "the artwork that was
*chosen*", with `Candidates` holding what it was chosen from. That is what keeps
the change additive: every consumer still reads `Poster`, and there is still
exactly one answer to "what is this node's poster".

A candidate carries `Source`, `Language` and `Rank` because those are what a
choice gets made on. Two of them are easy to misread and are documented at
length: an **empty `Language` means textless**, which is a positive property and
usually the right backdrop to sit under a clearlogo; and **`Rank` is not
comparable across sources**, since one source's likes and another's vote average
measure different things — it orders within a source and nowhere else.

`RoleArtwork` is neither a source role nor a consumer role: it enriches content
that already exists. Its doc comment says plainly that it does **not** satisfy
[ADR 0035](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0035-metadata-as-required-capability.md)'s
metadata requirement, because the tempting shortcut — declaring `RoleMetadata`
to reach `ContentMetadata`'s image fields — would pass the composition-root
check with a module that cannot name a film.

**`v0.20.0` adds `StreamRequest.Season` and `StreamRequest.Episode`** — neutral
coordinates for an episode, so a stream provider can be asked about content it
did not source ([ADR 0073](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0073-stream-resolution-is-decoupled-from-metadata-provenance.md)).

The Ref in that case carries a *shared* external identity (`imdb`, `tvdb`)
rather than the provider's own id, and the provider composes its native
addressing from the identity plus these two numbers. That is the point: one real
source addresses an episode as `tt0903747:1:2`, and a Platform that built that
string would have a provider's dialect in its kernel. Season and episode are
facts about television the Platform already models; the format is the module's
business.

`StreamProvider`'s doc says plainly what follows — a provider asked about
content it does not recognise returns an empty response and **no error**, since
failing there would fail a user's import over a title some other source
described.

**`v0.19.0` adds `SearchContentQuery.AttributesContain`** — a containment filter
over a node's module-owned `Attributes` document, so a capability can ask
"which of my works did some module tag *this* way".

Containment rather than a typed filter, because attributes are opaque to the
Platform by design (ADR 0013 assigns their correctness to the writing
capability). A typed filter would make the Platform learn what a module put
there, which is the coupling the arrangement exists to avoid; *does this
document contain this shape* is the one question that can be asked without
understanding the answer. It is the counterpart of `FindContentByExternalID`,
which has always worked this way over the neighbouring external-ids document.

It is a **storage-engine obligation**: any `StorageAdapter` must support
containment over both JSON documents. And a module can see what another module
wrote — deliberate, matching the rest of the read surface, but it means a
module's attribute keys are a published shape rather than a private one.

**`v0.18.0` adds `ContentMetadata.Watch`** — where a title can be watched
*outside* Mosaic, in one region, with `WatchAvailability`, `WatchOffer` and
`WatchOfferType` as new types.

It is the first read field that does not describe the title itself, and the
doc comments lean hard on one distinction because it is the easy thing to get
wrong: **an offer is not a source.** Every other read role answers "what can
Mosaic get you"; this answers "where else does this exist". Nothing in it
becomes a `Part`, nothing in it is playable through the Platform, and a client
that renders an offer as a play control is making a promise the Platform cannot
keep. `Link` is where an informational control should send the viewer.

Two smaller shapes follow from what the data is. It is **strictly regional**, so
a value names the one `Region` it describes and empty offers mean "none known
*here*" rather than "unavailable". And it carries an `Attribution` string,
because the upstreams that compile availability data generally require being
credited wherever it is shown — carrying it in the contract is what keeps the
Platform from having to know which upstream imposed that.

**`v0.17.0` grows `ContentMetadata` for a source that models more than one
title at a time** — `Keywords`, `Certification`, `Similar []RelatedItem`,
`Collection *Collection` and `Trailers []Trailer`, with `RelatedItem`,
`Collection` and `Trailer` as new types.

Three of those close gaps
[ADR 0034](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0034-rich-metadata-preview.md)
recorded rather than invented: a franchise ("the other Avatar films"), related
titles, and the finer-grained tags that make "more like this" mean anything. The
addon protocol carries none of them, so the fields waited for a source that does.

Two shape decisions are worth stating. **`Collection` is a descriptive
projection, not the object graph's collection** — the graph expresses membership
as a `RelationCollectionMember` edge between Works, and an edge needs two Works
to exist, which a virtual item is not; a detail screen must render a franchise
before anything has been materialised. And **`Trailer` carries a site and that
site's key rather than a URL**, because assembling a watch or embed URL is an
embed-policy decision that belongs to the client rather than to the contract.

`Certification` is display-only text for the same reason `Runtime` is: age
scales are national and not comparable, so the label is carried rather than
mapped onto an invented common scale. Empty means unknown, which a consumer must
not read as "suitable for everyone".

**`v0.16.0` adds `Artwork.Landscape`** — wide 16:9 key art, distinct from
`Backdrop`: a backdrop is scenery to sit *behind* a hero, this is a composed card
image to sit *in* one, which is what a resume rail wants. Added because a real
source supplies it and it was being dropped — Cinemeta has no such field, but an
addon proxying an artwork database returns one alongside the poster. Empty rather
than substituted when a source has none.

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
