package diff

// LineStatus describes how a RichLine relates to the two sides of the diff.
type LineStatus int

const (
	// Shared is a line present and identical on both sides (only emitted when
	// the caller chooses not to collapse the shared prefix).
	Shared LineStatus = iota
	// LeftOnly is a line present only on the left side (right ran out).
	LeftOnly
	// RightOnly is a line present only on the right side (left ran out).
	RightOnly
	// Changed is a line present on both sides but with differing content.
	Changed
	// Collapsed is a synthetic single row standing in for N shared rows.
	Collapsed
)

// RichLine is one row of a RichDiff. Text columns are caller-provided strings;
// Meta is an opaque value owned by the caller, surfaced via MetaRenderer.
type RichLine struct {
	LeftText  string
	RightText string
	Status    LineStatus
	Meta      any
}

// RichDiff is an ordered sequence of RichLines plus presentation hints.
// ForkIndex is the index into Lines at which a fork-point marker should be
// rendered (between Lines[ForkIndex-1] and Lines[ForkIndex]). If ForkIndex
// equals 0 or len(Lines), no marker is rendered.
type RichDiff struct {
	Lines      []RichLine
	ForkIndex  int
	LeftLabel  string
	RightLabel string
}
