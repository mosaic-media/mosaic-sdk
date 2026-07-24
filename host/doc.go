// Package host runs a Mosaic module as its own process, and lets the Platform
// talk to one (ADR 0064, ADR 0077).
//
// # Why this is a separate Go module
//
// The parent `sdk` module's go.mod is a module line and a Go version, and that
// is load-bearing rather than tidy: a third party compiles against the contract
// and against nothing the Platform happened to choose. ADR 0059 refused to
// re-export OpenTelemetry for exactly this reason, and a gRPC serving harness
// inside `sdk` would break the property by the same mechanism.
//
// So the harness lives here, in a nested module with its own go.mod. A module
// author importing only `sdk` gets the contract and can be tested with no
// transport at all. Adding `sdk/host` is what makes the module runnable as a
// process. The cost is one extra import and a nested-module tagging convention
// (host/vX.Y.Z alongside vX.Y.Z), which is a known Go wrinkle rather than a
// novel one.
//
// # Both sides live here
//
// This package is imported by the module *and* by the Platform, because a
// go-plugin plugin definition is inherently two-sided: the same handshake, the
// same plugin name and the same conversions have to be agreed by both ends, and
// a second copy in the Platform would be a second copy that can drift.
//
//   - A module calls [Serve] with its plain Go [v1.Capability]. Its main.go is
//     roughly `func main() { host.Serve(mymodule.New()) }` and nothing else in
//     the module changes — that is the property ADR 0064 is arranged around.
//   - The Platform uses [Plugin] with go-plugin's client to obtain something
//     implementing [v1.Capability] and whichever provider roles the module
//     serves. Its capability registry cannot tell that value from a local
//     struct.
//
// # What crossing the boundary does and does not change
//
// It changes [v1.Caller]. In process that value is meaningless outside the
// invocation; serialized, it becomes something a module can retain. So the
// Platform mints a handle per invocation and revokes it on return, and the
// module presents it on every callback. A module author never sees this: they
// forward the Caller they were given, exactly as ADR 0017 already required.
//
// It does not change error classification, and it would be easy to think it
// does because the wire carries a category. **A module cannot classify an error
// either way.** The Platform's error categories live in its own internal
// packages and are deliberately not published in the SDK, so in process a module
// receives an error whose category it cannot read, and out of process it
// receives the same. The category travels so the Platform end keeps it and so
// telemetry can record it — not so the module can branch on it. If a module ever
// should be able to, that is an SDK addition and this package already carries
// the value it would need.
//
// In the other direction the fidelity is exact: an error a module returns has no
// Platform category in process, where CategoryOf maps it to Internal, and it has
// none here either.
package host
