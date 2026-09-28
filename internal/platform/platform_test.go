package platform

import (
	"path/filepath"
	"testing"
)

func TestDirs(t *testing.T) {
	data, state := t.TempDir(), t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	t.Setenv("XDG_STATE_HOME", state)

	got, err := DataDir("tools", "foo")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(data, "xn", "tools", "foo"); got != want {
		t.Errorf("DataDir() = %q, want %q", got, want)
	}
	got, err = StateDir("self-update.json")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(state, "xn", "self-update.json"); got != want {
		t.Errorf("StateDir() = %q, want %q", got, want)
	}
}

func TestCandidates(t *testing.T) {
	if got := Candidates("semdiff"); len(got) != 1 || got[0] != "xeger/semdiff" {
		t.Errorf("Candidates(semdiff) = %v, want [xeger/semdiff]", got)
	}
	if got := Candidates("foo"); got != nil {
		t.Errorf("Candidates(foo) = %v, want nil", got)
	}
}
