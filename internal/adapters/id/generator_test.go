package id

import (
	"errors"
	"io"
	"regexp"
	"testing"
	"time"
)

func TestGeneratorNewIDUsesStableUUIDFormat(t *testing.T) {
	t.Parallel()

	generator := NewGenerator()
	got := generator.NewID()

	if !shipmentIDPattern.MatchString(got) {
		t.Fatalf("NewID() = %q, want shipment-prefixed UUID format", got)
	}
}

func TestGeneratorFallbackKeepsSameUUIDFormat(t *testing.T) {
	t.Parallel()

	var counter uint64
	generator := &Generator{
		random:  failingReader{},
		now:     func() time.Time { return time.Date(2026, 3, 19, 9, 0, 0, 0, time.UTC) },
		pid:     42,
		counter: &counter,
	}

	got := generator.NewID()
	if !shipmentIDPattern.MatchString(got) {
		t.Fatalf("fallback NewID() = %q, want shipment-prefixed UUID format", got)
	}
}

var shipmentIDPattern = regexp.MustCompile(`^shipment-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type failingReader struct{}

func (failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
}

var _ io.Reader = failingReader{}
