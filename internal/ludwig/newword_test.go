// Tests for newword.go

package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// buildWordFrame creates a frame containing the given line contents.
// Non-empty strings are written as line content; empty strings become
// blank lines (Used=0). The frame's Dot is placed on lines[0] at col 1.
// The returned lines slice gives direct access to each LineHdrObject.
func buildWordFrame(contents []string) (*FrameObject, []*LineHdrObject) {
	n := len(contents)
	frame, lines := setupTestFrame(n)

	// Adjust the frame to match production invariants: in the real editor,
	// the null/EOP line is part of the group and is the group's LastLine.
	// setupTestFrame leaves the EOP line as a separate FLink-only node
	// after the last real line, so we reattach it here.
	if frame != nil && frame.LastGroup != nil && frame.LastGroup.LastLine != nil {
		lastReal := frame.LastGroup.LastLine
		if lastReal.FLink != nil {
			eop := lastReal.FLink
			frame.LastGroup.LastLine = eop
			// Ensure the EOP line is associated with the group, if applicable.
			eop.Group = frame.LastGroup
		}
	}
	for i, s := range contents {
		if s != "" {
			setLineContent(lines[i], s)
		}
	}
	frame.Dot.Line = lines[0]
	frame.Dot.Col = 1
	return frame, lines
}

// newDotMark creates and registers a MarkObject at the given line/col.
func newDotMark(line *LineHdrObject, col int) *MarkObject {
	var m *MarkObject
	MarkCreate(line, col, &m)
	return m
}

// ---- currentWord -------------------------------------------------------

func TestCurrentWord(t *testing.T) {
	// "hello world foo"
	// h=1 e=2 l=3 l=4 o=5 _=6 w=7 o=8 r=9 l=10 d=11 _=12 f=13 o=14 o=15
	const line = "hello world foo"

	t.Run("DotInMiddleOfWord", func(t *testing.T) {
		_, lines := buildWordFrame([]string{line})
		dot := newDotMark(lines[0], 8) // 'o' in "world"
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.True(t, ok)
		assert.Equal(t, 7, dot.Col) // 'w' = start of "world"
	})

	t.Run("DotAtWordStart", func(t *testing.T) {
		_, lines := buildWordFrame([]string{line})
		dot := newDotMark(lines[0], 7) // 'w' start of "world"
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.True(t, ok)
		assert.Equal(t, 7, dot.Col) // unchanged
	})

	t.Run("DotAtFirstWordStart", func(t *testing.T) {
		_, lines := buildWordFrame([]string{line})
		dot := newDotMark(lines[0], 1)
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.True(t, ok)
		assert.Equal(t, 1, dot.Col)
	})

	t.Run("DotInSpaceBetweenWords", func(t *testing.T) {
		// col 6 is the space between "hello" and "world"
		// currentWord moves backward past the space and finds start of "hello"
		_, lines := buildWordFrame([]string{line})
		dot := newDotMark(lines[0], 6)
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.True(t, ok)
		assert.Equal(t, 1, dot.Col) // start of "hello"
	})

	t.Run("DotSlightlyPastEnd", func(t *testing.T) {
		// col = Used+1 → clamped to Used then word-start found
		_, lines := buildWordFrame([]string{"hello"})
		dot := newDotMark(lines[0], 6) // Used=5, col=6=Used+1
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.True(t, ok)
		assert.Equal(t, 1, dot.Col) // start of "hello"
	})

	t.Run("DotFarPastEndBlankFLink", func(t *testing.T) {
		// col > Used+1 and FLink (null line) has Used=0 → false
		_, lines := buildWordFrame([]string{"hello"})
		dot := newDotMark(lines[0], 8) // Used=5, col=8 > Used+2=7
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.False(t, ok)
	})

	t.Run("BlankLine", func(t *testing.T) {
		// Used=0 → dot.Col clamped to 0 → false
		_, lines := buildWordFrame([]string{""})
		dot := newDotMark(lines[0], 1)
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.False(t, ok)
	})

	t.Run("LeadingSpacesBLinkNil", func(t *testing.T) {
		// dot at col 1 which is a space, BLink is nil → false
		_, lines := buildWordFrame([]string{"  abc"})
		lines[0].BLink = nil // ensure no predecessor
		dot := newDotMark(lines[0], 1) // col 1 = ' '
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.False(t, ok)
	})

	t.Run("LeadingSpacesBLinkBlank", func(t *testing.T) {
		// dot at col 1 (space), BLink exists but is blank → false
		_, lines := buildWordFrame([]string{"", "  abc"})
		// lines[1] has Used=5 and starts with spaces
		dot := newDotMark(lines[1], 1) // col 1 = ' '
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.False(t, ok)
	})

	t.Run("CrossesToPreviousLineWord", func(t *testing.T) {
		// Line 1 is all spaces; BLink (line 0) has content.
		// currentWord crosses to line 0 and finds start of its last word.
		// "foo bar": f=1 o=2 o=3 _=4 b=5 a=6 r=7
		_, lines := buildWordFrame([]string{"foo bar", "  "})
		dot := newDotMark(lines[1], 1) // col 1 = ' ' on line 1
		defer MarkDestroy(&dot)

		ok := currentWord(dot)

		assert.True(t, ok)
		assert.Equal(t, lines[0], dot.Line)
		assert.Equal(t, 5, dot.Col) // 'b' = start of "bar"
	})
}

