package term

import "os"

// This example uses a Printer with Color forced off, matching what New
// auto-detects when UI is not an interactive terminal (e.g. Stdout/Stderr
// piped to a file, as happens when running `go test`).
func Example() {
	color := false
	p := New(Options{
		Data:  os.Stdout,
		UI:    os.Stdout,
		Color: &color,
	})

	p.Info("starting sync")
	p.Success("sync complete")
	p.Warn("2 records skipped")
	p.Table("id", "status")
	p.Table(1, "ok")

	// Output:
	// starting sync
	// sync complete
	// 2 records skipped
	// id, status
	// 1, ok
}
