package host

import (
	"os"
	"strings"
	"testing"
)

// TestParentSDKHasNoDependencies is the executable form of the reason this
// nested module exists.
//
// `sdk`'s go.mod is a module line and a Go version, and that is load-bearing:
// a third party compiles against the contract and against nothing the Platform
// happened to choose. ADR 0059 refused to re-export OpenTelemetry to keep it
// that way, and ADR 0064 split the harness out here rather than let gRPC in.
//
// The property decays silently — a `go get` in the wrong directory is all it
// takes, and nothing else fails. So it is asserted from the one module that
// would notice, and whose own dependency list is the temptation.
func TestParentSDKHasNoDependencies(t *testing.T) {
	data, err := os.ReadFile("../go.mod")
	if err != nil {
		t.Fatalf("reading the parent go.mod: %v", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "", strings.HasPrefix(trimmed, "//"):
			continue
		case strings.HasPrefix(trimmed, "module "), strings.HasPrefix(trimmed, "go "):
			continue
		case strings.HasPrefix(trimmed, "toolchain "):
			// A toolchain directive is a Go version statement, not a dependency.
			continue
		default:
			t.Errorf("sdk/go.mod must stay dependency-free, found: %q\n"+
				"If the SDK genuinely needs this, it belongs in sdk/host instead — "+
				"see ADR 0059 and ADR 0064.", trimmed)
		}
	}
}
