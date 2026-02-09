// Tests for functions in line.go

package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper functions for test setup

// createTestFrame creates a minimal frame for testing
func createTestFrame() *FrameObject {
	frame := &FrameObject{
		SpaceLeft:  MaxSpace,
		SpaceLimit: MaxSpace,
	}
	return frame
}

// createTestFrameWithEOP creates a frame with an EOP group
func createTestFrameWithEOP() (*FrameObject, *GroupObject) {
	frame := createTestFrame()
	var eopGroup *GroupObject
	LineEOPCreate(frame, &eopGroup)
	frame.FirstGroup = eopGroup
	frame.LastGroup = eopGroup
	return frame, eopGroup
}

// createTestGroup creates a group with specified number of lines
func createTestGroup(frame *FrameObject, firstLineNr int, nrLines int) *GroupObject {
	group := &GroupObject{
		Frame:       frame,
		FirstLineNr: firstLineNr,
		NrLines:     nrLines,
	}

	var prevLine *LineHdrObject
	for i := 0; i < nrLines; i++ {
		line := &LineHdrObject{
			Group:    group,
			OffsetNr: i,
			Len:      0,
			Used:     0,
		}

		if prevLine != nil {
			prevLine.FLink = line
			line.BLink = prevLine
		} else {
			group.FirstLine = line
		}

		prevLine = line
	}

	if prevLine != nil {
		group.LastLine = prevLine
	}

	return group
}

// linkGroups links two groups together
func linkGroups(first, second *GroupObject) {
	if first != nil {
		first.FLink = second
	}
	if second != nil {
		second.BLink = first
	}
}

// validateGroupStructure checks all invariants in a group
func validateGroupStructure(t *testing.T, group *GroupObject, expectedFirstLineNr int) {
	t.Helper()

	assert.NotNil(t, group, "Group is nil")

	assert.Equal(t, expectedFirstLineNr, group.FirstLineNr, "Group FirstLineNr mismatch")

	// Count lines and verify linkage
	lineCount := 0
	line := group.FirstLine
	var prevLine *LineHdrObject

	for line != nil && lineCount < group.NrLines {
		assert.Equal(t, group, line.Group, "Line %d: group pointer mismatch", lineCount)
		assert.Equal(t, lineCount, line.OffsetNr, "Line %d: OffsetNr mismatch", lineCount)
		// Only check BLink if not the first line in the entire structure
		if lineCount > 0 {
			assert.Equal(t, prevLine, line.BLink, "Line %d: BLink mismatch", lineCount)
		}
		if prevLine != nil {
			assert.Equal(t, line, prevLine.FLink, "Line %d: previous FLink mismatch", lineCount)
		}

		prevLine = line
		line = line.FLink
		lineCount++
	}

	assert.Equal(t, group.NrLines, lineCount, "Group line count mismatch")
	assert.Equal(t, prevLine, group.LastLine, "Group LastLine mismatch")
}

// validateFrameStructure validates all groups in a frame
func validateFrameStructure(t *testing.T, frame *FrameObject) {
	t.Helper()

	if frame.FirstGroup == nil {
		return // Empty frame is valid
	}

	group := frame.FirstGroup
	expectedLineNr := 1
	var prevGroup *GroupObject

	for group != nil {
		assert.Equal(t, prevGroup, group.BLink, "Group BLink mismatch at line %d", expectedLineNr)
		if prevGroup != nil {
			assert.Equal(t, group, prevGroup.FLink, "Previous group FLink mismatch at line %d", expectedLineNr)
		}

		validateGroupStructure(t, group, expectedLineNr)
		expectedLineNr += group.NrLines

		prevGroup = group
		group = group.FLink
	}

	assert.Equal(t, prevGroup, frame.LastGroup, "Frame LastGroup mismatch")
}

// Tests for LineToNumber

func TestLineToNumber(t *testing.T) {
	t.Run("FirstLineInGroup", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 5)

		var lineNr int
		result := LineToNumber(group.FirstLine, &lineNr)

		assert.True(t, result, "LineToNumber returned false")
		assert.Equal(t, 1, lineNr, "Expected line number 1")
	})

	t.Run("LastLineInGroup", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 10, 5)

		var lineNr int
		result := LineToNumber(group.LastLine, &lineNr)

		assert.True(t, result, "LineToNumber returned false")
		assert.Equal(t, 14, lineNr, "Expected line number 14")
	})

	t.Run("MiddleLineInGroup", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 100, 10)

		// Get the middle line (offset 5, which is line 105)
		line := group.FirstLine
		for i := 0; i < 5; i++ {
			line = line.FLink
		}

		var lineNr int
		result := LineToNumber(line, &lineNr)

		assert.True(t, result, "LineToNumber returned false")
		assert.Equal(t, 105, lineNr, "Expected line number 105")
	})

	t.Run("LineInSecondGroup", func(t *testing.T) {
		frame := createTestFrame()
		group1 := createTestGroup(frame, 1, 10)
		group2 := createTestGroup(frame, 11, 10)
		linkGroups(group1, group2)

		var lineNr int
		result := LineToNumber(group2.FirstLine, &lineNr)

		assert.True(t, result, "LineToNumber returned false")
		assert.Equal(t, 11, lineNr, "Expected line number 11")
	})
}

