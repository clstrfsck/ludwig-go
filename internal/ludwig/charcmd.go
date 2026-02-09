/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         CHARCMD
//
// Description:  Character Insert/Delete/Rubout commands.

package ludwig

import (
	"math"
)

// CharcmdInsert handles character insertion commands
func CharcmdInsert(cmd Commands, rept LeadParam, count int, fromSpan bool) bool {
	cmdStatus := false
	if rept == LeadParamMinus {
		rept = LeadParamNInt
	}
	count = int(math.Abs(float64(count)))

	oldDotCol := CurrentFrame.Dot.Col
	var maximum int
	if CurrentFrame.Dot.Col <= CurrentFrame.Dot.Line.Used {
		maximum = MaxStrLen - CurrentFrame.Dot.Line.Used
	} else {
		maximum = MaxStrLen - CurrentFrame.Dot.Col
	}

	inserted := 0
	var eqlCol int
	var key int

	for {
		cmdValid := count <= maximum
		if cmdValid {
			maximum -= count
			inserted += count
			if !TextInsert(true, 1, BlankString, count, CurrentFrame.Dot) {
				goto l9
			}
			if rept == LeadParamNInt {
				eqlCol = CurrentFrame.Dot.Col - count
			} else {
				eqlCol = CurrentFrame.Dot.Col
				CurrentFrame.Dot.Col -= count
			}
			cmdStatus = true
		}
		if fromSpan {
			goto l9
		}
		if cmdValid {
			ScreenFixup()
		} else {
			VduBeep()
		}
		key = VduGetKey()
		if TtControlC {
			goto l9
		}
		rept = LeadParamNone
		count = 1
		if ChIsPrintable(rune(key)) {
			cmd = CmdNoop
		} else {
			cmd = Lookup[key].Command
		}
		if cmd != CmdInsertChar {
			break
		}
	}
	VduTakeBackKey(key)

l9:
	if TtControlC {
		cmdStatus = false
		CurrentFrame.Dot.Col = oldDotCol
		var tempMark *MarkObject
		MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col+inserted, &tempMark)
		TextRemove(CurrentFrame.Dot, tempMark)
		MarkDestroy(&tempMark)
	} else {
		if cmdStatus {
			CurrentFrame.TextModified = true
			MarkCreate(
				CurrentFrame.Dot.Line,
				CurrentFrame.Dot.Col,
				&CurrentFrame.Marks[MarkModified-MinMarkNumber],
			)
			MarkCreate(CurrentFrame.Dot.Line, eqlCol, &CurrentFrame.Marks[MarkEquals-MinMarkNumber])
		}
	}
	return cmdStatus || !fromSpan
}

