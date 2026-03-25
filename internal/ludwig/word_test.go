// Tests for word.go

package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---- WordLeft -----------------------------------------------------------

func TestWordLeft(t *testing.T) {
	setup := func(content ...string) (*FrameObject, []*LineHdrObject) {
		frame, lines := buildWordFrame(content)
		frame.MarginLeft = 3
		frame.MarginRight = 20
		return frame, lines
	}

	t.Run("AddSpacesToReachMargin", func(t *testing.T) {
		// "hello": startChar=1 < MarginLeft=3, insert 2 spaces → "  hello"
		frame, lines := setup("hello")

		ok := WordLeft(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "  hello", getLineContent(lines[0]))
	})

	t.Run("NoChangeWhenAlreadyAtMargin", func(t *testing.T) {
		// "  hello": startChar=3 == MarginLeft=3, no change
		frame, lines := setup("  hello")

		ok := WordLeft(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "  hello", getLineContent(lines[0]))
	})

	t.Run("RemoveExcessLeadingSpaces", func(t *testing.T) {
		// "    hello": startChar=5 > MarginLeft=3, remove 2 spaces → "  hello"
		frame, lines := setup("    hello")

		ok := WordLeft(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "  hello", getLineContent(lines[0]))
	})

	t.Run("BlankLineReturnsFalse", func(t *testing.T) {
		frame, _ := setup("")

		ok := WordLeft(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("UsedLessThanMarginLeftReturnsFalse", func(t *testing.T) {
		// "ab": Used=2 < MarginLeft=3 → main loop returns false
		frame, _ := setup("ab")

		ok := WordLeft(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("UsedGreaterThanMarginRightReturnsFalse", func(t *testing.T) {
		// 21-char string exceeds MarginRight=20
		frame, _ := setup("  hello world foo bar!")

		ok := WordLeft(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("TwoLinesWithCount2", func(t *testing.T) {
		frame, lines := setup("hello", "world")

		ok := WordLeft(frame, LeadParamPInt, 2)

		assert.True(t, ok)
		assert.Equal(t, "  hello", getLineContent(lines[0]))
		assert.Equal(t, "  world", getLineContent(lines[1]))
	})

	t.Run("PIndef_ProcessesAllNonBlankLines", func(t *testing.T) {
		frame, lines := setup("hello", "world")

		ok := WordLeft(frame, LeadParamPIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, "  hello", getLineContent(lines[0]))
		assert.Equal(t, "  world", getLineContent(lines[1]))
	})

	t.Run("DotAdvancesToNextLine", func(t *testing.T) {
		// After processing lines[0], dot moves to lines[1] at MarginLeft
		frame, lines := setup("hello", "world")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		WordLeft(frame, LeadParamNone, 1)

		assert.Equal(t, lines[1], frame.Dot.Line)
		assert.Equal(t, 3, frame.Dot.Col)
	})
}

// ---- WordRight ----------------------------------------------------------

func TestWordRight(t *testing.T) {
	setup := func(content ...string) (*FrameObject, []*LineHdrObject) {
		frame, lines := buildWordFrame(content)
		frame.MarginLeft = 3
		frame.MarginRight = 10
		return frame, lines
	}

	t.Run("AddSpacesToReachMarginRight", func(t *testing.T) {
		// "  hello": Used=7, MarginRight=10, spaceToAdd=3 → "     hello"
		frame, lines := setup("  hello")

		ok := WordRight(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "     hello", getLineContent(lines[0]))
	})

	t.Run("NoChangeWhenAlreadyAtMarginRight", func(t *testing.T) {
		// "     hello": Used=10 == MarginRight=10, spaceToAdd=0
		frame, lines := setup("     hello")

		ok := WordRight(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "     hello", getLineContent(lines[0]))
	})

	t.Run("StartCharBelowMarginLeftReturnsFalse", func(t *testing.T) {
		// " hello": startChar=2 < MarginLeft=3 → false
		frame, _ := setup(" hello")

		ok := WordRight(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("UsedBelowMarginLeftReturnsFalse", func(t *testing.T) {
		// "ab": Used=2 < MarginLeft=3 → false
		frame, _ := setup("ab")

		ok := WordRight(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("UsedAboveMarginRightReturnsFalse", func(t *testing.T) {
		// "  hello world!" Used=14 > MarginRight=10 → false
		frame, _ := setup("  hello world!")

		ok := WordRight(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("BlankLineReturnsFalse", func(t *testing.T) {
		frame, _ := setup("")

		ok := WordRight(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("TwoLinesWithCount2", func(t *testing.T) {
		frame, lines := setup("  hi", "  hi")

		ok := WordRight(frame, LeadParamPInt, 2)

		assert.True(t, ok)
		assert.Equal(t, "        hi", getLineContent(lines[0]))
		assert.Equal(t, "        hi", getLineContent(lines[1]))
	})
}

// ---- WordCentre ---------------------------------------------------------

func TestWordCentre(t *testing.T) {
	// MarginLeft=3, MarginRight=11
	// formula: spaceToAdd = (11-3-(Used-startChar))/2 - (startChar-3)
	setup := func(content ...string) (*FrameObject, []*LineHdrObject) {
		frame, lines := buildWordFrame(content)
		frame.MarginLeft = 3
		frame.MarginRight = 11
		return frame, lines
	}

	t.Run("AddSpacesToCentre", func(t *testing.T) {
		// "  hello": Used=7, startChar=3; spaceToAdd=(8-4)/2-0=2 → "    hello"
		frame, lines := setup("  hello")

		ok := WordCentre(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "    hello", getLineContent(lines[0]))
	})

	t.Run("NoChangeWhenAlreadyCentred", func(t *testing.T) {
		// "    hello": Used=9, startChar=5; spaceToAdd=(8-4)/2-2=0
		frame, lines := setup("    hello")

		ok := WordCentre(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "    hello", getLineContent(lines[0]))
	})

	t.Run("RemoveExcessLeadingSpaces", func(t *testing.T) {
		// "      hello": Used=11, startChar=7; spaceToAdd=(8-4)/2-4=-2 → "    hello"
		frame, lines := setup("      hello")

		ok := WordCentre(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "    hello", getLineContent(lines[0]))
	})

	t.Run("UsedAboveMarginRightReturnsFalse", func(t *testing.T) {
		// "  hello world": Used=13 > MarginRight=11 → false
		frame, _ := setup("  hello world")

		ok := WordCentre(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("UsedBelowMarginLeftReturnsFalse", func(t *testing.T) {
		// "ab": Used=2 < MarginLeft=3 → false
		frame, _ := setup("ab")

		ok := WordCentre(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("StartCharBelowMarginLeftReturnsFalse", func(t *testing.T) {
		// " hello": startChar=2 < MarginLeft=3 → false
		frame, _ := setup(" hello")

		ok := WordCentre(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("BlankLineReturnsFalse", func(t *testing.T) {
		frame, _ := setup("")

		ok := WordCentre(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("DotAdvancesToNextLine", func(t *testing.T) {
		frame, lines := setup("  hello", "world")
		frame.MarginRight = 20 // widen to keep "world" valid

		WordCentre(frame, LeadParamNone, 1)

		assert.Equal(t, lines[1], frame.Dot.Line)
		assert.Equal(t, 3, frame.Dot.Col)
	})
}

// ---- WordJustify --------------------------------------------------------

func TestWordJustify(t *testing.T) {
	setup := func(content ...string) (*FrameObject, []*LineHdrObject) {
		frame, lines := buildWordFrame(content)
		frame.MarginLeft = 3
		frame.MarginRight = 20
		return frame, lines
	}

	t.Run("JustifiesWhenNextLineNonBlank", func(t *testing.T) {
		// "  hello world" Used=13 + non-blank FLink → add 7 spaces → Used=20
		frame, lines := setup("  hello world", "next line")

		ok := WordJustify(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, 20, lines[0].Used)
	})

	t.Run("SkipsLastLineOfParagraph", func(t *testing.T) {
		// FLink (EOP) Used=0: skip justification, advance dot, return true
		frame, lines := setup("  hello world")
		before := getLineContent(lines[0])

		ok := WordJustify(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, before, getLineContent(lines[0]))
	})

	t.Run("UsedAboveMarginRightReturnsFalse", func(t *testing.T) {
		// Used=22 > MarginRight=20 with non-blank FLink → false
		frame, _ := setup("  hello world foo bar!", "next line")

		ok := WordJustify(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("BlankLineReturnsFalse", func(t *testing.T) {
		frame, _ := setup("")

		ok := WordJustify(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("SingleWordNoChange", func(t *testing.T) {
		// 1 word → holes=0, fillRatio=0, no insertion; still advances dot
		frame, lines := setup("  hello", "next line")
		before := getLineContent(lines[0])

		ok := WordJustify(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, before, getLineContent(lines[0]))
	})

	t.Run("TwoLinesWithCount2", func(t *testing.T) {
		// Process 2 lines; first line justified, second line is last of para (skipped)
		frame, lines := setup("  hello world", "  foo bar")

		ok := WordJustify(frame, LeadParamPInt, 2)

		assert.True(t, ok)
		assert.Equal(t, 20, lines[0].Used)
	})
}

// ---- WordSqueeze --------------------------------------------------------

func TestWordSqueeze(t *testing.T) {
	t.Run("CollapsesMultipleSpaces", func(t *testing.T) {
		// "hello   world" (3 spaces) → "hello world"
		frame, lines := buildWordFrame([]string{"hello   world"})

		ok := WordSqueeze(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "hello world", getLineContent(lines[0]))
	})

	t.Run("NoChangeWithSingleSpaces", func(t *testing.T) {
		frame, lines := buildWordFrame([]string{"hello world"})

		ok := WordSqueeze(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "hello world", getLineContent(lines[0]))
	})

	t.Run("PreservesLeadingSpaces", func(t *testing.T) {
		// "  hello   world" → "  hello world" (leading spaces preserved)
		frame, lines := buildWordFrame([]string{"  hello   world"})

		ok := WordSqueeze(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "  hello world", getLineContent(lines[0]))
	})

	t.Run("CollapsesTwoGaps", func(t *testing.T) {
		// "a   b   c" → "a b c"
		frame, lines := buildWordFrame([]string{"a   b   c"})

		ok := WordSqueeze(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "a b c", getLineContent(lines[0]))
	})

	t.Run("BlankLineReturnsFalse", func(t *testing.T) {
		frame, _ := buildWordFrame([]string{""})

		ok := WordSqueeze(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("TwoLinesWithCount2", func(t *testing.T) {
		frame, lines := buildWordFrame([]string{"a   b", "c   d"})

		ok := WordSqueeze(frame, LeadParamPInt, 2)

		assert.True(t, ok)
		assert.Equal(t, "a b", getLineContent(lines[0]))
		assert.Equal(t, "c d", getLineContent(lines[1]))
	})
}

// ---- WordFill -----------------------------------------------------------

func TestWordFill(t *testing.T) {
	t.Run("SplitsLineThatExceedsMarginRight", func(t *testing.T) {
		// "hello world" Used=11 > MarginRight=10 → split at space → "hello" + "world"
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.MarginRight = 10

		ok := WordFill(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "hello", getLineContent(lines[0]))
		assert.Equal(t, "world", getLineContent(lines[0].FLink))
	})

	t.Run("PullsWordFromNextLineWhenSpaceAvailable", func(t *testing.T) {
		// "hello" (5) + "hi" (2): spaceToAdd=4, word fits → "hello hi"; FLink deleted
		frame, lines := buildWordFrame([]string{"hello", "hi"})
		frame.MarginRight = 10

		ok := WordFill(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "hello hi", getLineContent(lines[0]))
	})

	t.Run("NoSplitNoPullWhenNextLineBlank", func(t *testing.T) {
		// "hello" (5 <= 10), FLink is blank → no change, returns true
		frame, lines := buildWordFrame([]string{"hello"})
		frame.MarginRight = 10

		ok := WordFill(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "hello", getLineContent(lines[0]))
	})

	t.Run("BlankLineReturnsFalse", func(t *testing.T) {
		frame, _ := buildWordFrame([]string{""})
		frame.MarginRight = 10

		ok := WordFill(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("CannotSplitSingleLongWordReturnsFalse", func(t *testing.T) {
		// Single word longer than MarginRight with no space to split at
		frame, _ := buildWordFrame([]string{"helloworld"})
		frame.MarginRight = 5

		ok := WordFill(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("PIndef_ProcessesUntilBlankLine", func(t *testing.T) {
		// With PIndef, processes both lines; "hello world" split gives count++
		frame, lines := buildWordFrame([]string{"hello world", "foo"})
		frame.MarginRight = 10

		ok := WordFill(frame, LeadParamPIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, "hello", getLineContent(lines[0]))
	})
}

// ---- WordAdvanceWord ----------------------------------------------------

func TestWordAdvanceWord(t *testing.T) {
	// "hello world": h=1,e=2,l=3,l=4,o=5,' '=6,w=7,o=8,r=9,l=10,d=11

	t.Run("MarkerReturnsFalse", func(t *testing.T) {
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := WordAdvanceWord(frame, LeadParamMarker, 1)

		assert.False(t, ok)
	})

	t.Run("ForwardOneWord_NoneParam", func(t *testing.T) {
		// "hello world": dot at col 1 → col 7 (start of "world")
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := WordAdvanceWord(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, lines[0], frame.Dot.Line)
		assert.Equal(t, 7, frame.Dot.Col)
	})

	t.Run("ForwardTwoWords_PIntParam", func(t *testing.T) {
		// "hello world foo": dot at col 1, count=2 → col 13 (start of "foo")
		frame, lines := buildWordFrame([]string{"hello world foo"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := WordAdvanceWord(frame, LeadParamPInt, 2)

		assert.True(t, ok)
		assert.Equal(t, 13, frame.Dot.Col)
	})

	t.Run("PIndef_AdvancesToEOPLine", func(t *testing.T) {
		// PIndef: scans to blank line then advances to EOP
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1
		eop := lines[0].FLink // EOP line (Used=0)

		ok := WordAdvanceWord(frame, LeadParamPIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, eop, frame.Dot.Line)
	})

	t.Run("NIndef_GoesToParagraphStart", func(t *testing.T) {
		// NIndef: from col 8, finds paragraph start at col 1
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 8

		ok := WordAdvanceWord(frame, LeadParamNIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, lines[0], frame.Dot.Line)
		assert.Equal(t, 1, frame.Dot.Col)
	})

	t.Run("BackwardOneWord_LandsAtPreviousWordStart", func(t *testing.T) {
		// "hello world", dot at col 8 (in "world"), NInt -1 → col 1 ("hello")
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 8

		ok := WordAdvanceWord(frame, LeadParamNInt, -1)

		assert.True(t, ok)
		assert.Equal(t, 1, frame.Dot.Col)
	})

	t.Run("BackwardZeroCount_LandsAtCurrentWordStart", func(t *testing.T) {
		// PInt count=0 → backward mode with count=0 → start of current word
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 8 // in "world"

		ok := WordAdvanceWord(frame, LeadParamPInt, 0)

		assert.True(t, ok)
		assert.Equal(t, 7, frame.Dot.Col) // start of "world"
	})

	t.Run("ForwardFails_WhenNoNextWord", func(t *testing.T) {
		// "hello" single word, FLink=EOP → no next word → false
		frame, lines := buildWordFrame([]string{"hello"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := WordAdvanceWord(frame, LeadParamNone, 1)

		assert.False(t, ok)
	})

	t.Run("BackwardFails_AtFirstWordWithNoBLink", func(t *testing.T) {
		// "hello world", col 1, NInt -1: reaches col 0, BLink=nil → false
		frame, lines := buildWordFrame([]string{"hello world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := WordAdvanceWord(frame, LeadParamNInt, -1)

		assert.False(t, ok)
	})

	t.Run("ForwardAcrossLines", func(t *testing.T) {
		// "hello" on line0, "world" on line1; dot at col 1 → moves to line1 col 1
		frame, lines := buildWordFrame([]string{"hello", "world"})
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := WordAdvanceWord(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, lines[1], frame.Dot.Line)
		assert.Equal(t, 1, frame.Dot.Col)
	})
}

// ---- WordDeleteWord -----------------------------------------------------

func TestWordDeleteWord(t *testing.T) {
	// Use FrameOops=frame so delete takes TextRemove path (no oops buffer needed)
	withFrameOops := func(t *testing.T, frame *FrameObject) {
		t.Helper()
		old := FrameOops
		FrameOops = frame
		t.Cleanup(func() { FrameOops = old })
	}

	t.Run("MarkerReturnsFalse", func(t *testing.T) {
		frame, lines := buildWordFrame([]string{"hello world"})
		withFrameOops(t, frame)
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := WordDeleteWord(frame, LeadParamMarker, 1)

		assert.False(t, ok)
	})

	t.Run("DeleteOneWordForward", func(t *testing.T) {
		// "hello world", dot at col 1, NoneCount=1 → deletes "hello " → "world"
		frame, lines := buildWordFrame([]string{"hello world"})
		withFrameOops(t, frame)
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := WordDeleteWord(frame, LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, "world", getLineContent(lines[0]))
	})

	t.Run("FailsWhenNoNextWord_DotRestored", func(t *testing.T) {
		// "hello" single word: second advance fails → dot restored → false
		frame, lines := buildWordFrame([]string{"hello"})
		withFrameOops(t, frame)
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 3

		ok := WordDeleteWord(frame, LeadParamNone, 1)

		assert.False(t, ok)
		assert.Equal(t, 3, frame.Dot.Col) // dot restored
	})
}