// Tests for LineFromNumber

func TestLineFromNumber(t *testing.T) {
	t.Run("FirstLine", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 10)
		frame.FirstGroup = group
		frame.LastGroup = group

		var line *LineHdrObject
		result := LineFromNumber(frame, 1, &line)

		assert.True(t, result, "LineFromNumber returned false")
		assert.Equal(t, group.FirstLine, line, "Did not return first line")
	})

	t.Run("LastLine", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 10)
		frame.FirstGroup = group
		frame.LastGroup = group

		var line *LineHdrObject
		result := LineFromNumber(frame, 10, &line)

		assert.True(t, result, "LineFromNumber returned false")
		assert.Equal(t, group.LastLine, line, "Did not return last line")
	})

	t.Run("MiddleLine", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 10)
		frame.FirstGroup = group
		frame.LastGroup = group

		var line *LineHdrObject
		result := LineFromNumber(frame, 5, &line)

		assert.True(t, result, "LineFromNumber returned false")
		assert.Equal(t, 4, line.OffsetNr, "Expected offset 4")
	})

	t.Run("LineInSecondGroup", func(t *testing.T) {
		frame := createTestFrame()
		group1 := createTestGroup(frame, 1, 10)
		group2 := createTestGroup(frame, 11, 10)
		linkGroups(group1, group2)
		frame.FirstGroup = group1
		frame.LastGroup = group2

		var line *LineHdrObject
		result := LineFromNumber(frame, 15, &line)

		assert.True(t, result, "LineFromNumber returned false")
		assert.Equal(t, group2, line.Group, "Line not in second group")
		assert.Equal(t, 4, line.OffsetNr, "Expected offset 4")
	})

	t.Run("LineNumberTooLarge", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 10)
		frame.FirstGroup = group
		frame.LastGroup = group

		var line *LineHdrObject
		result := LineFromNumber(frame, 20, &line)

		assert.True(t, result, "LineFromNumber returned false")
		assert.Nil(t, line, "Expected nil for out-of-range line number")
	})

	t.Run("MultipleGroupsSearch", func(t *testing.T) {
		frame := createTestFrame()
		group1 := createTestGroup(frame, 1, 20)
		group2 := createTestGroup(frame, 21, 30)
		group3 := createTestGroup(frame, 51, 25)
		linkGroups(group1, group2)
		linkGroups(group2, group3)
		frame.FirstGroup = group1
		frame.LastGroup = group3

		var line *LineHdrObject
		result := LineFromNumber(frame, 60, &line)

		assert.True(t, result, "LineFromNumber returned false")
		assert.Equal(t, group3, line.Group, "Line not in third group")
		assert.Equal(t, 9, line.OffsetNr, "Expected offset 9")
	})
}

// Tests for LineChangeLength

