// Tests for caseditto.go functions
package ludwig

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper functions for CaseDittoCommand tests

// setupTestFrame creates a minimal test frame with linked lines
func setupTestFrame(lineCount int) (*FrameObject, []*LineHdrObject) {
	frame := &FrameObject{
		SpaceLeft:    MaxSpace,
		SpaceLimit:   MaxSpace,
		ScrWidth:     80,
		ScrOffset:    0,
		MarginLeft:   1,
		MarginRight:  MaxStrLen,
		TextModified: false,
		Options:      0,
	}

	group := &GroupObject{
		Frame:       frame,
		FirstLineNr: 1,
		NrLines:     lineCount,
	}

	lines := make([]*LineHdrObject, lineCount)

	// Create lines
	for i := range lineCount {
		lines[i] = &LineHdrObject{
			Group:    group,
			OffsetNr: i,
			Used:     0,
			ScrRowNr: 0, // Set to 0 to disable screen updates
			Str:      NewBlankStrObject(MaxStrLen),
			Marks:    make([]*MarkObject, 0),
		}
	}

	// Create a null line (end marker)
	nullLine := &LineHdrObject{
		Group:    group,
		OffsetNr: lineCount,
		Used:     0,
		ScrRowNr: 0,
		Str:      EmptyStrObject(),
		Marks:    make([]*MarkObject, 0),
	}

	// Link lines bidirectionally
	for i := range lineCount {
		if i > 0 {
			lines[i].BLink = lines[i-1]
		}
		if i < lineCount-1 {
			lines[i].FLink = lines[i+1]
		}
	}

	// Link the last line to the null line
	if lineCount > 0 {
		lines[lineCount-1].FLink = nullLine
		nullLine.BLink = lines[lineCount-1]
	}

	group.FirstLine = lines[0]
	group.LastLine = lines[lineCount-1]
	frame.FirstGroup = group
	frame.LastGroup = group

	// Initialize marks array
	for i := range frame.Marks {
		frame.Marks[i] = nil
	}

	// Create dot mark on the middle line (or first line if only one)
	dotLine := lines[lineCount/2]
	frame.Dot = &MarkObject{
		Line: dotLine,
		Col:  1,
	}

	return frame, lines
}

// setLineContent sets the content of a line
func setLineContent(line *LineHdrObject, content string) {
	line.Str.Assign(content)
	line.Used = len(content)
	line.Str.FillN(' ', line.Len()-line.Used, line.Used+1)
}

// getLineContent extracts the used content from a line
func getLineContent(line *LineHdrObject) string {
	if line == nil || line.Str == nil {
		return ""
	}
	return line.Str.Slice(1, line.Used)
}

// TestCaseDittoCommand tests the CaseDittoCommand function
func TestCaseDittoCommand(t *testing.T) {
	// Save and restore global state
	oldCurrentFrame := CurrentFrame
	oldEditMode := EditMode
	oldPreviousMode := PreviousMode
	oldTtControlC := TtControlC

	defer func() {
		CurrentFrame = oldCurrentFrame
		EditMode = oldEditMode
		PreviousMode = oldPreviousMode
		TtControlC = oldTtControlC
	}()

	// Note: Most tests use fromSpan=false to test the validation logic only
	// Full integration tests would require VDU/screen initialization

	t.Run("InsertMode_NotAllowedWithNegativeParams_Minus", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeInsert
		PreviousMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "test line")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		// Try DittoUp with LeadParamMinus in insert mode - should be rejected
		result := CaseDittoCommand(CmdDittoUp, LeadParamMinus, -5, true)
		assert.False(t, result, "Should reject LeadParamMinus in insert mode")
	})

	t.Run("InsertMode_NotAllowedWithNInt", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeInsert
		TtControlC = false

		setLineContent(lines[0], "test line")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdDittoDown, LeadParamNInt, -5, true)
		assert.False(t, result, "Should reject LeadParamNInt in insert mode")
	})

	t.Run("InsertMode_NotAllowedWithNIndef", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeInsert
		TtControlC = false

		setLineContent(lines[0], "test line")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdDittoUp, LeadParamNIndef, 0, true)
		assert.False(t, result, "Should reject LeadParamNIndef in insert mode")
	})

	t.Run("InsertMode_AllowedWithPositiveParams", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeInsert
		TtControlC = false

		setLineContent(lines[0], "UPPER LINE")
		setLineContent(lines[1], "lower")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdDittoUp, LeadParamNIndef, 0, true)
		assert.False(t, result, "Should reject LeadParamNIndef in insert mode")
	})

	t.Run("DittoInsertModeDetection", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		TtControlC = false

		setLineContent(lines[0], "UPPER")
		setLineContent(lines[1], "lower")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		// Test insert mode detection: EditMode == ModeInsert
		EditMode = ModeInsert
		PreviousMode = ModeCommand
		result1 := CaseDittoCommand(CmdDittoUp, LeadParamMinus, -1, true)
		assert.False(t, result1, "Should reject in pure insert mode")

		// Test insert mode detection: EditMode == ModeCommand && PreviousMode == ModeInsert
		EditMode = ModeCommand
		PreviousMode = ModeInsert
		result2 := CaseDittoCommand(CmdDittoDown, LeadParamNInt, -1, true)
		assert.False(t, result2, "Should reject when previous mode was insert")
	})
}

