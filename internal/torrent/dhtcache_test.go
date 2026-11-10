package torrent

import (
	"fmt"
	"slices"
	"testing"
)

// A lookup's closest nodes are remembered for the next torrent, once
// each, up to the bound.
func TestTheEngineRemembersDHTNodes(t *testing.T) {
	t.Parallel()
	var e Engine
	if got := e.knownDHT(); len(got) != 0 {
		t.Fatalf("a new engine knows %q", got)
	}
	e.rememberDHT([]string{"a:1", "b:1", "a:1", ""})
	if got, want := e.knownDHT(), []string{"a:1", "b:1"}; !slices.Equal(got, want) {
		t.Fatalf("known DHT nodes: got %q, want %q", got, want)
	}
	for i := 0; i < maxKnownDHT+8; i++ {
		e.rememberDHT([]string{fmt.Sprintf("node-%d:1", i)})
	}
	if got := len(e.knownDHT()); got != maxKnownDHT {
		t.Fatalf("the engine keeps %d DHT nodes, want %d", got, maxKnownDHT)
	}
	var nilEngine *Engine
	if got := nilEngine.knownDHT(); got != nil {
		t.Fatalf("a nil engine knows %q", got)
	}
}