func TestLineChangeLength(t *testing.T) {
	t.Run("ExpandFromZero", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 1)
		line := group.FirstLine

		originalSpace := frame.SpaceLeft

		result := LineChangeLength(line, 15)

		assert.True(t, result, "LineChangeLength returned false")
		assert.NotNil(t, line.Str, "Str should not be nil")
		// Should be quantized to 20
		assert.Equal(t, 20, line.Len, "Expected length 20 (quantized)")
		assert.Equal(t, originalSpace-20, frame.SpaceLeft, "SpaceLeft not updated correctly")
	})

	t.Run("ExpandLength", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 1)
		line := group.FirstLine

		// Set initial length
		LineChangeLength(line, 10)
		space1 := frame.SpaceLeft

		// Expand to 25
		result := LineChangeLength(line, 25)

		assert.True(t, result, "LineChangeLength returned false")
		// Should be quantized to 30
		assert.Equal(t, 30, line.Len, "Expected length 30 (quantized)")
		// Space change: -30 + 20 = -10 from space1
		expectedSpace := space1 - 30 + 20
		assert.Equal(t, expectedSpace, frame.SpaceLeft, "SpaceLeft not updated correctly")
	})

	t.Run("ShrinkLength", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 1)
		line := group.FirstLine

		// Set initial length
		LineChangeLength(line, 100)
		oldLen := line.Len // Will be 110 (quantized from 100)
		space1 := frame.SpaceLeft

		// Shrink to 20
		result := LineChangeLength(line, 20)

		assert.True(t, result, "LineChangeLength returned false")
		// Should be quantized to 30 (20/10 + 1) * 10 = 30
		assert.Equal(t, 30, line.Len, "Expected length 30 (quantized)")
		// Space freed: oldLen - 30
		expectedSpace := space1 + oldLen - 30
		assert.Equal(t, expectedSpace, frame.SpaceLeft, "SpaceLeft not updated correctly")
	})

	t.Run("SetToZero", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 1)
		line := group.FirstLine

		// Set initial length
		LineChangeLength(line, 50)
		space1 := frame.SpaceLeft

		// Set to zero
		result := LineChangeLength(line, 0)

		assert.True(t, result, "LineChangeLength returned false")
		assert.Nil(t, line.Str, "Str should be nil when length is 0")
		assert.Equal(t, 0, line.Len, "Expected length 0")
		// All space should be freed
		assert.Equal(t, space1+60, frame.SpaceLeft, "SpaceLeft not updated correctly")
	})

	t.Run("QuantizationBehavior", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 1)
		line := group.FirstLine

		// Quantization formula: (newLength/10 + 1) * 10
		// This always rounds UP to next 10
		testCases := []struct {
			requested int
			expected  int
		}{
			{1, 10},                     // (1/10 + 1) * 10 = 10
			{5, 10},                     // (5/10 + 1) * 10 = 10
			{9, 10},                     // (9/10 + 1) * 10 = 10
			{10, 20},                    // (10/10 + 1) * 10 = 20
			{11, 20},                    // (11/10 + 1) * 10 = 20
			{19, 20},                    // (19/10 + 1) * 10 = 20
			{20, 30},                    // (20/10 + 1) * 10 = 30
			{21, 30},                    // (21/10 + 1) * 10 = 30
			{95, 100},                   // (95/10 + 1) * 10 = 100
			{MaxStrLen - 20, 390},       // (380/10 + 1) * 10 = 390 (still below threshold)
			{MaxStrLen - 11, 390},       // (389/10 + 1) * 10 = 390 (still below threshold)
			{MaxStrLen - 10, MaxStrLen}, // >= MaxStrLen - 10, caps to MaxStrLen
			{MaxStrLen - 5, MaxStrLen},  // >= MaxStrLen - 10
			{MaxStrLen, MaxStrLen},
			{MaxStrLen + 10, MaxStrLen}, // Should cap at MaxStrLen
		}

		for _, tc := range testCases {
			result := LineChangeLength(line, tc.requested)
			assert.True(t, result, "LineChangeLength(%d) returned false", tc.requested)
			assert.Equal(t, tc.expected, line.Len, "LineChangeLength(%d): expected %d", tc.requested, tc.expected)
		}
	})

	t.Run("LineWithoutGroup", func(t *testing.T) {
		// Line not attached to a group (no frame)
		line := &LineHdrObject{
			Len: 0,
		}

		result := LineChangeLength(line, 20)

		assert.True(t, result, "LineChangeLength returned false")
		// Should be quantized to 30
		assert.Equal(t, 30, line.Len, "Expected length 30 (quantized)")
		// Should not panic even without a group
	})
}

// Tests for LinesCreate and LinesDestroy

func TestLinesCreate(t *testing.T) {
	t.Run("CreateSingleLine", func(t *testing.T) {
		var firstLine, lastLine *LineHdrObject

		result := LinesCreate(1, &firstLine, &lastLine)

		assert.True(t, result, "LinesCreate returned false")
		assert.NotNil(t, firstLine, "firstLine is nil")
		assert.NotNil(t, lastLine, "lastLine is nil")
		assert.Equal(t, lastLine, firstLine, "For single line, firstLine should equal lastLine")
		assert.Nil(t, firstLine.FLink, "Single line FLink should be nil")
		assert.Nil(t, firstLine.BLink, "Single line BLink should be nil")
	})

	t.Run("CreateMultipleLines", func(t *testing.T) {
		var firstLine, lastLine *LineHdrObject

		result := LinesCreate(5, &firstLine, &lastLine)

		assert.True(t, result, "LinesCreate returned false")
		assert.NotNil(t, firstLine, "firstLine is nil")
		assert.NotNil(t, lastLine, "lastLine is nil")

		// Count lines
		count := 0
		line := firstLine
		for line != nil {
			count++
			if line.FLink == nil {
				assert.Equal(t, lastLine, line, "Last line in chain doesn't match lastLine")
				break
			}
			line = line.FLink
		}

		assert.Equal(t, 5, count, "Expected 5 lines")
	})

	t.Run("VerifyLinkage", func(t *testing.T) {
		var firstLine, lastLine *LineHdrObject

		LinesCreate(10, &firstLine, &lastLine)

		// Verify forward and backward linkage
		line := firstLine
		var prevLine *LineHdrObject
		count := 0

		for line != nil {
			assert.Equal(t, prevLine, line.BLink, "Line %d: BLink mismatch", count)
			prevLine = line
			line = line.FLink
			count++
		}

		assert.Equal(t, 10, count, "Expected 10 lines")
	})

	t.Run("ZeroLines", func(t *testing.T) {
		var firstLine, lastLine *LineHdrObject

		result := LinesCreate(0, &firstLine, &lastLine)

		assert.True(t, result, "LinesCreate returned false")
		// With 0 lines, both should remain nil (or undefined)
	})

	t.Run("VerifyInitialization", func(t *testing.T) {
		var firstLine, lastLine *LineHdrObject

		LinesCreate(3, &firstLine, &lastLine)

		line := firstLine
		for line != nil {
			assert.Nil(t, line.Group, "Line Group should be nil")
			assert.Equal(t, 0, line.OffsetNr, "Line OffsetNr should be 0")
			assert.Nil(t, line.Marks, "Line Marks should be nil")
			assert.Nil(t, line.Str, "Line Str should be nil")
			assert.Equal(t, 0, line.Len, "Line Len should be 0")
			assert.Equal(t, 0, line.Used, "Line Used should be 0")
			assert.Equal(t, 0, line.ScrRowNr, "Line ScrRowNr should be 0")
			line = line.FLink
		}
	})
}

