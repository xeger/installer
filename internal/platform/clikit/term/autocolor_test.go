package term

import (
	"bytes"
	"os"
	"testing"
)

func TestAutoColorNotATerminal(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	if got := autoColor(w); got {
		t.Error("autoColor(pipe) = true, want false (not a TTY)")
	}
}

func TestAutoColorNonFileWriter(t *testing.T) {
	var buf bytes.Buffer
	if got := autoColor(&buf); got {
		t.Error("autoColor(*bytes.Buffer) = true, want false")
	}
}

func TestAutoColorCliColorForce(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	t.Run("forces on for a non-tty *os.File", func(t *testing.T) {
		t.Setenv("CLICOLOR_FORCE", "1")
		if got := autoColor(w); !got {
			t.Error("autoColor() = false with CLICOLOR_FORCE=1, want true")
		}
	})

	t.Run("forces on for a non-file writer too", func(t *testing.T) {
		t.Setenv("CLICOLOR_FORCE", "1")
		var buf bytes.Buffer
		if got := autoColor(&buf); !got {
			t.Error("autoColor() = false with CLICOLOR_FORCE=1, want true")
		}
	})

	t.Run("CLICOLOR_FORCE=0 does not force", func(t *testing.T) {
		t.Setenv("CLICOLOR_FORCE", "0")
		if got := autoColor(w); got {
			t.Error("autoColor() = true with CLICOLOR_FORCE=0, want false")
		}
	})

	t.Run("empty CLICOLOR_FORCE does not force", func(t *testing.T) {
		t.Setenv("CLICOLOR_FORCE", "")
		if got := autoColor(w); got {
			t.Error("autoColor() = true with CLICOLOR_FORCE=\"\", want false")
		}
	})
}

// NO_COLOR takes priority over CLICOLOR_FORCE: it always wins, per
// no-color.org, even when something is explicitly trying to force color
// on.
func TestAutoColorNoColorWinsOverForce(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	t.Setenv("CLICOLOR_FORCE", "1")
	t.Setenv("NO_COLOR", "1")

	if got := autoColor(w); got {
		t.Error("autoColor() = true with both NO_COLOR and CLICOLOR_FORCE set, want false (NO_COLOR wins)")
	}
}

func TestAutoColorDumbTerm(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	t.Setenv("TERM", "dumb")
	if got := autoColor(w); got {
		t.Error("autoColor() = true with TERM=dumb, want false")
	}
}
