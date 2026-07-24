package host

import (
	"context"
	"fmt"
	"sync"

	modulev1 "github.com/mosaic-media/contracts/gen/mosaic/module/v1"
	v1 "github.com/mosaic-media/sdk/contracts/platform/v1"
)

// The telemetry bridge (ADR 0059). A module records through the SDK's own
// interface — which is why this crosses a process boundary at all: ADR 0059
// declared an interface rather than re-exporting OpenTelemetry, so the module
// calls the same methods over a different transport and its surface is
// unaffected.
//
// **One fidelity loss, stated rather than hidden.** [v1.Field]'s Value is `any`;
// on the wire it is a string, rendered with fmt.Sprint. A sink that would have
// written 42 as a number writes "42" instead. Carrying a typed oneof would
// preserve it, and is not worth the schema surface for an observational channel
// whose fields are already heterogeneous — the SDK's own Duration() helper
// flattens to a string before it ever reaches here. If a sink ever needs the
// type, this is the place that dropped it.

func fieldToWire(f v1.Field) *modulev1.Field {
	return &modulev1.Field{
		Key:       f.Key,
		Value:     fmt.Sprint(f.Value),
		Redaction: redactionToWire(f.Redaction),
	}
}

func fieldsToWire(fs []v1.Field) []*modulev1.Field {
	out := make([]*modulev1.Field, 0, len(fs))
	for _, f := range fs {
		out = append(out, fieldToWire(f))
	}
	return out
}

func fieldsFromWire(fs []*modulev1.Field) []v1.Field {
	out := make([]v1.Field, 0, len(fs))
	for _, f := range fs {
		if f == nil {
			continue
		}
		out = append(out, v1.Field{
			Key:       f.GetKey(),
			Value:     f.GetValue(),
			Redaction: redactionFromWire(f.GetRedaction()),
		})
	}
	return out
}

// ─── Module side ────────────────────────────────────────────────────────────

// telemetryClient implements [v1.Telemetry] by calling the Platform back.
//
// Logging is fire-and-forget: a telemetry failure must never fail the operation
// being observed, so errors are dropped rather than returned. The SDK's
// signatures say the same thing — Debug/Info/Warn/Error return nothing, and a
// module that could not record something has no recovery to perform.
type telemetryClient struct {
	client modulev1.TelemetryServiceClient
}

var _ v1.Telemetry = (*telemetryClient)(nil)

func (t *telemetryClient) log(level modulev1.LogLevel, message string, fields []v1.Field) {
	//nolint:errcheck // deliberate: see the type comment.
	_, _ = t.client.Log(context.Background(), &modulev1.LogRequest{
		Level:   level,
		Message: message,
		Fields:  fieldsToWire(fields),
	})
}

func (t *telemetryClient) Debug(message string, fields ...v1.Field) {
	t.log(modulev1.LogLevel_LOG_LEVEL_DEBUG, message, fields)
}

func (t *telemetryClient) Info(message string, fields ...v1.Field) {
	t.log(modulev1.LogLevel_LOG_LEVEL_INFO, message, fields)
}

func (t *telemetryClient) Warn(message string, fields ...v1.Field) {
	t.log(modulev1.LogLevel_LOG_LEVEL_WARN, message, fields)
}

func (t *telemetryClient) Error(message string, fields ...v1.Field) {
	t.log(modulev1.LogLevel_LOG_LEVEL_ERROR, message, fields)
}

// Span opens a span on the Platform and returns a handle to it. The returned
// context carries the span id so a nested Span becomes a child rather than a
// sibling — which is the property the SDK's doc comment promises and the reason
// the context must be passed down.
func (t *telemetryClient) Span(ctx context.Context, name string, attrs ...v1.Field) (context.Context, v1.Span) {
	resp, err := t.client.StartSpan(ctx, &modulev1.StartSpanRequest{
		Name:         name,
		Attributes:   fieldsToWire(attrs),
		ParentSpanId: parentSpanFrom(ctx),
	})
	if err != nil {
		// A span that could not be opened must not break the work it measures.
		// The no-op behaves exactly as the SDK's own fallback for a module
		// running outside a Platform.
		return ctx, noopSpan{}
	}
	id := resp.GetSpanId()
	return context.WithValue(ctx, spanKey{}, id), &remoteSpan{client: t.client, id: id}
}

type spanKey struct{}

func parentSpanFrom(ctx context.Context) string {
	if id, ok := ctx.Value(spanKey{}).(string); ok {
		return id
	}
	return ""
}

type remoteSpan struct {
	client modulev1.TelemetryServiceClient
	id     string
	once   sync.Once
}

var _ v1.Span = (*remoteSpan)(nil)

