// Tests for arrow.go functions
package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsArrowCommand tests the isArrowCommand function
func TestIsArrowCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  Commands
		expected bool
	}{
		{"CmdReturn", CmdReturn, true},
		{"CmdHome", CmdHome, true},
		{"CmdTab", CmdTab, true},
		{"CmdBacktab", CmdBacktab, true},
		{"CmdLeft", CmdLeft, true},
		{"CmdRight", CmdRight, true},
		{"CmdDown", CmdDown, true},
		{"CmdUp", CmdUp, true},
		{"CmdDeleteLine", CmdDeleteLine, false},
		{"CmdInsertLine", CmdInsertLine, false},
		{"CmdQuit", CmdQuit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isArrowCommand(tt.command)
			assert.Equal(t, tt.expected, result, "isArrowCommand(%v)", tt.command)
		})
	}
}

// TestDoCmdLeft tests the left arrow command logic
func TestDoCmdLeft(t *testing.T) {
	t.Run("MoveLeftByOne", func(t *testing.T) {
		// Setup a minimal frame
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 10,
			},
			MarginLeft: 1,
		}

		var newEql MarkObject
		result := doCmdLeft(frame, LeadParamNone, 1, &newEql)

		assert.True(t, result, "Expected move left to succeed")
		assert.Equal(t, 9, frame.Dot.Col, "Expected column to be 9")
		assert.Equal(t, 10, newEql.Col, "Expected newEql to store old position")
	})

	t.Run("MoveLeftByCount", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 10,
			},
			MarginLeft: 1,
		}

		var newEql MarkObject
		result := doCmdLeft(frame, LeadParamPInt, 5, &newEql)

		assert.True(t, result, "Expected move left to succeed")
		assert.Equal(t, 5, frame.Dot.Col, "Expected column to be 5")
	})

	t.Run("MoveLeftBeyondBoundary", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 3,
			},
			MarginLeft: 1,
		}

		var newEql MarkObject
		result := doCmdLeft(frame, LeadParamNone, 5, &newEql)

		assert.False(t, result, "Expected move left to fail at boundary")
		assert.Equal(t, 3, frame.Dot.Col, "Expected column unchanged")
	})

	t.Run("MoveLeftToMargin", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 20,
			},
			MarginLeft: 5,
		}

		var newEql MarkObject
		result := doCmdLeft(frame, LeadParamPIndef, 0, &newEql)

		assert.True(t, result, "Expected move to margin to succeed")
		assert.Equal(t, 5, frame.Dot.Col, "Expected column at left margin")
	})

	t.Run("AlreadyAtMargin", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 5,
			},
			MarginLeft: 10,
		}

		var newEql MarkObject
		result := doCmdLeft(frame, LeadParamPIndef, 0, &newEql)

		assert.False(t, result, "Expected move to fail when already beyond margin")
		assert.Equal(t, 5, frame.Dot.Col, "Expected column unchanged")
	})
}

// TestDoCmdRight tests the right arrow command logic
func TestDoCmdRight(t *testing.T) {
	t.Run("MoveRightByOne", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 10,
			},
			MarginRight: MaxStrLenP,
		}

		var newEql MarkObject
		result := doCmdRight(frame, LeadParamNone, 1, &newEql)

		assert.True(t, result, "Expected move right to succeed")
		assert.Equal(t, 11, frame.Dot.Col, "Expected column to be 11")
		assert.Equal(t, 10, newEql.Col, "Expected newEql to store old position")
	})

	t.Run("MoveRightByCount", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 10,
			},
			MarginRight: MaxStrLenP,
		}

		var newEql MarkObject
		result := doCmdRight(frame, LeadParamPInt, 5, &newEql)

		assert.True(t, result, "Expected move right to succeed")
		assert.Equal(t, 15, frame.Dot.Col, "Expected column to be 15")
	})

	t.Run("MoveRightBeyondBoundary", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: MaxStrLenP - 2,
			},
			MarginRight: MaxStrLenP,
		}

		var newEql MarkObject
		result := doCmdRight(frame, LeadParamNone, 5, &newEql)

		assert.False(t, result, "Expected move right to fail at boundary")
		assert.Equal(t, MaxStrLenP-2, frame.Dot.Col, "Expected column unchanged")
	})

	t.Run("MoveRightToMargin", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 10,
			},
			MarginRight: 80,
		}

		var newEql MarkObject
		result := doCmdRight(frame, LeadParamPIndef, 0, &newEql)

		assert.True(t, result, "Expected move to margin to succeed")
		assert.Equal(t, 80, frame.Dot.Col, "Expected column at right margin")
	})

	t.Run("AlreadyBeyondMargin", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 90,
			},
			MarginRight: 80,
		}

		var newEql MarkObject
		result := doCmdRight(frame, LeadParamPIndef, 0, &newEql)

		assert.False(t, result, "Expected move to fail when already beyond margin")
		assert.Equal(t, 90, frame.Dot.Col, "Expected column unchanged")
	})
}