// ---- nextWord ----------------------------------------------------------

func TestNextWord(t *testing.T) {
	// "hello world": h=1 e=2 l=3 l=4 o=5 _=6 w=7 o=8 r=9 l=10 d=11

	t.Run("AdvancesFromStartOfWord", func(t *testing.T) {
		_, lines := buildWordFrame([]string{"hello world"})
		dot := newDotMark(lines[0], 1) // 'h'
		defer MarkDestroy(&dot)

		ok := nextWord(dot)

		assert.True(t, ok)
		assert.Equal(t, 7, dot.Col) // start of "world"
	})

	t.Run("AdvancesFromMiddleOfWord", func(t *testing.T) {
		_, lines := buildWordFrame([]string{"hello world"})
		dot := newDotMark(lines[0], 3) // 'l' in "hello"
		defer MarkDestroy(&dot)

		ok := nextWord(dot)

		assert.True(t, ok)
		assert.Equal(t, 7, dot.Col) // start of "world"
	})

	t.Run("AdvancesToWordOnNextLine", func(t *testing.T) {
		_, lines := buildWordFrame([]string{"hello", "world"})
		dot := newDotMark(lines[0], 3) // 'l' in "hello"
		defer MarkDestroy(&dot)

		ok := nextWord(dot)

		assert.True(t, ok)
		assert.Equal(t, lines[1], dot.Line)
		assert.Equal(t, 1, dot.Col) // start of "world"
	})

	t.Run("FailsOnBlankLine", func(t *testing.T) {
		_, lines := buildWordFrame([]string{""})
		dot := newDotMark(lines[0], 1)
		defer MarkDestroy(&dot)

		ok := nextWord(dot)

		assert.False(t, ok)
	})

	t.Run("FailsAtEndOfLastWordBlankFLink", func(t *testing.T) {
		// Single word; FLink is null (Used=0) → false
		_, lines := buildWordFrame([]string{"hello"})
		dot := newDotMark(lines[0], 3) // 'l'
		defer MarkDestroy(&dot)

		ok := nextWord(dot)

		assert.False(t, ok)
	})

	t.Run("FailsAtEndOfParagraph", func(t *testing.T) {
		// Last word before blank line (paragraph end) → false
		_, lines := buildWordFrame([]string{"hello world", ""})
		dot := newDotMark(lines[0], 7) // 'w' start of "world"
		defer MarkDestroy(&dot)

		ok := nextWord(dot)

		assert.False(t, ok) // FLink.Used == 0
	})

	t.Run("ColPastEndClampsToUsed", func(t *testing.T) {
		// dot.Col > Used → clamped to Used, then advance; FLink blank → false
		_, lines := buildWordFrame([]string{"hello world"})
		dot := newDotMark(lines[0], 20) // past end
		defer MarkDestroy(&dot)

		ok := nextWord(dot)

		assert.False(t, ok)
	})
}

// ---- previousWord ------------------------------------------------------

