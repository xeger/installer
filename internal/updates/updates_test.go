package updates

import (
	"testing"
	"time"
)

func TestNewer(t *testing.T) {
	tests := []struct {
		current, latest string
		want            bool
	}{
		{"v1.9.0", "v1.10.0", true}, // the old string comparison got this wrong
		{"v1.10.0", "v1.9.0", false},
		{"v1.2.3", "v1.2.3", false},
		{"v1.2.3-rc.1", "v1.2.3", true},
		{"", "v1.0.0", true},
		{"nightly-1", "nightly-2", true},
		{"nightly-1", "nightly-1", false},
		{"v1.0.0", "", false},
	}
	for _, tt := range tests {
		if got := Newer(tt.current, tt.latest); got != tt.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestDue(t *testing.T) {
	now := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC)
	if !Due(time.Time{}, now) {
		t.Error("never checked should be due")
	}
	if Due(now.Add(-Interval+time.Minute), now) {
		t.Error("checked within the interval should not be due")
	}
	if !Due(now.Add(-Interval), now) {
		t.Error("checked exactly one interval ago should be due")
	}
}

func TestIsRelease(t *testing.T) {
	for v, want := range map[string]bool{
		"v1.2.3":                             true,
		"dev":                                false,
		"v1.2.3-rc.1":                        false,
		"v0.0.0-20260927200005-e48ea5a74d4f": false,
		"v1.2.3+dirty":                       false,
	} {
		if got := IsRelease(v); got != want {
			t.Errorf("IsRelease(%q) = %v, want %v", v, got, want)
		}
	}
}