func TestLinesDestroy(t *testing.T) {
	t.Run("DestroySingleLine", func(t *testing.T) {
		var firstLine, lastLine *LineHdrObject
		LinesCreate(1, &firstLine, &lastLine)

		result := LinesDestroy(&firstLine, &lastLine)

		assert.True(t, result, "LinesDestroy returned false")
		assert.Nil(t, firstLine, "firstLine should be nil after destroy")
		assert.Nil(t, lastLine, "lastLine should be nil after destroy")
	})

	t.Run("DestroyMultipleLines", func(t *testing.T) {
		var firstLine, lastLine *LineHdrObject
		LinesCreate(10, &firstLine, &lastLine)

		result := LinesDestroy(&firstLine, &lastLine)

		assert.True(t, result, "LinesDestroy returned false")
		assert.Nil(t, firstLine, "firstLine should be nil after destroy")
		assert.Nil(t, lastLine, "lastLine should be nil after destroy")
	})

	t.Run("DestroyLinesWithStrObjects", func(t *testing.T) {
		var firstLine, lastLine *LineHdrObject
		LinesCreate(5, &firstLine, &lastLine)

		// Add Str objects to some lines
		line := firstLine
		for i := 0; i < 3 && line != nil; i++ {
			line.Str = NewFilled(' ', MaxStrLen)
			line.Len = 10
			line = line.FLink
		}

		result := LinesDestroy(&firstLine, &lastLine)

		assert.True(t, result, "LinesDestroy returned false")
		// Should not panic and should clean up Str objects
	})

	t.Run("DestroyNilLines", func(t *testing.T) {
		var firstLine, lastLine *LineHdrObject

		result := LinesDestroy(&firstLine, &lastLine)

		assert.True(t, result, "LinesDestroy returned false")
		// Should handle nil gracefully
	})
}

// Tests for LineEOPCreate and LineEOPDestroy

func TestLineEOPCreate(t *testing.T) {
	t.Run("CreateEOPGroup", func(t *testing.T) {
		frame := createTestFrame()
		var group *GroupObject

		result := LineEOPCreate(frame, &group)

		assert.True(t, result, "LineEOPCreate returned false")
		assert.NotNil(t, group, "group is nil")
		assert.Equal(t, frame, group.Frame, "Group frame mismatch")
		assert.NotNil(t, group.FirstLine, "FirstLine is nil")
		assert.NotNil(t, group.LastLine, "LastLine is nil")
		assert.Equal(t, group.LastLine, group.FirstLine, "FirstLine should equal LastLine for EOP")
		assert.Equal(t, 1, group.FirstLineNr, "Expected FirstLineNr=1")
		assert.Equal(t, 1, group.NrLines, "Expected NrLines=1")
	})

	t.Run("VerifyEOPLineInitialization", func(t *testing.T) {
		frame := createTestFrame()
		var group *GroupObject

		LineEOPCreate(frame, &group)

		line := group.FirstLine
		assert.Nil(t, line.FLink, "EOP line FLink should be nil")
		assert.Nil(t, line.BLink, "EOP line BLink should be nil")
		assert.Equal(t, group, line.Group, "EOP line Group mismatch")
		assert.Equal(t, 0, line.OffsetNr, "EOP line OffsetNr should be 0")
		assert.Nil(t, line.Marks, "EOP line Marks should be nil")
		assert.Nil(t, line.Str, "EOP line Str should be nil")
		assert.Equal(t, 0, line.Len, "EOP line Len should be 0")
		assert.Equal(t, 0, line.Used, "EOP line Used should be 0")
	})
}