// CharcmdDelete handles character deletion commands
func CharcmdDelete(cmd Commands, rept LeadParam, count int, fromSpan bool) bool {
	cmdStatus := false
	oldDotCol := CurrentFrame.Dot.Col
	oldStr := NewStrObjectCopy(
		CurrentFrame.Dot.Line.Str,
		1,
		CurrentFrame.Dot.Line.Used,
		CurrentFrame.Dot.Line.Used,
	)
	deleted := 0
	var key int

	for {
		cmdValid := true
		switch rept {
		case LeadParamNone, LeadParamPlus, LeadParamPInt:
			if count > MaxStrLenP-CurrentFrame.Dot.Col {
				cmdValid = false
			}
		case LeadParamPIndef:
			count = MaxStrLenP - CurrentFrame.Dot.Col
		case LeadParamMinus, LeadParamNInt:
			count = -count
			if count < CurrentFrame.Dot.Col {
				CurrentFrame.Dot.Col -= count
			} else {
				cmdValid = false
			}
		case LeadParamNIndef:
			count = CurrentFrame.Dot.Col - 1
			CurrentFrame.Dot.Col = 1
		}

		if cmdValid {
			// Update the text of the line
			oldUsed := CurrentFrame.Dot.Line.Used
			length := (CurrentFrame.Dot.Line.Used + 1) - (CurrentFrame.Dot.Col + count)
			if length > 0 {
				l := CurrentFrame.Dot.Line
				dotCol := CurrentFrame.Dot.Col
				l.Str.Erase(count, dotCol)
				l.Str.FillN(' ', count, l.Used+1-count)
				l.Used -= count
			} else if CurrentFrame.Dot.Col <= CurrentFrame.Dot.Line.Used {
				d := CurrentFrame.Dot
				d.Line.Str.FillN(' ', d.Line.Used+1-d.Col, d.Col)
				d.Line.Used = d.Line.Str.Length(' ', d.Col)
			}

			// Update the screen
			scrCol := CurrentFrame.Dot.Col - CurrentFrame.ScrOffset
			if (CurrentFrame.Dot.Line.ScrRowNr != 0) && (count != 0) &&
				(CurrentFrame.Dot.Col <= oldUsed) && (scrCol <= CurrentFrame.ScrWidth) {
				if scrCol <= 0 {
					scrCol = 1
				}
				VduMoveCurs(scrCol, CurrentFrame.Dot.Line.ScrRowNr)
				length = CurrentFrame.ScrWidth + 1 - scrCol
				if count < length {
					length = count
					VduDeleteChars(count)
				} else {
					VduClearEOL()
				}
				firstCol := CurrentFrame.ScrOffset + CurrentFrame.ScrWidth + 1 - length
				if firstCol <= CurrentFrame.Dot.Line.Used {
					VduMoveCurs(
						CurrentFrame.ScrWidth+1-length, CurrentFrame.Dot.Line.ScrRowNr,
					)
					if length > CurrentFrame.Dot.Line.Used+1-firstCol {
						length = CurrentFrame.Dot.Line.Used + 1 - firstCol
					}
					VduDisplayStr(CurrentFrame.Dot.Line.Str.Slice(firstCol, length), 3)
				}
			}
			deleted += count
			cmdStatus = true
		}

		if fromSpan {
			goto l9
		}
		if cmdValid {
			ScreenFixup()
		} else {
			VduBeep()
		}
		key = VduGetKey()
		if TtControlC {
			goto l9
		}
		rept = LeadParamNone
		count = 1
		if ChIsPrintable(rune(key)) {
			cmd = CmdNoop
		} else {
			cmd = Lookup[key].Command
		}
		if (cmd == CmdRubout) && (EditMode == ModeInsert) {
			// In insert_mode treat RUBOUT as \-D
			rept = LeadParamMinus
			count = -1
			cmd = CmdDeleteChar
		}
		if cmd != CmdDeleteChar {
			break
		}
	}
	VduTakeBackKey(key)

l9:
	if TtControlC {
		cmdStatus = false
		CurrentFrame.Dot.Col = 1
		TextOvertype(false, 1, oldStr, MaxStrLen, CurrentFrame.Dot)
		CurrentFrame.Dot.Col = oldDotCol
	} else if cmdStatus {
		oldDotCol = CurrentFrame.Dot.Col
		count = MaxStrLenP - oldDotCol
		if deleted > count {
			deleted = count
		}
		line := CurrentFrame.Dot.Line
		MarksSqueeze(line, oldDotCol, line, oldDotCol+deleted)
		MarksShift(
			line,
			oldDotCol+deleted,
			MaxStrLenP-(oldDotCol+deleted)+1,
			line,
			oldDotCol,
		)
		CurrentFrame.TextModified = true
		MarkCreate(line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified-MinMarkNumber])
		if CurrentFrame.Marks[MarkEquals-MinMarkNumber] != nil {
			MarkDestroy(&CurrentFrame.Marks[MarkEquals-MinMarkNumber])
		}
	}
	return cmdStatus || !fromSpan
}

// CharcmdRubout handles rubout commands
func CharcmdRubout(cmd Commands, rept LeadParam, count int, fromSpan bool) bool {
	var cmdStatus bool
	if EditMode == ModeInsert {
		if rept == LeadParamPIndef {
			rept = LeadParamNIndef
		} else {
			rept = LeadParamNInt
		}
		cmdStatus = CharcmdDelete(CmdDeleteChar, rept, -count, fromSpan)
	} else {
		cmdStatus = false
		oldDotCol := CurrentFrame.Dot.Col
		dotUsed := CurrentFrame.Dot.Line.Used
		oldStr := NewStrObjectCopy(CurrentFrame.Dot.Line.Str, 1, dotUsed, dotUsed)
		var key int
		var eqlCol int

		for {
			if rept == LeadParamPIndef {
				count = CurrentFrame.Dot.Col - 1
			}
			cmdValid := (count <= CurrentFrame.Dot.Col-1)
			if cmdValid {
				eqlCol = CurrentFrame.Dot.Col
				CurrentFrame.Dot.Col -= count
				if !TextOvertype(true, 1, BlankString, count, CurrentFrame.Dot) {
					goto l9
				}
				CurrentFrame.Dot.Col -= count
				cmdStatus = true
			}
			if fromSpan {
				goto l9
			}
			if cmdValid {
				ScreenFixup()
			} else {
				VduBeep()
			}
			key = VduGetKey()
			if TtControlC {
				goto l9
			}
			rept = LeadParamNone
			count = 1
			if ChIsPrintable(rune(key)) {
				cmd = CmdNoop
			} else {
				cmd = Lookup[key].Command
			}
			if cmd != CmdRubout {
				break
			}
		}
		VduTakeBackKey(key)

	l9:
		if TtControlC {
			cmdStatus = false
			CurrentFrame.Dot.Col = 1
			TextOvertype(false, 1, oldStr, dotUsed, CurrentFrame.Dot)
			CurrentFrame.Dot.Col = oldDotCol
		} else if cmdStatus {
			CurrentFrame.TextModified = true
			MarkCreate(
				CurrentFrame.Dot.Line,
				CurrentFrame.Dot.Col,
				&CurrentFrame.Marks[MarkModified-MinMarkNumber],
			)
			MarkCreate(CurrentFrame.Dot.Line, eqlCol, &CurrentFrame.Marks[MarkEquals-MinMarkNumber])
		}
	}
	return cmdStatus || !fromSpan
}
