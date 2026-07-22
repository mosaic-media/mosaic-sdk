# Claude Instructions — Mosaic SDK

This repository is the **published contract surface** between the Platform and
the Modules that extend it (ADR 0008, ADR 0016). It is `github.com/mosaic-media/sdk`,
consumed as an ordinary tagged dependency with no `replace`.

## This is hand-written Go. It is not generated, and it is not protobuf.

**Read this before adding a file.** Mosaic has two published contract
repositories and they are built in opposite ways, which is a reasonable thing
to get wrong:

| Repository | Form | Source of truth |
|---|---|---|
| **`sdk`** (this one) | **hand-written Go** | the `.go` files in `contracts/platform/v1/` |
| **`sdui`** | **protobuf**, Go and TS generated | `proto/**/*.proto`, generated into `gen/` |

[ADR 0044](https://github.com/mosaic-media/architecture/blob/main/docs/adr/0044-contracts-protobuf-workspace.md)
made the **SDUI and session** contracts protobuf. Its title names that scope,
and it does not extend here.

The reason is not historical accident. **This SDK's job is Go interfaces with
behaviour** — `Capability`, `ContentService`, the provider roles, `Telemetry` —
which a module *implements* in its own process. Protobuf describes messages and
RPC services; it cannot express an interface a third party satisfies in-process.
`sdui` is the opposite case: a wire format, consumed by four client languages,
where codegen is exactly right.

So: add a `.go` file beside `capability.go` and `provider.go`. Do not add a
`.proto`, do not add a `buf.yaml`, and do not generate anything. There are no
generated files here and no build step — `go build ./...` is the whole thing.

## Non-negotiable rules

- **No dependencies.** `go.mod` is a module line and a Go version, and that is
  load-bearing: a third party compiles against this contract and against
  nothing the Platform happened to choose. Adding a dependency here forces it
  on every module author and pins them to a version the Platform picked.
  This is why the telemetry surface (ADR 0059) declares its own interface
  rather than re-exporting OpenTelemetry.
- **Nothing here imports the Platform.** The dependency points one way. If a
  capability needs a private Platform import, the contracts are not ready to
  publish — that is the stop point, and it governs any change here.
- **No storage contracts, no transaction type, no identity or configuration
  models.** A capability calls application services, never stores (ADR 0012).
- **Apache-2.0**, unlike the Platform's AGPL. This is the permissive surface a
  third party compiles against. Files here carry no SPDX header — match the
  files already present rather than importing the Platform's convention.

## Versioning and release

Pre-1.0 on purpose. A change is a **minor** bump (`v0.13.0` → `v0.14.0`), tagged
and pushed, and the Platform's `require` is bumped to match:

```bash
git tag v0.14.0 && git push origin main && git push origin v0.14.0
```

For local cross-repo work, add `replace github.com/mosaic-media/sdk => ../sdk`
to the Platform's `go.mod` temporarily — then tag, push, bump, and remove the
`replace` before committing. A `replace` must never land in a commit.

Update the **Status** section of `README.md` in the same change: it is the
per-version changelog, and it is how anyone finds out what a tag contains.

## Workflow

- Commit and push this repository **separately** from `platform`. It is its own
  git repository despite sitting beside the others on disk.
- **Commit author identity** must be `AdamNi-7080 <anicholls41@gmail.com>`.
- `gofmt`, `go build ./...`, `go vet ./...` and `go test ./...` before pushing.
- Every exported type and function carries a doc comment that says *why*, not
  only what. This is a published contract read by people who cannot read the
  Platform's source; the comments are the documentation.