func TestLineEOPDestroy(t *testing.T) {
	t.Run("DestroyEOPGroup", func(t *testing.T) {
		frame := createTestFrame()
		var group *GroupObject
		LineEOPCreate(frame, &group)

		result := LineEOPDestroy(&group)

		assert.True(t, result, "LineEOPDestroy returned false")
		assert.Nil(t, group, "group should be nil after destroy")
	})

	t.Run("DestroyEOPWithStr", func(t *testing.T) {
		frame := createTestFrame()
		var group *GroupObject
		LineEOPCreate(frame, &group)

		// Add a Str object to the EOP line
		group.FirstLine.Str = NewFilled(' ', MaxStrLen)
		group.FirstLine.Len = 10

		result := LineEOPDestroy(&group)

		assert.True(t, result, "LineEOPDestroy returned false")
		assert.Nil(t, group, "group should be nil after destroy")
	})
}

// Tests for GroupsDestroy

func TestGroupsDestroy(t *testing.T) {
	t.Run("DestroySingleGroup", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 5)
		firstGroup := group
		lastGroup := group

		result := GroupsDestroy(&firstGroup, &lastGroup)

		assert.True(t, result, "GroupsDestroy returned false")
		assert.Nil(t, firstGroup, "firstGroup should be nil")
		assert.Nil(t, lastGroup, "lastGroup should be nil")
	})

	t.Run("DestroyMultipleGroups", func(t *testing.T) {
		frame := createTestFrame()
		group1 := createTestGroup(frame, 1, 10)
		group2 := createTestGroup(frame, 11, 10)
		group3 := createTestGroup(frame, 21, 10)
		linkGroups(group1, group2)
		linkGroups(group2, group3)

		firstGroup := group1
		lastGroup := group3

		result := GroupsDestroy(&firstGroup, &lastGroup)

		assert.True(t, result, "GroupsDestroy returned false")
		assert.Nil(t, firstGroup, "firstGroup should be nil")
		assert.Nil(t, lastGroup, "lastGroup should be nil")
	})
}

// Tests for LinesInject - Basic cases

func TestLinesInject_Basic(t *testing.T) {
	// Save and clear global screen variables
	origScrFrame := ScrFrame
	origScrTopLine := ScrTopLine
	origScrBotLine := ScrBotLine

	t.Cleanup(func() {
		ScrFrame = origScrFrame
		ScrTopLine = origScrTopLine
		ScrBotLine = origScrBotLine
	})

	ScrFrame = nil
	ScrTopLine = nil
	ScrBotLine = nil

	t.Run("InjectIntoEOPGroup", func(t *testing.T) {
		frame, eopGroup := createTestFrameWithEOP()
		beforeLine := eopGroup.FirstLine

		// Create lines to inject
		var firstLine, lastLine *LineHdrObject
		LinesCreate(3, &firstLine, &lastLine)

		// Set some content on lines for space tracking
		line := firstLine
		for i := 0; i < 3; i++ {
			LineChangeLength(line, 10)
			line.Used = 10
			line = line.FLink
		}

		originalSpace := frame.SpaceLeft

		result := LinesInject(firstLine, lastLine, beforeLine)

		assert.True(t, result, "LinesInject returned false")

		// Verify structure
		validateFrameStructure(t, frame)

		// Check line numbers
		var lineNr int
		LineToNumber(firstLine, &lineNr)
		assert.Equal(t, 1, lineNr, "First injected line should be line 1")

		LineToNumber(beforeLine, &lineNr)
		assert.Equal(t, 4, lineNr, "EOP line should now be line 4")

		// Check space was updated
		expectedSpace := originalSpace - 3*20 // 3 lines with length 20 each (quantized from 10)
		assert.Equal(t, expectedSpace, frame.SpaceLeft, "SpaceLeft: expected %d", expectedSpace)
	})

	t.Run("InjectSingleLine", func(t *testing.T) {
		frame, eopGroup := createTestFrameWithEOP()
		beforeLine := eopGroup.FirstLine

		var firstLine, lastLine *LineHdrObject
		LinesCreate(1, &firstLine, &lastLine)

		result := LinesInject(firstLine, lastLine, beforeLine)

		assert.True(t, result, "LinesInject returned false")

		validateFrameStructure(t, frame)

		// Verify the line is properly linked
		assert.Equal(t, firstLine, beforeLine.BLink, "beforeLine.BLink should be firstLine")
		assert.Equal(t, beforeLine, firstLine.FLink, "firstLine.FLink should be beforeLine")
	})

	t.Run("InjectWithinGroupCapacity", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 10)
		frame.FirstGroup = group
		frame.LastGroup = group

		// Get a line in the middle to inject before
		beforeLine := group.FirstLine
		for i := 0; i < 5; i++ {
			beforeLine = beforeLine.FLink
		}

		originalGroupCount := 1
		originalNrLines := group.NrLines

		var firstLine, lastLine *LineHdrObject
		LinesCreate(5, &firstLine, &lastLine)

		result := LinesInject(firstLine, lastLine, beforeLine)

		assert.True(t, result, "LinesInject returned false")

		validateFrameStructure(t, frame)

		// Should still be one group if within capacity
		groupCount := 0
		g := frame.FirstGroup
		for g != nil {
			groupCount++
			g = g.FLink
		}

		// We started with 10 lines and added 5, total 15 which fits in one group
		if groupCount != originalGroupCount {
			t.Logf("Group count changed from %d to %d (may need new groups)", originalGroupCount, groupCount)
		}

		// Total lines should be correct
		totalLines := 0
		g = frame.FirstGroup
		for g != nil {
			totalLines += g.NrLines
			g = g.FLink
		}
		expectedTotal := originalNrLines + 5
		assert.Equal(t, expectedTotal, totalLines, "Total lines: expected %d", expectedTotal)
	})
}