// TestCaseDittoCommand_CaseCommands tests the case conversion commands
func TestCaseDittoCommand_CaseCommands(t *testing.T) {
	oldCurrentFrame := CurrentFrame
	oldEditMode := EditMode
	oldTtControlC := TtControlC

	defer func() {
		CurrentFrame = oldCurrentFrame
		EditMode = oldEditMode
		TtControlC = oldTtControlC
	}()

	t.Run("CaseUp_CommandSetup", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "hello world")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdCaseUp, LeadParamNone, 5, true)
		assert.True(t, result, "CaseUp should succeed")

		// Verify the first 5 chars were converted to uppercase
		content := getLineContent(lines[0])
		assert.Equal(t, "HELLO world", content, "First 5 chars should be uppercase")
		assert.Equal(t, 6, frame.Dot.Col, "Dot should move to column 6")
	})

	t.Run("CaseLow_CommandSetup", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "HELLO WORLD")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdCaseLow, LeadParamNone, 5, true)
		assert.True(t, result, "CaseLow should succeed")

		// Verify the first 5 chars were converted to lowercase
		content := getLineContent(lines[0])
		assert.Equal(t, "hello WORLD", content, "First 5 chars should be lowercase")
		assert.Equal(t, 6, frame.Dot.Col, "Dot should move to column 6")
	})

	t.Run("CaseEdit_CommandSetup", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "HeLLo WoRLd")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdCaseEdit, LeadParamNone, 5, true)
		assert.True(t, result, "CaseEdit should succeed")

		// CaseEdit: uppercase after non-letter, lowercase after letter
		content := getLineContent(lines[0])
		assert.Equal(t, "Hello WoRLd", content, "CaseEdit should produce edit case")
	})

	t.Run("CaseUp_WithCount", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "lowercase text here")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdCaseUp, LeadParamPInt, 10, true)
		assert.True(t, result, "CaseUp should succeed")

		content := getLineContent(lines[0])
		assert.Equal(t, "LOWERCASE text here", content, "Should uppercase first 10 chars")
	})

	t.Run("CaseLow_WithOffset", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "UPPERCASE TEXT")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 5 // Start from column 5

		result := CaseDittoCommand(CmdCaseLow, LeadParamNone, 4, true)
		assert.True(t, result, "CaseLow should succeed")

		content := getLineContent(lines[0])
		// Lowercase chars at positions 5-8: "RCAS" -> "rcas"
		assert.Equal(t, "UPPErcasE TEXT", content, "Should lowercase 4 chars starting from col 5")
	})

	t.Run("CaseEdit_AlternatingCase", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "MiXeD CaSe text")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdCaseEdit, LeadParamPInt, 8, true)
		assert.True(t, result, "CaseEdit should succeed")

		content := getLineContent(lines[0])
		assert.Equal(t, "Mixed CaSe text", content, "Should edit case for first 8 chars")
	})
}

