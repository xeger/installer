package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/xeger/installer/internal/github"
	"github.com/xeger/installer/internal/platform"
	"github.com/xeger/installer/internal/ui"
	"github.com/xeger/installer/internal/updates"
)

type selfState struct {
	CheckedAt time.Time `json:"checked_at"`
}

// NotifySelfUpdate tells the user when a newer launcher release is available
// from platform.SelfRepo, checking at most once per updates.Interval. It is
// best-effort and silent on failure, and skipped for dev builds.
func NotifySelfUpdate(ctx context.Context) {
	current := platform.Version()
	if !updates.IsRelease(current) {
		return
	}
	path, err := platform.StateDir("self-update.json")
	if err != nil {
		return
	}
	var state selfState
	if data, err := os.ReadFile(path); err == nil { //nolint:gosec // the launcher's own state file
		_ = json.Unmarshal(data, &state)
	}
	if !updates.Due(state.CheckedAt, now()) {
		return
	}

	rel, err := github.Latest(ctx, platform.SelfRepo)
	if err != nil {
		return
	}
	state.CheckedAt = now()
	if data, err := json.Marshal(state); err == nil {
		if os.MkdirAll(filepath.Dir(path), 0o755) == nil { //nolint:gosec // not secret
			_ = os.WriteFile(path, data, 0o644) //nolint:gosec // not secret
		}
	}
	if updates.Newer(current, rel.Tag) {
		ui.Default.Info(platform.Name, rel.Tag, "is available (you have "+current+"): https://github.com/"+platform.SelfRepo+"/releases/latest")
	}
}