// TestDoCmdTabBacktab tests tab and backtab functionality
func TestDoCmdTabBacktab(t *testing.T) {
	t.Run("TabToNextStop", func(t *testing.T) {
		var tabStops TabArray
		tabStops[10] = true
		tabStops[20] = true
		tabStops[30] = true

		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 5,
			},
			TabStops:    tabStops,
			MarginLeft:  1,
			MarginRight: MaxStrLenP,
		}

		var newEql MarkObject
		result := doCmdTabBacktab(frame, 1, 1, &newEql)

		assert.True(t, result, "Expected tab to succeed")
		assert.Equal(t, 10, frame.Dot.Col, "Expected column at tab stop 10")
		assert.Equal(t, 5, newEql.Col, "Expected newEql to store old position")
	})

	t.Run("TabMultipleStops", func(t *testing.T) {
		var tabStops TabArray
		tabStops[10] = true
		tabStops[20] = true
		tabStops[30] = true

		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 5,
			},
			TabStops:    tabStops,
			MarginLeft:  1,
			MarginRight: MaxStrLenP,
		}

		var newEql MarkObject
		result := doCmdTabBacktab(frame, 1, 2, &newEql)

		assert.True(t, result, "Expected tab to succeed")
		assert.Equal(t, 20, frame.Dot.Col, "Expected column at tab stop 20")
	})

	t.Run("BacktabToPreviousStop", func(t *testing.T) {
		var tabStops TabArray
		tabStops[10] = true
		tabStops[20] = true
		tabStops[30] = true

		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 25,
			},
			TabStops:    tabStops,
			MarginLeft:  1,
			MarginRight: MaxStrLenP,
		}

		var newEql MarkObject
		result := doCmdTabBacktab(frame, -1, 1, &newEql)

		assert.True(t, result, "Expected backtab to succeed")
		assert.Equal(t, 20, frame.Dot.Col, "Expected column at tab stop 20")
	})

	t.Run("TabToMarginLeft", func(t *testing.T) {
		var tabStops TabArray
		// No explicit tab stops

		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 5,
			},
			TabStops:    tabStops,
			MarginLeft:  15,
			MarginRight: MaxStrLenP,
		}

		var newEql MarkObject
		result := doCmdTabBacktab(frame, 1, 1, &newEql)

		assert.True(t, result, "Expected tab to succeed")
		assert.Equal(t, 15, frame.Dot.Col, "Expected column at left margin")
	})

	t.Run("TabToMarginRight", func(t *testing.T) {
		var tabStops TabArray
		// No explicit tab stops between current and margin

		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 70,
			},
			TabStops:    tabStops,
			MarginLeft:  1,
			MarginRight: 80,
		}

		var newEql MarkObject
		result := doCmdTabBacktab(frame, 1, 1, &newEql)

		assert.True(t, result, "Expected tab to succeed")
		assert.Equal(t, 80, frame.Dot.Col, "Expected column at right margin")
	})

	t.Run("TabBeyondBoundary", func(t *testing.T) {
		var tabStops TabArray
		// No tab stops or margins that would stop before boundary

		frame := &FrameObject{
			Dot: &MarkObject{
				Col: MaxStrLenP - 2,
			},
			TabStops:    tabStops,
			MarginLeft:  1,
			MarginRight: MaxStrLenP + 10, // Beyond boundary
		}

		var newEql MarkObject
		result := doCmdTabBacktab(frame, 1, 1, &newEql)

		assert.False(t, result, "Expected tab to fail at boundary")
		assert.Equal(t, MaxStrLenP-2, frame.Dot.Col, "Expected column unchanged")
	})

	t.Run("BacktabBeyondBoundary", func(t *testing.T) {
		var tabStops TabArray
		// No tab stops or margins that would stop before boundary

		frame := &FrameObject{
			Dot: &MarkObject{
				Col: 2,
			},
			TabStops:    tabStops,
			MarginLeft:  -5, // Beyond boundary
			MarginRight: MaxStrLenP,
		}

		var newEql MarkObject
		result := doCmdTabBacktab(frame, -1, 1, &newEql)

		assert.False(t, result, "Expected backtab to fail at boundary")
		assert.Equal(t, 2, frame.Dot.Col, "Expected column unchanged")
	})
}

// TestDoCmdHome tests the home command
func TestDoCmdHome(t *testing.T) {
	t.Run("HomeInScreenFrame", func(t *testing.T) {
		topLine := &LineHdrObject{}

		frame := &FrameObject{
			Dot: &MarkObject{
				Line: &LineHdrObject{},
				Col:  50,
			},
			ScrOffset: 10,
		}

		oldScrFrame := ScrFrame
		oldScrTopLine := ScrTopLine

		ScrFrame = frame
		ScrTopLine = topLine

		defer func() {
			ScrFrame = oldScrFrame
			ScrTopLine = oldScrTopLine
		}()

		var newEql MarkObject
		doCmdHome(frame, &newEql)

		assert.Equal(t, 50, newEql.Col, "Expected newEql to store old column")
		assert.Equal(t, 11, frame.Dot.Col, "Expected column to be 1 after home")
	})

	t.Run("HomeInNonScreenFrame", func(t *testing.T) {
		frame := &FrameObject{
			Dot: &MarkObject{
				Line: &LineHdrObject{},
				Col:  50,
			},
		}

		otherFrame := &FrameObject{}
		oldScrFrame := ScrFrame
		ScrFrame = otherFrame // Different from frame

		defer func() {
			ScrFrame = oldScrFrame
		}()

		var newEql MarkObject
		doCmdHome(frame, &newEql)

		assert.Equal(t, 50, newEql.Col, "Expected newEql to store old column")
	})
}