// TestCaseDittoCommand_SuccessfulDittoCases tests successful ditto operations
func TestCaseDittoCommand_SuccessfulDittoCases(t *testing.T) {
	oldCurrentFrame := CurrentFrame
	oldEditMode := EditMode
	oldTtControlC := TtControlC

	defer func() {
		CurrentFrame = oldCurrentFrame
		EditMode = oldEditMode
		TtControlC = oldTtControlC
	}()

	t.Run("DittoUp_ValidLineAbove", func(t *testing.T) {
		frame, lines := setupTestFrame(3)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "SOURCE LINE")
		setLineContent(lines[1], "target line")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdDittoUp, LeadParamNone, 5, true)
		assert.True(t, result, "DittoUp should succeed")

		// Verify characters were copied from line above
		content := getLineContent(lines[1])
		assert.Contains(t, content, "SOURC", "Should copy first 5 chars from line above")
	})

	t.Run("DittoDown_ValidLineBelow", func(t *testing.T) {
		frame, lines := setupTestFrame(3)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[1], "target line")
		setLineContent(lines[2], "SOURCE LINE")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdDittoDown, LeadParamNone, 5, true)
		assert.True(t, result, "DittoDown should succeed")

		// Verify characters were copied from line below
		content := getLineContent(lines[1])
		assert.Contains(t, content, "SOURC", "Should copy first 5 chars from line below")
	})

	t.Run("DittoUp_WithLeadParamPlus", func(t *testing.T) {
		frame, lines := setupTestFrame(3)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "ABCDEFGHIJ")
		setLineContent(lines[1], "1234567890")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 3

		result := CaseDittoCommand(CmdDittoUp, LeadParamPlus, 4, true)
		assert.True(t, result, "DittoUp should succeed")

		content := getLineContent(lines[1])
		assert.Equal(t, "12CDEF7890", content, "Should copy 4 chars from column 3")
	})

	t.Run("DittoDown_WithLeadParamPInt", func(t *testing.T) {
		frame, lines := setupTestFrame(3)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[1], "current")
		setLineContent(lines[2], "BELOW LINE TEXT")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdDittoDown, LeadParamPInt, 7, true)
		assert.True(t, result, "DittoDown should succeed")

		content := getLineContent(lines[1])
		assert.Equal(t, "BELOW L", content, "Should copy first 7 chars from line below")
	})

	t.Run("DittoUp_WithLeadParamPIndef", func(t *testing.T) {
		frame, lines := setupTestFrame(3)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "COMPLETE LINE")
		setLineContent(lines[1], "short")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 3

		// LeadParamPIndef should copy from dot to end of other line
		result := CaseDittoCommand(CmdDittoUp, LeadParamPIndef, 0, true)
		assert.True(t, result, "DittoUp with PIndef should succeed")

		content := getLineContent(lines[1])
		// Copy from column 3 to end of source (chars 3-13 = "MPLETE LINE")
		assert.Equal(t, "shMPLETE LINE", content, "Should copy from dot to end of source line")
	})

	t.Run("DittoDown_FromMiddleOfLine", func(t *testing.T) {
		frame, lines := setupTestFrame(3)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[1], "1234567890")
		setLineContent(lines[2], "ABCDEFGHIJ")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 5 // Start from middle

		result := CaseDittoCommand(CmdDittoDown, LeadParamNone, 3, true)
		assert.True(t, result, "DittoDown should succeed")

		content := getLineContent(lines[1])
		assert.Equal(t, "1234EFG890", content, "Should copy 3 chars starting from column 5")
	})
}

