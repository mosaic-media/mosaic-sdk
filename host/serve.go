package host

import (
	"encoding/json"
	"os"

	goplugin "github.com/hashicorp/go-plugin"

	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// ManifestFlag, when passed to a module binary, makes it print its manifest as
// JSON and exit rather than serve. A module's release uses it to learn the
// module's own identity — id, version, name, roles — without hardcoding it in a
// workflow where it would drift from the code. It is the module's Manifest()
// method rendered to JSON, which is the single source of truth for that
// identity, so a role added in Go appears in the published manifest with no
// second edit.
const ManifestFlag = "--mosaic-manifest"

// manifestDoc is the JSON shape ManifestFlag prints — the SDK-level identity a
// module knows about itself. The distribution manifest a repository serves adds
// the schema, the SDK major and the per-platform binary digests, which the
// build has and the module does not; `modulesign build-manifest` combines the
// two.
type manifestDoc struct {
	ID       string   `json:"id"`
	Version  string   `json:"version"`
	Name     string   `json:"name"`
	Provides []string `json:"provides"`
}

// Serve runs a module as a plugin process and blocks until the Platform
// disconnects. It is the whole of a module's main.go:
//
//	func main() { host.Serve(mymodule.New()) }
//
// Everything else the author wrote stays exactly as it was — the plain Go
// [v1.Capability], its provider roles, its tests with no transport at all. That
// property is what ADR 0064 is arranged around, and it is why moving a module
// between tiers is a build change rather than a rewrite.
//
// Serve does not return in normal operation. go-plugin writes the handshake to
// stdout, so a module must not print to stdout itself; anything written there
// corrupts the handshake. Use the [v1.Telemetry] reached from the context — that
// is what it is for, and it reaches the Platform's observability plane rather
// than a stream nobody is reading.
//
// The one thing Serve prints to stdout deliberately is the manifest, and only
// when invoked with [ManifestFlag] — a mode that never serves, so there is no
// handshake to corrupt.
func Serve(capability v1.Capability) {
	if emitManifestIfAsked(capability) {
		return
	}

	// Route all of this process's egress through the Platform's proxy before
	// serving, so a module's first outbound call is already covered (ADR 0064).
	// Modules build their HTTP clients lazily, at invocation time, well after
	// this runs.
	configureEgressProxy()

	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         ServePluginMap(capability),

		// gRPC rather than net/rpc. It is heavier, and it keeps the door open
		// to a module written in a language other than Go, which net/rpc would
		// close permanently (ADR 0077).
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}

// emitManifestIfAsked prints the manifest and reports true when ManifestFlag is
// present, so Serve returns instead of serving. It exits the process rather than
// only returning, so a module whose main does more after Serve still stops here
// — the flag means "tell me what you are and stop", not "and then run".
func emitManifestIfAsked(capability v1.Capability) bool {
	for _, arg := range os.Args[1:] {
		if arg != ManifestFlag {
			continue
		}
		m := capability.Manifest()
		doc := manifestDoc{ID: m.ID, Version: m.Version, Name: m.Name}
		for _, r := range m.Provides {
			doc.Provides = append(doc.Provides, string(r))
		}
		if err := json.NewEncoder(os.Stdout).Encode(doc); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	return false
}
