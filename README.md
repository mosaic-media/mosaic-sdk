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

Pre-1.0 on purpose: the surface still changes as modules find its gaps.

## License

Apache License, Version 2.0 (see [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE)).
The SDK is deliberately permissive: it is the contract a Module builds against,
so a Module author may use it under any license. This is independent of the
Platform's license (AGPL-3.0 with a Module Linking Exception).