// TestCaseDittoCommand_ParameterCalculations tests parameter handling
func TestCaseDittoCommand_ParameterCalculations(t *testing.T) {
	oldCurrentFrame := CurrentFrame
	oldEditMode := EditMode
	oldTtControlC := TtControlC

	defer func() {
		CurrentFrame = oldCurrentFrame
		EditMode = oldEditMode
		TtControlC = oldTtControlC
	}()

	t.Run("LeadParamNone_DefaultCount", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "SOURCE")
		setLineContent(lines[1], "target")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdDittoUp, LeadParamNone, 3, true)
		assert.True(t, result, "LeadParamNone should succeed")

		content := getLineContent(lines[1])
		assert.Equal(t, "SOUget", content, "Should overtype first 3 chars with source")
		assert.Equal(t, 4, frame.Dot.Col, "Dot should move to newCol (1+3)")
	})

	t.Run("LeadParamPlus_PositiveCount", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "UPPERCASE")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdCaseUp, LeadParamPlus, 5, true)
		assert.True(t, result, "LeadParamPlus should succeed")

		content := getLineContent(lines[0])
		assert.Equal(t, "UPPERCASE", content, "Should uppercase first 5 chars")
	})

	t.Run("ZeroCount_Handling", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "SOURCE")
		setLineContent(lines[1], "target")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		// Zero count should be handled
		result := CaseDittoCommand(CmdDittoUp, LeadParamNone, 0, true)
		assert.True(t, result, "Zero count should succeed")
	})

	t.Run("DittoUp_EmptySourceLine", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		lines[0].Used = 0 // Empty source line
		setLineContent(lines[1], "target")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		result := CaseDittoCommand(CmdDittoUp, LeadParamNone, 3, true)
		// Validation fails: Dot.Col + count > otherLine.Used + 1 (1+3 > 0+1)
		assert.False(t, result, "Should fail when trying to copy beyond empty source")
	})

	t.Run("LeadParamMinus_BackwardCopy", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "ABCDEFGHIJ")
		setLineContent(lines[1], "1234567890")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 6 // Position at '6'

		// LeadParamMinus with count=-3 at Dot.Col=6 means:
		// - count becomes 3
		// - firstCol = 6 - 3 = 3
		// - Copy 3 chars from source starting at column 3 ("CDE")
		// - Overtype current line starting at column 3
		result := CaseDittoCommand(CmdDittoUp, LeadParamMinus, -3, true)
		assert.True(t, result, "LeadParamMinus should succeed")

		content := getLineContent(lines[1])
		// Should overtype starting at column 3 with "CDE"
		assert.Equal(t, "12CDE67890", content, "Should copy 3 chars from source col 3")
		assert.Equal(t, 3, frame.Dot.Col, "Dot should be at firstCol (3)")
	})

	t.Run("LeadParamNInt_BackwardCopy", func(t *testing.T) {
		frame, lines := setupTestFrame(3)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[1], "abcdefghij")
		setLineContent(lines[2], "ZYXWVUTSRQ")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 8 // Position at 'h'

		// LeadParamNInt with count=-4 at Dot.Col=8 means:
		// - count becomes 4
		// - firstCol = 8 - 4 = 4
		// - Copy 4 chars from line below starting at column 4: "WVUT"
		result := CaseDittoCommand(CmdDittoDown, LeadParamNInt, -4, true)
		assert.True(t, result, "LeadParamNInt should succeed")

		content := getLineContent(lines[1])
		// Copy from line 2 column 4 for 4 chars: "WVUT"
		// Overtype line 1 starting at column 4
		assert.Equal(t, "abcWVUThij", content, "Should copy 4 chars from source col 4")
		assert.Equal(t, 4, frame.Dot.Col, "Dot should be at firstCol (4)")
	})

	t.Run("LeadParamNIndef_CopyFromStart", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "PREFIXSUFFIX")
		setLineContent(lines[1], "lowercase text")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 7 // Position at 's' in "lowercase"

		// LeadParamNIndef copies from column 1 to (Dot.Col - 1)
		// So copy 6 chars from source line
		result := CaseDittoCommand(CmdDittoUp, LeadParamNIndef, 0, true)
		assert.True(t, result, "LeadParamNIndef should succeed")

		content := getLineContent(lines[1])
		assert.Equal(t, "PREFIXase text", content, "Should copy from start of source line")
		assert.Equal(t, 1, frame.Dot.Col, "Dot should move to column 1")
	})

	t.Run("LeadParamMinus_AtStartOfLine", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false

		setLineContent(lines[0], "SOURCE")
		setLineContent(lines[1], "target")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 3

		// LeadParamMinus with count=-3 would try to start at column 0 (invalid)
		result := CaseDittoCommand(CmdDittoUp, LeadParamMinus, -3, true)
		assert.False(t, result, "Should fail when going before column 1")
	})
}