func TestLinesInject_RequiringNewGroups(t *testing.T) {
	// Save and clear global screen variables
	origScrFrame := ScrFrame
	origScrTopLine := ScrTopLine
	origScrBotLine := ScrBotLine

	t.Cleanup(func() {
		ScrFrame = origScrFrame
		ScrTopLine = origScrTopLine
		ScrBotLine = origScrBotLine
	})

	ScrFrame = nil
	ScrTopLine = nil
	ScrBotLine = nil

	t.Run("InjectExceedingGroupCapacity", func(t *testing.T) {
		frame, eopGroup := createTestFrameWithEOP()
		beforeLine := eopGroup.FirstLine

		// Create more lines than MaxGroupLines
		lineCount := MaxGroupLines + 10
		var firstLine, lastLine *LineHdrObject
		LinesCreate(lineCount, &firstLine, &lastLine)

		result := LinesInject(firstLine, lastLine, beforeLine)

		assert.True(t, result, "LinesInject returned false")

		validateFrameStructure(t, frame)

		// Count total lines and groups
		totalLines := 0
		groupCount := 0
		g := frame.FirstGroup
		for g != nil {
			totalLines += g.NrLines
			groupCount++
			g = g.FLink
		}

		expectedTotal := lineCount + 1 // +1 for EOP
		assert.Equal(t, expectedTotal, totalLines, "Total lines: expected %d", expectedTotal)

		// Should have created multiple groups
		assert.GreaterOrEqual(t, groupCount, 2, "Expected at least 2 groups")
	})

	t.Run("InjectTwoMaxGroupLines", func(t *testing.T) {
		frame, eopGroup := createTestFrameWithEOP()
		beforeLine := eopGroup.FirstLine

		// Create exactly 2 * MaxGroupLines
		lineCount := 2 * MaxGroupLines
		var firstLine, lastLine *LineHdrObject
		LinesCreate(lineCount, &firstLine, &lastLine)

		result := LinesInject(firstLine, lastLine, beforeLine)

		assert.True(t, result, "LinesInject returned false")

		validateFrameStructure(t, frame)

		// Count groups
		groupCount := 0
		g := frame.FirstGroup
		for g != nil {
			groupCount++
			g = g.FLink
		}

		// Should have at least 3 groups (2 full + EOP group)
		assert.GreaterOrEqual(t, groupCount, 3, "Expected at least 3 groups")
	})
}

// Tests for LinesExtract - Basic cases

