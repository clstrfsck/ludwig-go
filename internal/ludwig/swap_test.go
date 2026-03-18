// Tests for swap.go

package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupSwapFrame builds a frame suitable for SwapLine tests.
// It uses setupLinkedLines (group.LastLine = nullLine) which is required for
// LinesInject to work correctly, and initialises frame.Dot on lines[0].
func setupSwapFrame(lineCount int) (*FrameObject, []*LineHdrObject) {
	frame, lines := setupLinkedLines(lineCount)
	frame.Dot = &MarkObject{
		Line: lines[0],
		Col:  1,
	}
	return frame, lines
}

// collectLineContents walks the linked list from start, collecting content of
// each real line. It stops at the null sentinel (the line with FLink==nil).
func collectLineContents(start *LineHdrObject) []string {
	var result []string
	for line := start; line != nil && line.FLink != nil; line = line.FLink {
		result = append(result, getLineContent(line))
	}
	return result
}

func TestSwapLine(t *testing.T) {
	oldCurrentFrame := CurrentFrame
	defer func() { CurrentFrame = oldCurrentFrame }()

	t.Run("Returns false when no next line", func(t *testing.T) {
		// Dot on the null sentinel line (FLink==nil) — nothing to swap with
		frame, lines := setupSwapFrame(2)
		CurrentFrame = frame
		nullLine := lines[1].FLink
		frame.Dot.Line = nullLine
		frame.Dot.Col = 1

		assert.False(t, SwapLine(LeadParamNone, 1))
	})

	t.Run("Returns false when forward count exceeds list", func(t *testing.T) {
		// Dot on the last real line: count=1 would require nullLine.FLink (nil)
		frame, lines := setupSwapFrame(2)
		CurrentFrame = frame
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		assert.False(t, SwapLine(LeadParamNone, 1))
	})

	t.Run("Returns false when backward count exceeds list", func(t *testing.T) {
		// Dot on the first line: BLink is nil, cannot go further back
		frame, lines := setupSwapFrame(2)
		CurrentFrame = frame
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		assert.False(t, SwapLine(LeadParamMinus, -1))
	})

	t.Run("Basic forward swap", func(t *testing.T) {
		// Dot on lines[0], count=1: lines[0] content moves past lines[1]
		// Before: alpha → beta → gamma → delta
		// After:  beta → alpha → gamma → delta
		frame, lines := setupSwapFrame(4)
		setLineContent(lines[0], "alpha")
		setLineContent(lines[1], "beta")
		setLineContent(lines[2], "gamma")
		setLineContent(lines[3], "delta")
		CurrentFrame = frame
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := SwapLine(LeadParamNone, 1)

		assert.True(t, ok)
		assert.True(t, frame.TextModified)
		assert.Equal(t, []string{"beta", "alpha", "gamma", "delta"},
			collectLineContents(frame.FirstGroup.FirstLine))
	})

	t.Run("Forward swap by two", func(t *testing.T) {
		// Dot on lines[0], count=2: lines[0] content moves past lines[1] and lines[2]
		// Before: alpha → beta → gamma → delta → epsilon
		// After:  beta → gamma → alpha → delta → epsilon
		frame, lines := setupSwapFrame(5)
		setLineContent(lines[0], "alpha")
		setLineContent(lines[1], "beta")
		setLineContent(lines[2], "gamma")
		setLineContent(lines[3], "delta")
		setLineContent(lines[4], "epsilon")
		CurrentFrame = frame
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := SwapLine(LeadParamPInt, 2)

		assert.True(t, ok)
		assert.Equal(t, []string{"beta", "gamma", "alpha", "delta", "epsilon"},
			collectLineContents(frame.FirstGroup.FirstLine))
	})

	t.Run("Backward swap one", func(t *testing.T) {
		// Dot on lines[2], count=-1: lines[2] content moves before lines[1]
		// Before: alpha → beta → gamma → delta
		// After:  alpha → gamma → beta → delta
		frame, lines := setupSwapFrame(4)
		setLineContent(lines[0], "alpha")
		setLineContent(lines[1], "beta")
		setLineContent(lines[2], "gamma")
		setLineContent(lines[3], "delta")
		CurrentFrame = frame
		frame.Dot.Line = lines[2]
		frame.Dot.Col = 1

		ok := SwapLine(LeadParamMinus, -1)

		assert.True(t, ok)
		assert.Equal(t, []string{"alpha", "gamma", "beta", "delta"},
			collectLineContents(frame.FirstGroup.FirstLine))
	})

	t.Run("Backward swap two", func(t *testing.T) {
		// Dot on lines[3], count=-2: lines[3] content moves before lines[1]
		// Before: alpha → beta → gamma → delta → epsilon
		// After:  alpha → delta → beta → gamma → epsilon
		frame, lines := setupSwapFrame(5)
		setLineContent(lines[0], "alpha")
		setLineContent(lines[1], "beta")
		setLineContent(lines[2], "gamma")
		setLineContent(lines[3], "delta")
		setLineContent(lines[4], "epsilon")
		CurrentFrame = frame
		frame.Dot.Line = lines[3]
		frame.Dot.Col = 1

		ok := SwapLine(LeadParamNInt, -2)

		assert.True(t, ok)
		assert.Equal(t, []string{"alpha", "delta", "beta", "gamma", "epsilon"},
			collectLineContents(frame.FirstGroup.FirstLine))
	})

	t.Run("Pindef moves to last line", func(t *testing.T) {
		// destLine = LastGroup.LastLine = nullLine; content moves to end of buffer
		// Before: first → second → third → fourth → fifth
		// After:  second → third → fourth → fifth → first
		frame, lines := setupSwapFrame(5)
		setLineContent(lines[0], "first")
		setLineContent(lines[1], "second")
		setLineContent(lines[2], "third")
		setLineContent(lines[3], "fourth")
		setLineContent(lines[4], "fifth")
		CurrentFrame = frame
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := SwapLine(LeadParamPIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, []string{"second", "third", "fourth", "fifth", "first"},
			collectLineContents(frame.FirstGroup.FirstLine))
	})

	t.Run("Nindef moves to first line", func(t *testing.T) {
		// destLine = FirstGroup.FirstLine = lines[0]; content moves to start of buffer
		// Before: first → second → third → fourth → fifth
		// After:  fourth → first → second → third → fifth
		frame, lines := setupSwapFrame(5)
		setLineContent(lines[0], "first")
		setLineContent(lines[1], "second")
		setLineContent(lines[2], "third")
		setLineContent(lines[3], "fourth")
		setLineContent(lines[4], "fifth")
		CurrentFrame = frame
		frame.Dot.Line = lines[3]
		frame.Dot.Col = 1

		ok := SwapLine(LeadParamNIndef, 0)

		assert.True(t, ok)
		assert.Equal(t, []string{"fourth", "first", "second", "third", "fifth"},
			collectLineContents(frame.FirstGroup.FirstLine))
	})

	t.Run("Marker swap", func(t *testing.T) {
		// Place MarkEquals on lines[3]; dot on lines[0]
		// lines[0] content moves to just before lines[3]
		// Before: first → second → third → fourth → fifth
		// After:  second → third → first → fourth → fifth
		frame, lines := setupSwapFrame(5)
		setLineContent(lines[0], "first")
		setLineContent(lines[1], "second")
		setLineContent(lines[2], "third")
		setLineContent(lines[3], "fourth")
		setLineContent(lines[4], "fifth")
		CurrentFrame = frame
		MarkCreate(lines[3], 1, &frame.Marks[MarkEquals])
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := SwapLine(LeadParamMarker, MarkEquals)

		assert.True(t, ok)
		assert.Equal(t, []string{"second", "third", "first", "fourth", "fifth"},
			collectLineContents(frame.FirstGroup.FirstLine))
	})

	t.Run("Dot follows swapped line", func(t *testing.T) {
		// After swap, dot should be on the line now holding the moved content
		frame, lines := setupSwapFrame(4)
		setLineContent(lines[0], "moved")
		setLineContent(lines[1], "stays")
		CurrentFrame = frame
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		SwapLine(LeadParamNone, 1)

		assert.Equal(t, "moved", getLineContent(frame.Dot.Line),
			"dot should be on the line holding the moved content")
	})

	t.Run("Dot column preserved", func(t *testing.T) {
		frame, lines := setupSwapFrame(4)
		setLineContent(lines[0], "line with content")
		setLineContent(lines[1], "another line")
		CurrentFrame = frame
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 7

		ok := SwapLine(LeadParamNone, 1)

		assert.True(t, ok)
		assert.Equal(t, 7, frame.Dot.Col, "dot column should be preserved after swap")
	})

	t.Run("Text modified set on success", func(t *testing.T) {
		frame, lines := setupSwapFrame(3)
		CurrentFrame = frame
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1
		frame.TextModified = false

		SwapLine(LeadParamNone, 1)

		assert.True(t, frame.TextModified)
	})

	t.Run("Text modified not set on failure", func(t *testing.T) {
		// Backward past start returns false without modifying the buffer
		frame, lines := setupSwapFrame(2)
		CurrentFrame = frame
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1
		frame.TextModified = false

		SwapLine(LeadParamMinus, -1)

		assert.False(t, frame.TextModified)
	})

	t.Run("Modified mark set on success", func(t *testing.T) {
		frame, lines := setupSwapFrame(3)
		CurrentFrame = frame
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		ok := SwapLine(LeadParamNone, 1)

		assert.True(t, ok)
		assert.NotNil(t, frame.Marks[MarkModified],
			"MarkModified should be set after a successful swap")
	})
}
