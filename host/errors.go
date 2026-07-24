package host

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	modulev1 "github.com/mosaic-media/contracts/gen/mosaic/module/v1"
)

// Error handling across the boundary, and the layering constraint that shapes
// it.
//
// The Platform's error categories live in `internal/platform/contracts` and are
// deliberately not published in the SDK (ADR 0016: a capability calls
// application services, and the categories are the Platform's own vocabulary).
// This package compiles against the SDK, so **it cannot read a category off an
// error** — it only ever sees `error`.
//
// That is why [CategoryFunc] exists. The Platform, which does know its own
// vocabulary, injects a function that names the category of an error it
// produced. Nothing here knows what the strings mean; it passes them through so
// the far end and the telemetry plane keep them.
//
// Two consequences worth stating rather than discovering:
//
//   - **A module still cannot classify an error, and this boundary does not
//     change that.** In process it receives an error whose category it has no
//     type to read; here it receives the same. The category on the wire is for
//     the Platform end and for telemetry. If a module should ever be able to
//     branch on a category, that is an SDK addition, and the value is already
//     being carried for it.
//   - **An error a module returns has no category, exactly as in process.**
//     The Platform's CategoryOf maps an uncategorised error to Internal, and a
//     module has no way to construct a categorised one, so nothing is lost by
//     sending it as a plain message.

// CategoryFunc reports the Platform error category of err, as the string the
// Platform uses for it ("not_found", "conflict", …), or "" when err carries
// none. The Platform supplies one; a module never does.
type CategoryFunc func(err error) string

// categoryToWire maps the Platform's own category strings onto the wire enum.
// It is a switch on strings rather than a shared constant set because the
// constants are Platform-internal and this package cannot import them — the
// coupling is real and is better visible here than hidden behind a cast.
func categoryToWire(s string) modulev1.ErrorCategory {
	switch s {
	case "invalid_argument":
		return modulev1.ErrorCategory_ERROR_CATEGORY_INVALID_ARGUMENT
	case "unauthenticated":
		return modulev1.ErrorCategory_ERROR_CATEGORY_UNAUTHENTICATED
	case "permission_denied":
		return modulev1.ErrorCategory_ERROR_CATEGORY_PERMISSION_DENIED
	case "not_found":
		return modulev1.ErrorCategory_ERROR_CATEGORY_NOT_FOUND
	case "conflict":
		return modulev1.ErrorCategory_ERROR_CATEGORY_CONFLICT
	case "unavailable":
		return modulev1.ErrorCategory_ERROR_CATEGORY_UNAVAILABLE
	case "internal":
		return modulev1.ErrorCategory_ERROR_CATEGORY_INTERNAL
	default:
		return modulev1.ErrorCategory_ERROR_CATEGORY_UNSPECIFIED
	}
}

// categoryToCode gives the gRPC status a sensible code. The code is a transport
// hint and not the contract: the category rides as a status *detail*, because
// the two vocabularies are not in one-to-one correspondence and round-tripping
// through codes alone would turn a Conflict into whatever code happened to be
// closest.
func categoryToCode(c modulev1.ErrorCategory) codes.Code {
	switch c {
	case modulev1.ErrorCategory_ERROR_CATEGORY_INVALID_ARGUMENT:
		return codes.InvalidArgument
	case modulev1.ErrorCategory_ERROR_CATEGORY_UNAUTHENTICATED:
		return codes.Unauthenticated
	case modulev1.ErrorCategory_ERROR_CATEGORY_PERMISSION_DENIED:
		return codes.PermissionDenied
	case modulev1.ErrorCategory_ERROR_CATEGORY_NOT_FOUND:
		return codes.NotFound
	case modulev1.ErrorCategory_ERROR_CATEGORY_CONFLICT:
		return codes.FailedPrecondition
	case modulev1.ErrorCategory_ERROR_CATEGORY_UNAVAILABLE:
		return codes.Unavailable
	default:
		return codes.Unknown
	}
}

// errorToWire turns an error into a gRPC status carrying the category as a
// detail. categoryOf may be nil, which is the module side: it has no categories
// to report and sends the message alone.
func errorToWire(err error, categoryOf CategoryFunc) error {
	if err == nil {
		return nil
	}
	category := modulev1.ErrorCategory_ERROR_CATEGORY_UNSPECIFIED
	if categoryOf != nil {
		category = categoryToWire(categoryOf(err))
	}

	st := status.New(categoryToCode(category), err.Error())
	withDetail, detailErr := st.WithDetails(&modulev1.Error{
		Category: category,
		Message:  err.Error(),
	})
	if detailErr != nil {
		// Attaching a detail can fail only on a marshalling problem. Losing the
		// category is much better than losing the error, so fall back to the
		// bare status rather than reporting the marshalling failure in its
		// place.
		return st.Err()
	}
	return withDetail.Err()
}

// errorFromWire reconstructs an error from a gRPC status. The message is
// preserved exactly; the category is read from the detail when present and is
// currently observable only through [CategoryOfWireError], since the SDK
// publishes no category type for a caller to branch on.
func errorFromWire(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	for _, d := range st.Details() {
		if e, ok := d.(*modulev1.Error); ok {
			return &wireError{category: e.GetCategory(), message: e.GetMessage()}
		}
	}
	return errors.New(st.Message())
}

// wireError carries a category across the boundary without publishing a
// category type in the SDK. It is an ordinary error to anyone who does not know
// to look.
type wireError struct {
	category modulev1.ErrorCategory
	message  string
}

func (e *wireError) Error() string { return e.message }

// CategoryOfWireError reports the Platform category string carried by an error
// that crossed the boundary, or "" if it carries none. It exists for the
// Platform end and for telemetry; a module has no reason to call it and no
// vocabulary to compare the result against.
func CategoryOfWireError(err error) string {
	var we *wireError
	if !errors.As(err, &we) {
		return ""
	}
	switch we.category {
	case modulev1.ErrorCategory_ERROR_CATEGORY_INVALID_ARGUMENT:
		return "invalid_argument"
	case modulev1.ErrorCategory_ERROR_CATEGORY_UNAUTHENTICATED:
		return "unauthenticated"
	case modulev1.ErrorCategory_ERROR_CATEGORY_PERMISSION_DENIED:
		return "permission_denied"
	case modulev1.ErrorCategory_ERROR_CATEGORY_NOT_FOUND:
		return "not_found"
	case modulev1.ErrorCategory_ERROR_CATEGORY_CONFLICT:
		return "conflict"
	case modulev1.ErrorCategory_ERROR_CATEGORY_UNAVAILABLE:
		return "unavailable"
	case modulev1.ErrorCategory_ERROR_CATEGORY_INTERNAL:
		return "internal"
	default:
		return ""
	}
}