func TestLinesExtract_Basic(t *testing.T) {
	// Save and clear global screen variables
	origScrFrame := ScrFrame
	origScrTopLine := ScrTopLine
	origScrBotLine := ScrBotLine

	t.Cleanup(func() {
		ScrFrame = origScrFrame
		ScrTopLine = origScrTopLine
		ScrBotLine = origScrBotLine
	})

	ScrFrame = nil
	ScrTopLine = nil
	ScrBotLine = nil

	t.Run("ExtractSingleLine", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 10)
		frame.FirstGroup = group
		frame.LastGroup = group

		// Extract the middle line (offset 5)
		extractLine := group.FirstLine
		for i := 0; i < 5; i++ {
			extractLine = extractLine.FLink
		}

		// Add some content so space is tracked
		LineChangeLength(extractLine, 10)
		originalSpace := frame.SpaceLeft

		result := LinesExtract(extractLine, extractLine)

		assert.True(t, result, "LinesExtract returned false")

		validateFrameStructure(t, frame)

		// Should have 9 lines now
		assert.Equal(t, 9, group.NrLines, "Expected 9 lines in group")

		// Verify extracted line is unlinked
		assert.Nil(t, extractLine.FLink, "Extracted line FLink should be nil")
		assert.Nil(t, extractLine.BLink, "Extracted line BLink should be nil")

		// Space should be freed
		assert.Greater(t, frame.SpaceLeft, originalSpace, "SpaceLeft should have increased: was %d", originalSpace)
	})

	t.Run("ExtractMultipleLines", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 10)
		frame.FirstGroup = group
		frame.LastGroup = group

		// Extract lines 3-5 (offsets 2-4)
		firstLine := group.FirstLine
		for i := 0; i < 2; i++ {
			firstLine = firstLine.FLink
		}
		lastLine := firstLine.FLink.FLink

		result := LinesExtract(firstLine, lastLine)

		assert.True(t, result, "LinesExtract returned false")

		validateFrameStructure(t, frame)

		// Should have 7 lines now
		assert.Equal(t, 7, group.NrLines, "Expected 7 lines in group")
	})

	t.Run("ExtractFirstLine", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 10)
		frame.FirstGroup = group
		frame.LastGroup = group

		firstLine := group.FirstLine

		result := LinesExtract(firstLine, firstLine)

		assert.True(t, result, "LinesExtract returned false")

		validateFrameStructure(t, frame)

		assert.Equal(t, 9, group.NrLines, "Expected 9 lines")
		assert.NotEqual(t, firstLine, group.FirstLine, "FirstLine should have changed")
	})

	t.Run("ExtractLastLine", func(t *testing.T) {
		frame := createTestFrame()
		group := createTestGroup(frame, 1, 10)
		// Add an EOP group after the main group
		var eopGroup *GroupObject
		LineEOPCreate(frame, &eopGroup)
		eopGroup.FirstLineNr = 11

		linkGroups(group, eopGroup)
		// Link the last line of group to the EOP line
		group.LastLine.FLink = eopGroup.FirstLine
		eopGroup.FirstLine.BLink = group.LastLine

		frame.FirstGroup = group
		frame.LastGroup = eopGroup

		lastLine := group.LastLine

		result := LinesExtract(lastLine, lastLine)

		assert.True(t, result, "LinesExtract returned false")

		validateFrameStructure(t, frame)

		assert.Equal(t, 9, group.NrLines, "Expected 9 lines")
		assert.NotEqual(t, lastLine, group.LastLine, "LastLine should have changed")
		// Verify the group's last line now links to EOP
		assert.Equal(t, eopGroup.FirstLine, group.LastLine.FLink, "Group's last line should link to EOP")
	})

	t.Run("ExtractAllLinesInGroup", func(t *testing.T) {
		frame := createTestFrame()
		group1 := createTestGroup(frame, 1, 10)
		group2 := createTestGroup(frame, 11, 10)
		group3 := createTestGroup(frame, 21, 10)
		linkGroups(group1, group2)
		linkGroups(group2, group3)

		// Link lines between groups
		group1.LastLine.FLink = group2.FirstLine
		group2.FirstLine.BLink = group1.LastLine
		group2.LastLine.FLink = group3.FirstLine
		group3.FirstLine.BLink = group2.LastLine

		frame.FirstGroup = group1
		frame.LastGroup = group3

		firstLine := group2.FirstLine
		lastLine := group2.LastLine

		result := LinesExtract(firstLine, lastLine)

		assert.True(t, result, "LinesExtract returned false")

		validateFrameStructure(t, frame)

		// group2 should be removed
		assert.Equal(t, group3, frame.FirstGroup.FLink, "group2 should have been removed")

		// Verify line numbering in remaining groups
		assert.Equal(t, 11, group3.FirstLineNr, "group3 FirstLineNr should be 11")

		// Verify group1 now links directly to group3
		assert.Equal(t, group3.FirstLine, group1.LastLine.FLink, "group1 should now link directly to group3")
	})
}