func TestPreviousWord(t *testing.T) {
	// "hello world foo": h=1…o=5 _=6 w=7…d=11 _=12 f=13 o=14 o=15

	t.Run("GoesToPreviousWordSameLine", func(t *testing.T) {
		_, lines := buildWordFrame([]string{"hello world foo"})
		dot := newDotMark(lines[0], 9) // 'r' in "world"
		defer MarkDestroy(&dot)

		ok := previousWord(dot)

		assert.True(t, ok)
		assert.Equal(t, 1, dot.Col) // start of "hello"
	})

	t.Run("GoesToPreviousWordFromWordStart", func(t *testing.T) {
		// At start of "world"; previousWord → start of "hello"
		_, lines := buildWordFrame([]string{"hello world"})
		dot := newDotMark(lines[0], 7) // 'w'
		defer MarkDestroy(&dot)

		ok := previousWord(dot)

		assert.True(t, ok)
		assert.Equal(t, 1, dot.Col) // start of "hello"
	})

	t.Run("FailsAtFirstWordNoBLink", func(t *testing.T) {
		// currentWord finds start of "hello", BLink is nil → false
		_, lines := buildWordFrame([]string{"hello world"})
		dot := newDotMark(lines[0], 3) // 'l' in "hello"
		defer MarkDestroy(&dot)

		ok := previousWord(dot)

		assert.False(t, ok)
	})

	t.Run("FailsAtFirstWordBlankBLink", func(t *testing.T) {
		// Paragraph boundary: BLink.Used == 0 → false
		_, lines := buildWordFrame([]string{"", "hello world"})
		dot := newDotMark(lines[1], 3)
		defer MarkDestroy(&dot)

		ok := previousWord(dot)

		assert.False(t, ok)
	})

	t.Run("CrossesToPreviousLine", func(t *testing.T) {
		// dot at start of "hello" on line 1; crosses to last word of line 0
		// "foo bar": f=1 o=2 o=3 _=4 b=5 a=6 r=7
		_, lines := buildWordFrame([]string{"foo bar", "hello"})
		dot := newDotMark(lines[1], 1) // 'h' at col 1
		defer MarkDestroy(&dot)

		ok := previousWord(dot)

		assert.True(t, ok)
		assert.Equal(t, lines[0], dot.Line)
		assert.Equal(t, 5, dot.Col) // 'b' = start of "bar"
	})
}

// ---- NewwordAdvanceWord ------------------------------------------------

