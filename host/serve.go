package host

import (
	goplugin "github.com/hashicorp/go-plugin"

	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

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
func Serve(capability v1.Capability) {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins:         ServePluginMap(capability),

		// gRPC rather than net/rpc. It is heavier, and it keeps the door open
		// to a module written in a language other than Go, which net/rpc would
		// close permanently (ADR 0077).
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}