func (s *remoteSpan) SetAttributes(attrs ...v1.Field) {
	//nolint:errcheck // deliberate: telemetry never fails the operation.
	_, _ = s.client.SetSpanAttributes(context.Background(), &modulev1.SpanAttributesRequest{
		SpanId:     s.id,
		Attributes: fieldsToWire(attrs),
	})
}

func (s *remoteSpan) Fail(err error) {
	if err == nil {
		// The SDK documents span.Fail(err) as safe on the success path, so a
		// nil error is ignored rather than sent.
		return
	}
	//nolint:errcheck // deliberate.
	_, _ = s.client.FailSpan(context.Background(), &modulev1.FailSpanRequest{
		SpanId:  s.id,
		Message: err.Error(),
	})
}

// End is idempotent, as the SDK requires. sync.Once is what makes that true
// across a boundary where a second End would otherwise be a second RPC against
// a span the Platform has already dropped.
func (s *remoteSpan) End() {
	s.once.Do(func() {
		//nolint:errcheck // deliberate.
		_, _ = s.client.EndSpan(context.Background(), &modulev1.EndSpanRequest{SpanId: s.id})
	})
}

type noopSpan struct{}

func (noopSpan) SetAttributes(...v1.Field) {}
func (noopSpan) Fail(error)                {}
func (noopSpan) End()                      {}

// ─── Platform side ──────────────────────────────────────────────────────────

// telemetryServer runs in the Platform and turns the module's calls into calls
// on the Platform's real Telemetry.
//
// It holds the live spans, because a span is stateful and the wire is not: the
// module refers to a span by id across several calls, and something has to hold
// the [v1.Span] those ids mean. Spans are removed on End, so an abandoned span
// leaks an entry — which matches the SDK's own contract that an unended span is
// dropped, and is bounded by the module process's lifetime.
type telemetryServer struct {
	modulev1.UnimplementedTelemetryServiceServer

	impl       v1.Telemetry
	categoryOf CategoryFunc

	mu    sync.Mutex
	spans map[string]v1.Span
	next  uint64
}

func (s *telemetryServer) Log(_ context.Context, req *modulev1.LogRequest) (*modulev1.LogResponse, error) {
	if s.impl == nil {
		return &modulev1.LogResponse{}, nil
	}
	fields := fieldsFromWire(req.GetFields())
	switch req.GetLevel() {
	case modulev1.LogLevel_LOG_LEVEL_DEBUG:
		s.impl.Debug(req.GetMessage(), fields...)
	case modulev1.LogLevel_LOG_LEVEL_WARN:
		s.impl.Warn(req.GetMessage(), fields...)
	case modulev1.LogLevel_LOG_LEVEL_ERROR:
		s.impl.Error(req.GetMessage(), fields...)
	default:
		s.impl.Info(req.GetMessage(), fields...)
	}
	return &modulev1.LogResponse{}, nil
}

func (s *telemetryServer) StartSpan(ctx context.Context, req *modulev1.StartSpanRequest) (*modulev1.StartSpanResponse, error) {
	if s.impl == nil {
		return &modulev1.StartSpanResponse{}, nil
	}
	_, span := s.impl.Span(ctx, req.GetName(), fieldsFromWire(req.GetAttributes())...)

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.spans == nil {
		s.spans = map[string]v1.Span{}
	}
	s.next++
	id := fmt.Sprintf("s%d", s.next)
	s.spans[id] = span
	return &modulev1.StartSpanResponse{SpanId: id}, nil
}

func (s *telemetryServer) span(id string) v1.Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.spans[id]
}

func (s *telemetryServer) SetSpanAttributes(_ context.Context, req *modulev1.SpanAttributesRequest) (*modulev1.SpanAttributesResponse, error) {
	if sp := s.span(req.GetSpanId()); sp != nil {
		sp.SetAttributes(fieldsFromWire(req.GetAttributes())...)
	}
	return &modulev1.SpanAttributesResponse{}, nil
}

func (s *telemetryServer) FailSpan(_ context.Context, req *modulev1.FailSpanRequest) (*modulev1.FailSpanResponse, error) {
	if sp := s.span(req.GetSpanId()); sp != nil {
		sp.Fail(&wireError{category: req.GetCategory(), message: req.GetMessage()})
	}
	return &modulev1.FailSpanResponse{}, nil
}

func (s *telemetryServer) EndSpan(_ context.Context, req *modulev1.EndSpanRequest) (*modulev1.EndSpanResponse, error) {
	s.mu.Lock()
	sp := s.spans[req.GetSpanId()]
	delete(s.spans, req.GetSpanId())
	s.mu.Unlock()
	if sp != nil {
		sp.End()
	}
	return &modulev1.EndSpanResponse{}, nil
}