func TestNewwordAdvanceWord(t *testing.T) {
	// "hello world foo": h=1…o=5 _=6 w=7…d=11 _=12 f=13 o=14 o=15

	t.Run("ForwardOneWord", func(t *testing.T) {
		frame, lines := buildWordFrame([]string{"hello world foo"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordAdvanceWord(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, 7, frame.Dot.Col) // start of "world"
	})

	t.Run("ForwardTwoWords", func(t *testing.T) {
		frame, lines := buildWordFrame([]string{"hello world foo"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordAdvanceWord(frame, LeadParamPInt, 2)

		assert.True(t, ok)
		assert.Equal(t, 13, frame.Dot.Col) // start of "foo"
	})

	t.Run("ZeroCountGoesToCurrentWord", func(t *testing.T) {
		// LeadParamPInt count=0 is treated as NInt: go to start of current word
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 9 // 'r' in "world"

		ok := NewwordAdvanceWord(frame, LeadParamPInt, 0)

		assert.True(t, ok)
		assert.Equal(t, 7, frame.Dot.Col) // start of "world"
	})

	t.Run("BackwardOneWordNInt", func(t *testing.T) {
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 9 // 'r' in "world"

		ok := NewwordAdvanceWord(frame, LeadParamNInt, -1)

		assert.True(t, ok)
		assert.Equal(t, 1, frame.Dot.Col) // start of "hello"
	})

	t.Run("BackwardOneWordMinus", func(t *testing.T) {
		frame, lines := buildWordFrame([]string{"hello world foo"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 13 // start of "foo"

		ok := NewwordAdvanceWord(frame, LeadParamMinus, -1)

		assert.True(t, ok)
		assert.Equal(t, 7, frame.Dot.Col) // start of "world"
	})

	t.Run("ForwardToEndOfParagraph", func(t *testing.T) {
		// LeadParamPIndef advances through all words then positions at Used+2
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordAdvanceWord(frame, LeadParamPIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, lines[0], frame.Dot.Line)
		assert.Equal(t, lines[0].Used+2, frame.Dot.Col) // 11+2=13
	})

	t.Run("ParagraphEndFailsOnBlankLine", func(t *testing.T) {
		frame, lines := buildWordFrame([]string{""})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordAdvanceWord(frame, LeadParamPIndef, 0)

		assert.False(t, ok)
	})

	t.Run("BackwardToStartOfParagraph", func(t *testing.T) {
		// LeadParamNIndef advances backward to first word in paragraph
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 11 // 'd' end of "world"

		ok := NewwordAdvanceWord(frame, LeadParamNIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, 1, frame.Dot.Col) // start of "hello"
	})

	t.Run("MarkerUsesMarkPosition", func(t *testing.T) {
		// LeadParamMarker: starting position taken from frame.Marks[1],
		// then acts as NInt count=0 → currentWord from that position
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1
		MarkCreate(lines[0], 9, &frame.Marks[1]) // mark 1 at 'r' in "world"
		defer MarkDestroy(&frame.Marks[1])

		ok := NewwordAdvanceWord(frame, LeadParamMarker, 1)

		assert.True(t, ok)
		assert.Equal(t, 7, frame.Dot.Col) // start of "world"
	})

	t.Run("FailsWhenNoNextWord", func(t *testing.T) {
		// Single word with null FLink → nextWord returns false
		frame, lines := buildWordFrame([]string{"hello"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordAdvanceWord(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("FailsWhenNoPreviousWord", func(t *testing.T) {
		// At first word with no BLink → previousWord returns false
		frame, lines := buildWordFrame([]string{"hello"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordAdvanceWord(frame, LeadParamNInt, -1)

		assert.False(t, ok)
	})
}

// ---- currentParagraph --------------------------------------------------

func TestCurrentParagraph(t *testing.T) {
	// Structure: para1 (line0), blank (line1), para2 (line2)
	setup := func() (*FrameObject, []*LineHdrObject) {
		return buildWordFrame([]string{"hello world", "", "foo bar"})
	}

	t.Run("AlreadyAtParaStart", func(t *testing.T) {
		_, lines := setup()
		dot := newDotMark(lines[0], 1)
		defer MarkDestroy(&dot)

		ok := currentParagraph(dot)

		assert.True(t, ok)
		assert.Equal(t, lines[0], dot.Line)
		assert.Equal(t, 1, dot.Col)
	})

	t.Run("MiddleOfFirstParagraph", func(t *testing.T) {
		_, lines := setup()
		dot := newDotMark(lines[0], 5) // 'o' in "hello"
		defer MarkDestroy(&dot)

		ok := currentParagraph(dot)

		assert.True(t, ok)
		assert.Equal(t, lines[0], dot.Line)
		assert.Equal(t, 1, dot.Col)
	})

	t.Run("SecondParagraphGoesToItsStart", func(t *testing.T) {
		_, lines := setup()
		dot := newDotMark(lines[2], 5) // 'b' in "foo bar"
		defer MarkDestroy(&dot)

		ok := currentParagraph(dot)

		assert.True(t, ok)
		assert.Equal(t, lines[2], dot.Line)
		assert.Equal(t, 1, dot.Col)
	})

	t.Run("BlankSeparatorGoesToPrecedingParaStart", func(t *testing.T) {
		// A blank line between paragraphs is treated as part of the preceding
		// paragraph by currentParagraph
		_, lines := setup()
		dot := newDotMark(lines[1], 1) // blank separator line
		defer MarkDestroy(&dot)

		ok := currentParagraph(dot)

		assert.True(t, ok)
		assert.Equal(t, lines[0], dot.Line) // preceding paragraph
		assert.Equal(t, 1, dot.Col)
	})

	t.Run("FailsOnBlankLineAtTopOfBuffer", func(t *testing.T) {
		// Blank line at very top (BLink is nil) → false
		_, lines := buildWordFrame([]string{"", "hello world"})
		dot := newDotMark(lines[0], 1)
		defer MarkDestroy(&dot)

		ok := currentParagraph(dot)

		assert.False(t, ok)
	})
}

// ---- nextParagraph -----------------------------------------------------

func TestNextParagraph(t *testing.T) {
	// Structure: para1 (line0), blank (line1), para2 (line2)
	setup := func() (*FrameObject, []*LineHdrObject) {
		return buildWordFrame([]string{"hello world", "", "foo bar"})
	}

	t.Run("AdvancesToNextParagraph", func(t *testing.T) {
		_, lines := setup()
		dot := newDotMark(lines[0], 1)
		defer MarkDestroy(&dot)

		ok := nextParagraph(dot)

		assert.True(t, ok)
		assert.Equal(t, lines[2], dot.Line) // "foo bar"
		assert.Equal(t, 1, dot.Col)
	})

	t.Run("AdvancesFromMiddleOfParagraph", func(t *testing.T) {
		_, lines := setup()
		dot := newDotMark(lines[0], 5) // 'o' in "hello"
		defer MarkDestroy(&dot)

		ok := nextParagraph(dot)

		assert.True(t, ok)
		assert.Equal(t, lines[2], dot.Line)
		assert.Equal(t, 1, dot.Col)
	})

	t.Run("FailsWhenNoNextParagraph", func(t *testing.T) {
		// Already at the last paragraph → no next → false
		_, lines := setup()
		dot := newDotMark(lines[2], 1)
		defer MarkDestroy(&dot)

		ok := nextParagraph(dot)

		assert.False(t, ok)
	})

	t.Run("FailsOnSingleParagraph", func(t *testing.T) {
		_, lines := buildWordFrame([]string{"hello world"})
		dot := newDotMark(lines[0], 1)
		defer MarkDestroy(&dot)

		ok := nextParagraph(dot)

		assert.False(t, ok)
	})
}

// ---- NewwordAdvanceParagraph -------------------------------------------

func TestNewwordAdvanceParagraph(t *testing.T) {
	// Structure: para1 (line0), blank (line1), para2 (line2)
	setup := func() (*FrameObject, []*LineHdrObject) {
		return buildWordFrame([]string{"hello world", "", "foo bar"})
	}

	t.Run("ForwardOneParagraph", func(t *testing.T) {
		frame, lines := setup()
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordAdvanceParagraph(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, lines[2], frame.Dot.Line)
		assert.Equal(t, 1, frame.Dot.Col)
	})

	t.Run("ForwardWithPInt", func(t *testing.T) {
		frame, lines := setup()
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 3

		ok := NewwordAdvanceParagraph(frame, LeadParamPInt, 1)

		assert.True(t, ok)
		assert.Equal(t, lines[2], frame.Dot.Line)
		assert.Equal(t, 1, frame.Dot.Col)
	})

	t.Run("ZeroCountGoesToCurrentParagraphStart", func(t *testing.T) {
		// LeadParamPInt count=0 → treated as NInt → currentParagraph
		frame, lines := setup()
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 5 // middle of "hello world"

		ok := NewwordAdvanceParagraph(frame, LeadParamPInt, 0)

		assert.True(t, ok)
		assert.Equal(t, lines[0], frame.Dot.Line)
		assert.Equal(t, 1, frame.Dot.Col) // start of paragraph
	})

	t.Run("BackwardOneParagraph", func(t *testing.T) {
		// From para2 → go back one paragraph → para1 start
		frame, lines := setup()
		frame.Dot.Line = lines[2]
		frame.Dot.Col = 3

		ok := NewwordAdvanceParagraph(frame, LeadParamNInt, -1)

		assert.True(t, ok)
		assert.Equal(t, lines[0], frame.Dot.Line)
		assert.Equal(t, 1, frame.Dot.Col)
	})

	t.Run("BackwardWithMinus", func(t *testing.T) {
		frame, lines := setup()
		frame.Dot.Line = lines[2]
		frame.Dot.Col = 1

		ok := NewwordAdvanceParagraph(frame, LeadParamMinus, -1)

		assert.True(t, ok)
		assert.Equal(t, lines[0], frame.Dot.Line)
	})

	t.Run("PIndefMovesToLastLine", func(t *testing.T) {
		// LeadParamPIndef → frame.LastGroup.LastLine at MarginLeft
		frame, _ := setup()
		lastLine := frame.LastGroup.LastLine
		frame.Dot.Line = frame.FirstGroup.FirstLine
		frame.Dot.Col = 1

		ok := NewwordAdvanceParagraph(frame, LeadParamPIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, lastLine, frame.Dot.Line)
		assert.Equal(t, frame.MarginLeft, frame.Dot.Col)
	})

	t.Run("NIndefMovesToFirstParagraph", func(t *testing.T) {
		// LeadParamNIndef → go to first non-blank line in frame
		frame, lines := setup()
		frame.Dot.Line = lines[2]
		frame.Dot.Col = 3

		ok := NewwordAdvanceParagraph(frame, LeadParamNIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, lines[0], frame.Dot.Line)
		assert.Equal(t, 1, frame.Dot.Col)
	})

	t.Run("NIndefFailsWhenNoPrecedingParagraph", func(t *testing.T) {
		// Blank lines at top of buffer with no paragraph behind → false
		frame, lines := buildWordFrame([]string{"", "", "foo bar"})
		frame.Dot.Line = lines[0] // blank line at top
		frame.Dot.Col = 1

		ok := NewwordAdvanceParagraph(frame, LeadParamNIndef, 0)

		assert.False(t, ok)
	})

	t.Run("FailsForwardWhenNoNextParagraph", func(t *testing.T) {
		// Already at last paragraph → forward fails
		frame, lines := setup()
		frame.Dot.Line = lines[2]
		frame.Dot.Col = 1

		ok := NewwordAdvanceParagraph(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("MarkerUsesMarkPosition", func(t *testing.T) {
		// LeadParamMarker: use mark position, then NInt count=0 = currentParagraph
		frame, lines := setup()
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1
		MarkCreate(lines[2], 3, &frame.Marks[1]) // mark 1 in para2
		defer MarkDestroy(&frame.Marks[1])

		ok := NewwordAdvanceParagraph(frame, LeadParamMarker, 1)

		assert.True(t, ok)
		// currentParagraph from lines[2] col 3 → lines[2] col 1
		assert.Equal(t, lines[2], frame.Dot.Line)
		assert.Equal(t, 1, frame.Dot.Col)
	})
}

// ---- NewwordDeleteWord -------------------------------------------------

func TestNewwordDeleteWord(t *testing.T) {
	// Use FrameOops=frame so delete takes TextRemove path (no oops buffer needed)
	withOops := func(t *testing.T, frame *FrameObject) {
		t.Helper()
		old := FrameOops
		FrameOops = frame
		t.Cleanup(func() { FrameOops = old })
	}

	t.Run("DeleteOneWordForward", func(t *testing.T) {
		// "hello world" → delete from "hello" to "world" → "world" remains
		frame, lines := buildWordFrame([]string{"hello world"})
		withOops(t, frame)
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordDeleteWord(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "world", getLineContent(lines[0]))
	})

	t.Run("DeleteOnlyWordFails", func(t *testing.T) {
		// Single word with no next word → second advance fails → false
		frame, lines := buildWordFrame([]string{"hello"})
		withOops(t, frame)
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordDeleteWord(frame, LeadParamNone, 1)

		assert.False(t, ok)
		// Dot restored to original position
		assert.Equal(t, 1, frame.Dot.Col)
	})
}

// ---- NewwordDeleteParagraph --------------------------------------------

func TestNewwordDeleteParagraph(t *testing.T) {
	withOops := func(t *testing.T, frame *FrameObject) {
		t.Helper()
		old := FrameOops
		FrameOops = frame
		t.Cleanup(func() { FrameOops = old })
	}

	t.Run("DeleteFirstParagraph", func(t *testing.T) {
		// para1 (line0), blank (line1), para2 (line2)
		// Delete para1 → para2 remains as first content
		frame, lines := buildWordFrame([]string{"hello world", "", "foo bar"})
		withOops(t, frame)
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := NewwordDeleteParagraph(frame, LeadParamNone, 1)

		assert.True(t, ok)
	})

	t.Run("DeleteFailsWhenNoNextParagraph", func(t *testing.T) {
		// Single paragraph, no next → forward advance fails → false, dot restored
		frame, lines := buildWordFrame([]string{"hello world"})
		withOops(t, frame)
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 3

		ok := NewwordDeleteParagraph(frame, LeadParamNone, 1)

		assert.False(t, ok)
		assert.Equal(t, 3, frame.Dot.Col) // dot restored
	})
}
