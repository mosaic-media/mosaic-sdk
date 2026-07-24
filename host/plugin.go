package host

import (
	"context"

	goplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	modulev1 "github.com/mosaic-media/contracts/gen/mosaic/module/v1"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// PluginName is the key both ends use in go-plugin's plugin map. There is
// exactly one plugin per module process: a module is a single Capability, and
// the provider roles are methods on it rather than separate plugins.
const PluginName = "capability"

// callbackBrokerID is the go-plugin broker stream the Platform serves its
// ContentService and Telemetry on, and the module dials for callbacks.
//
// It is a constant rather than an id negotiated per invocation, and that is a
// deliberate simplification worth explaining. go-plugin's broker multiplexes
// many streams over one connection, each identified by a number both ends agree
// on; the usual way to agree is to put the id in the request. That would mean a
// broker id field on every request that can call back.
//
// It is unnecessary here because **the connection is not what scopes
// authority — the Caller handle is** (ADR 0064). One long-lived callback
// connection per module process, with a handle minted and revoked per
// invocation, gives exactly the property the per-invocation design was reaching
// for: a retained connection is useless without a live handle, and a handle
// stops resolving the instant the invocation returns.
//
// There is one broker per plugin connection, so a fixed id cannot collide.
const callbackBrokerID = 1

// SDKMajor is the SDK major version this harness speaks, and it is the whole
// compatibility story (ADR 0064). A module and a Platform are compatible when
// they share an SDK major, so there is one number a user reasons about rather
// than two — the proto package version tracks it, and go-plugin's
// ProtocolVersion below carries it on the wire.
//
// While the SDK is pre-1.0 the compatibility unit is the *minor* version, which
// is effectively exact pinning. That is correct rather than unfortunate: Go
// gives v0.x no compatibility guarantee, and there are no third-party authors
// yet to inconvenience. Reaching SDK v1.0 is a precondition for a third-party
// ecosystem, not for building this tier.
const SDKMajor = 0

// Handshake is go-plugin's mutual identification. It is not security — the
// magic cookie is compiled into both sides and public — it is what makes a
// process launched by mistake fail with a clear message instead of hanging on a
// protocol it does not speak.
//
// ProtocolVersion is the SDK major. Bumping it refuses every module built
// against the previous major, which is the intended behaviour: a major bump is
// where the contract broke.
var Handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  SDKMajor + 1, // go-plugin requires a non-zero version.
	MagicCookieKey:   "MOSAIC_MODULE",
	MagicCookieValue: "mosaic-module-v1",
}

// Plugin is the two-sided plugin definition: the same handshake, plugin name
// and conversions agreed by both ends, in one place so there is no second copy
// to drift.
//
// The two sides populate different fields. A module sets Impl (via [Serve]) and
// nothing else. The Platform sets Content, Telemetry and CategoryOf, and leaves
// Impl nil.
type Plugin struct {
	goplugin.NetRPCUnsupportedPlugin

	// Impl is the module author's plain Go Capability. Nothing about it knows
	// it is being served over a socket. Module side only.
	Impl v1.Capability

	// Content and Telemetry are what the Platform serves back to the module
	// over the broker. Platform side only.
	Content   v1.ContentService
	Telemetry v1.Telemetry

	// CategoryOf lets the Platform name the category of an error it produced,
	// without this package having to know the vocabulary. See errors.go.
	// Platform side only; nil is valid and means categories are not reported.
	CategoryOf CategoryFunc
}

var _ goplugin.GRPCPlugin = (*Plugin)(nil)

// GRPCServer runs in the module process. It registers the capability server on
// the server go-plugin created, handing it the broker so callbacks can dial
// back to the Platform.
func (p *Plugin) GRPCServer(broker *goplugin.GRPCBroker, s *grpc.Server) error {
	modulev1.RegisterCapabilityServiceServer(s, &capabilityServer{
		impl:   p.Impl,
		broker: broker,
	})
	return nil
}

// GRPCClient runs in the Platform process. It starts serving ContentService and
// Telemetry on the callback stream, then returns a value implementing
// [v1.Capability] so the capability registry holds it exactly as it holds a
// compiled-in module.
func (p *Plugin) GRPCClient(ctx context.Context, broker *goplugin.GRPCBroker, c *grpc.ClientConn) (any, error) {
	// AcceptAndServe blocks until the module dials, so it runs in its own
	// goroutine. It returns when the connection closes, which is process
	// shutdown.
	go broker.AcceptAndServe(callbackBrokerID, func(opts []grpc.ServerOption) *grpc.Server {
		s := grpc.NewServer(opts...)
		modulev1.RegisterContentServiceServer(s, &contentServer{
			impl:       p.Content,
			categoryOf: p.CategoryOf,
		})
		modulev1.RegisterTelemetryServiceServer(s, &telemetryServer{
			impl:       p.Telemetry,
			categoryOf: p.CategoryOf,
		})
		return s
	})

	return &capabilityClient{client: modulev1.NewCapabilityServiceClient(c)}, nil
}

// ServePluginMap is what a module passes to go-plugin. [Serve] uses it; a
// module author does not call it directly.
func ServePluginMap(impl v1.Capability) map[string]goplugin.Plugin {
	return map[string]goplugin.Plugin{PluginName: &Plugin{Impl: impl}}
}

// ClientPluginMap is what the Platform passes to go-plugin's client. The
// services given here are what the module calls back into, and every call it
// makes re-authorises as the invoking user (ADR 0017).
func ClientPluginMap(content v1.ContentService, telemetry v1.Telemetry, categoryOf CategoryFunc) map[string]goplugin.Plugin {
	return map[string]goplugin.Plugin{
		PluginName: &Plugin{
			Content:    content,
			Telemetry:  telemetry,
			CategoryOf: categoryOf,
		},
	}
}