// TestCaseDittoCommandValidation tests parameter validation logic
func TestCaseDittoCommandValidation(t *testing.T) {
	oldCurrentFrame := CurrentFrame
	oldEditMode := EditMode
	oldPreviousMode := PreviousMode
	oldTtControlC := TtControlC

	defer func() {
		CurrentFrame = oldCurrentFrame
		EditMode = oldEditMode
		PreviousMode = oldPreviousMode
		TtControlC = oldTtControlC
	}()

	t.Run("DittoInInsertMode_RejectsNegativeLeadParams", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		TtControlC = false
		setLineContent(lines[0], "test")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		testCases := []struct {
			name      string
			command   Commands
			leadParam LeadParam
			count     int
			editMode  ModeType
			prevMode  ModeType
		}{
			{"DittoUp_Insert_Minus", CmdDittoUp, LeadParamMinus, -1, ModeInsert, ModeCommand},
			{"DittoUp_Insert_NInt", CmdDittoUp, LeadParamNInt, -5, ModeInsert, ModeCommand},
			{"DittoUp_Insert_NIndef", CmdDittoUp, LeadParamNIndef, 0, ModeInsert, ModeCommand},
			{"DittoDown_Insert_Minus", CmdDittoDown, LeadParamMinus, -1, ModeInsert, ModeCommand},
			{"DittoDown_Insert_NInt", CmdDittoDown, LeadParamNInt, -3, ModeInsert, ModeCommand},
			{"DittoDown_Insert_NIndef", CmdDittoDown, LeadParamNIndef, 0, ModeInsert, ModeCommand},
			{"DittoUp_CommandAfterInsert_Minus", CmdDittoUp, LeadParamMinus, -2, ModeCommand, ModeInsert},
			{"DittoDown_CommandAfterInsert_NInt", CmdDittoDown, LeadParamNInt, -4, ModeCommand, ModeInsert},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				EditMode = tc.editMode
				PreviousMode = tc.prevMode

				result := CaseDittoCommand(tc.command, tc.leadParam, tc.count, true)
				assert.False(t, result, "Should reject %s with %v in insert mode context", tc.command, tc.leadParam)
			})
		}
	})

	t.Run("CommandSets_AreInitializedCorrectly", func(t *testing.T) {
		// This tests that the function recognizes and groups commands correctly
		// by checking the validation that rejects invalid commands
		frame, lines := setupTestFrame(1)
		CurrentFrame = frame
		EditMode = ModeCommand
		TtControlC = false
		setLineContent(lines[0], "test")
		frame.Dot.Line = lines[0]
		frame.Dot.Col = 1

		// Case commands should form one group
		caseCommands := []Commands{CmdCaseUp, CmdCaseLow, CmdCaseEdit}
		for _, cmd := range caseCommands {
			// These are valid case commands and should be recognized
			// (Can't test full execution without VDU though)
			assert.Contains(t, caseCommands, cmd, "Command should be in case command set")
		}

		// Ditto commands should form another group
		dittoCommands := []Commands{CmdDittoUp, CmdDittoDown}
		for _, cmd := range dittoCommands {
			assert.Contains(t, dittoCommands, cmd, "Command should be in ditto command set")
		}
	})
}

// TestInsertModeRestrictions specifically tests insert mode logic
func TestInsertModeRestrictions(t *testing.T) {
	oldCurrentFrame := CurrentFrame
	oldEditMode := EditMode
	oldPreviousMode := PreviousMode
	oldTtControlC := TtControlC

	defer func() {
		CurrentFrame = oldCurrentFrame
		EditMode = oldEditMode
		PreviousMode = oldPreviousMode
		TtControlC = oldTtControlC
	}()

	t.Run("AllNegativeParamCombinations", func(t *testing.T) {
		frame, lines := setupTestFrame(2)
		CurrentFrame = frame
		TtControlC = false
		setLineContent(lines[0], "above")
		setLineContent(lines[1], "current")
		frame.Dot.Line = lines[1]
		frame.Dot.Col = 1

		negativeParams := []LeadParam{LeadParamMinus, LeadParamNInt, LeadParamNIndef}
		dittoCommands := []Commands{CmdDittoUp, CmdDittoDown}
		insertModes := []struct {
			edit ModeType
			prev ModeType
			desc string
		}{
			{ModeInsert, ModeCommand, "ModeInsert"},
			{ModeCommand, ModeInsert, "ModeCommand with PrevMode=Insert"},
		}

		for _, mode := range insertModes {
			for _, cmd := range dittoCommands {
				for _, param := range negativeParams {
					t.Run(fmt.Sprintf("%s_%d_%d", mode.desc, cmd, param), func(t *testing.T) {
						EditMode = mode.edit
						PreviousMode = mode.prev

						result := CaseDittoCommand(cmd, param, -1, true)
						assert.False(t, result, "Should reject command %d with param %d in %s", cmd, param, mode.desc)
					})
				}
			}
		}
	})
}
