package term

import (
	"fmt"
	"strings"
)

// unwrapMulti matches errors that unwrap to more than one error, such as
// those produced by errors.Join.
type unwrapMulti interface {
	Unwrap() []error
}

// ErrorDetail prints an Error message built from prefaceContext, then an
// indented tree describing err on the UI stream. It descends errors
// implementing interface{ Unwrap() []error } (as errors.Join does),
// recursing into each child; a single-error Unwrap is followed
// transparently, and an Unwrap() []error that returns no errors is treated
// as a leaf. Multi-line error messages are re-indented to match their tree
// depth.
//
// When a multi-child error's own Error() text is more than just its
// children's messages joined by "\n" (the format errors.Join produces) —
// for example a fmt.Errorf("syncing site %s: %w; %w", site, a, b), which
// adds its own context around the two wrapped errors — that message is
// printed as its own line above the children, instead of being discarded
// in favor of a generic "multiple errors:" label.
//
// err may be nil, in which case only the preface is printed (or "<nil>" if
// prefaceContext is also empty). When prefaceContext is empty, no preface
// line is printed at all — just the tree.
//
// The full output (preface header and tree) is built up front and written
// to the UI stream in a single Write call, so a String/Error method
// invoked while building the tree can safely log through this same Printer
// without deadlocking on it.
func (p *Printer) ErrorDetail(err error, prefaceContext ...any) {
	var b strings.Builder

	switch {
	case err == nil && len(prefaceContext) == 0:
		b.WriteString("<nil>\n")
	case err == nil:
		b.WriteString(p.formatLevel(errorLevel, prefaceContext...))
	default:
		if len(prefaceContext) > 0 {
			b.WriteString(p.formatLevel(errorLevel, prefaceContext...))
		}
		explainError(&b, 1, err)
	}

	out := b.String()
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprint(p.ui, out)
}

// explainError writes an indented tree describing err into b. It performs
// no I/O and takes no lock, so it's safe to call while building output
// that will later be written under a single lock (see ErrorDetail).
func explainError(b *strings.Builder, nesting int, err error) {
	indent := strings.Repeat(" ", nesting*2)

	wm, ok := err.(unwrapMulti)
	if !ok {
		writeLeaf(b, indent, err)
		return
	}

	nested := wm.Unwrap()
	switch len(nested) {
	case 0:
		// Unwrap() []error with no children: treat as a leaf.
		writeLeaf(b, indent, err)
		return
	case 1:
		explainError(b, nesting, nested[0])
		return
	}

	own := err.Error()
	if own != joinChildMessages(nested) {
		// err's own message adds context beyond just its children's
		// messages; keep it, then descend into the children beneath it.
		writeLeaf(b, indent, err)
		for _, e := range nested {
			explainError(b, nesting+1, e)
		}
		return
	}

	fmt.Fprintf(b, "%smultiple errors:\n", indent)
	for _, e := range nested {
		explainError(b, nesting+1, e)
	}
}

// writeLeaf writes err's message, indented, as a tree leaf.
func writeLeaf(b *strings.Builder, indent string, err error) {
	text := err.Error()
	text = strings.ReplaceAll(text, "\n", "\n"+indent)
	fmt.Fprintf(b, "%s%s\n", indent, text)
}

// joinChildMessages reproduces the string errors.Join(nested...).Error()
// would produce: each child's Error() text, joined with "\n".
func joinChildMessages(nested []error) string {
	msgs := make([]string, len(nested))
	for i, e := range nested {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "\n")
}