func TestLinesExtract_MultipleGroups(t *testing.T) {
	// Save and clear global screen variables
	origScrFrame := ScrFrame
	origScrTopLine := ScrTopLine
	origScrBotLine := ScrBotLine

	t.Cleanup(func() {
		ScrFrame = origScrFrame
		ScrTopLine = origScrTopLine
		ScrBotLine = origScrBotLine
	})

	ScrFrame = nil
	ScrTopLine = nil
	ScrBotLine = nil

	t.Run("ExtractSpanningTwoGroups", func(t *testing.T) {
		frame := createTestFrame()
		group1 := createTestGroup(frame, 1, 20)
		group2 := createTestGroup(frame, 21, 20)
		linkGroups(group1, group2)

		// Link lines between groups
		group1.LastLine.FLink = group2.FirstLine
		group2.FirstLine.BLink = group1.LastLine

		// Add an EOP group at the end
		var eopGroup *GroupObject
		LineEOPCreate(frame, &eopGroup)
		eopGroup.FirstLineNr = 41
		linkGroups(group2, eopGroup)
		group2.LastLine.FLink = eopGroup.FirstLine
		eopGroup.FirstLine.BLink = group2.LastLine

		frame.FirstGroup = group1
		frame.LastGroup = eopGroup

		// Extract from line 15 to line 25 (spans both groups)
		firstLine := group1.FirstLine
		for i := 0; i < 14; i++ {
			firstLine = firstLine.FLink
		}
		lastLine := group2.FirstLine
		for i := 0; i < 4; i++ {
			lastLine = lastLine.FLink
		}

		result := LinesExtract(firstLine, lastLine)

		assert.True(t, result, "LinesExtract returned false")

		validateFrameStructure(t, frame)

		// Count total remaining lines
		totalLines := 0
		g := frame.FirstGroup
		for g != nil {
			totalLines += g.NrLines
			g = g.FLink
		}

		expectedTotal := 40 - 11 + 1 // Started with 40, removed 11 lines (15-25 inclusive), plus 1 EOP
		assert.Equal(t, expectedTotal, totalLines, "Total lines: expected %d", expectedTotal)
	})

	t.Run("ExtractLeavingEmptyGroups", func(t *testing.T) {
		frame := createTestFrame()
		group1 := createTestGroup(frame, 1, 10)
		group2 := createTestGroup(frame, 11, 10)
		group3 := createTestGroup(frame, 21, 10)
		linkGroups(group1, group2)
		linkGroups(group2, group3)

		// Link lines between groups
		group1.LastLine.FLink = group2.FirstLine
		group2.FirstLine.BLink = group1.LastLine
		group2.LastLine.FLink = group3.FirstLine
		group3.FirstLine.BLink = group2.LastLine

		// Add EOP after group3
		var eopGroup *GroupObject
		LineEOPCreate(frame, &eopGroup)
		eopGroup.FirstLineNr = 31
		linkGroups(group3, eopGroup)
		group3.LastLine.FLink = eopGroup.FirstLine
		eopGroup.FirstLine.BLink = group3.LastLine

		frame.FirstGroup = group1
		frame.LastGroup = eopGroup

		// Extract all of group2
		firstLine := group2.FirstLine
		lastLine := group2.LastLine

		result := LinesExtract(firstLine, lastLine)

		assert.True(t, result, "LinesExtract returned false")

		validateFrameStructure(t, frame)

		// Count groups (should be 3 now: group1, group3, eopGroup)
		groupCount := 0
		g := frame.FirstGroup
		for g != nil {
			groupCount++
			g = g.FLink
		}

		assert.Equal(t, 3, groupCount, "Expected 3 groups after removing empty group")

		// Verify group3 line numbering adjusted
		assert.Equal(t, 11, group3.FirstLineNr, "group3 FirstLineNr: expected 11")

		// Verify eopGroup line numbering adjusted
		assert.Equal(t, 21, eopGroup.FirstLineNr, "eopGroup FirstLineNr: expected 21")
	})
}

// Integration tests

func TestLinesInjectExtractIntegration(t *testing.T) {
	// Save and clear global screen variables
	origScrFrame := ScrFrame
	origScrTopLine := ScrTopLine
	origScrBotLine := ScrBotLine

	t.Cleanup(func() {
		ScrFrame = origScrFrame
		ScrTopLine = origScrTopLine
		ScrBotLine = origScrBotLine
	})

	ScrFrame = nil
	ScrTopLine = nil
	ScrBotLine = nil

	t.Run("InjectThenExtractSameLines", func(t *testing.T) {
		frame, eopGroup := createTestFrameWithEOP()
		beforeLine := eopGroup.FirstLine

		// Create and inject lines
		var firstLine, lastLine *LineHdrObject
		LinesCreate(5, &firstLine, &lastLine)
		LinesInject(firstLine, lastLine, beforeLine)

		originalSpace := frame.SpaceLeft

		// Extract the same lines
		result := LinesExtract(firstLine, lastLine)

		assert.True(t, result, "LinesExtract returned false")

		validateFrameStructure(t, frame)

		// Should be back to just EOP
		assert.Equal(t, 1, frame.FirstGroup.NrLines, "Expected 1 line (EOP)")

		// Space should be restored
		assert.Equal(t, originalSpace, frame.SpaceLeft, "SpaceLeft: expected %d", originalSpace)
	})

	t.Run("MultipleInjectExtractOperations", func(t *testing.T) {
		frame, eopGroup := createTestFrameWithEOP()
		beforeLine := eopGroup.FirstLine

		// Operation 1: Inject 10 lines
		var first1, last1 *LineHdrObject
		LinesCreate(10, &first1, &last1)
		LinesInject(first1, last1, beforeLine)

		validateFrameStructure(t, frame)

		// Operation 2: Inject 5 more lines at beginning
		var first2, last2 *LineHdrObject
		LinesCreate(5, &first2, &last2)
		LinesInject(first2, last2, first1)

		validateFrameStructure(t, frame)

		// Operation 3: Extract middle 5 lines
		extractFirst := first1
		extractLast := first1
		for i := 0; i < 4; i++ {
			extractLast = extractLast.FLink
		}
		LinesExtract(extractFirst, extractLast)

		validateFrameStructure(t, frame)

		// Count total lines
		totalLines := 0
		g := frame.FirstGroup
		for g != nil {
			totalLines += g.NrLines
			g = g.FLink
		}

		expectedTotal := 5 + 5 + 1 // 5 from op2, 5 remaining from op1, 1 EOP
		assert.Equal(t, expectedTotal, totalLines, "Total lines: expected %d", expectedTotal)
	})
}
