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

// CharcmdInsert handles character insertion commands
func CharcmdInsert(frame *FrameObject, rept LeadParam, count int, fromSpan bool) (result bool) {
	cmdStatus := false
	if rept == LeadParamMinus {
		rept = LeadParamNInt
	}
	count = iabs(count)

	oldDotCol := frame.Dot.Col
	var maximum int
	if frame.Dot.Col <= frame.Dot.Line.Used {
		maximum = MaxStrLen - frame.Dot.Line.Used
	} else {
		maximum = MaxStrLen - frame.Dot.Col
	}

	inserted := 0
	var eqlCol int
	var key int

	defer func() {
		if TtControlC {
			cmdStatus = false
			frame.Dot.Col = oldDotCol
			var tempMark *MarkObject
			MarkCreate(frame.Dot.Line, frame.Dot.Col+inserted, &tempMark)
			TextRemove(frame.Dot, tempMark)
			MarkDestroy(&tempMark)
		} else if cmdStatus {
			frame.TextModified = true
			MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
			MarkCreate(frame.Dot.Line, eqlCol, &frame.Marks[MarkEquals])
		}
		result = cmdStatus || !fromSpan
	}()

	for {
		cmdValid := count <= maximum
		if cmdValid {
			maximum -= count
			inserted += count
			if !TextInsert(true, 1, BlankString, count, frame.Dot) {
				return
			}
			if rept == LeadParamNInt {
				eqlCol = frame.Dot.Col - count
			} else {
				eqlCol = frame.Dot.Col
				frame.Dot.Col -= count
			}
			cmdStatus = true
		}
		if fromSpan {
			return
		}
		if cmdValid {
			Screen.Fixup(frame)
		} else {
			VduBeep()
		}
		key = VduGetKey()
		if TtControlC {
			return
		}
		rept = LeadParamNone
		count = 1
		var cmd Commands
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
	return
}

func joinLines(frame *FrameObject) bool {
	// Only join lines if we are in the newline mode and there is a previous line to join to
	if !frame.Options.Has(OptNewLine) {
		return false
	}
	bLine := frame.Dot.Line.BLink
	if bLine == nil {
		return false
	}
	var theOtherMark *MarkObject
	MarkCreate(bLine, bLine.Used+1, &theOtherMark)
	defer MarkDestroy(&theOtherMark)
	if TextRemove(theOtherMark, frame.Dot) {
		frame.TextModified = true
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
		return true
	}
	return false
}

// CharcmdDelete handles character deletion commands
func CharcmdDelete(frame *FrameObject, rept LeadParam, count int, fromSpan bool) (result bool) {
	cmdStatus := false
	oldDotCol := frame.Dot.Col
	oldStr := NewStrObjectCopy(
		frame.Dot.Line.Str,
		1,
		frame.Dot.Line.Used,
		frame.Dot.Line.Used,
	)
	deleted := 0
	var key int

	defer func() {
		if TtControlC {
			cmdStatus = false
			frame.Dot.Col = 1
			TextOvertype(false, 1, oldStr, oldStr.Len(), frame.Dot)
			frame.Dot.Col = oldDotCol
		} else if cmdStatus {
			oldDotCol = frame.Dot.Col
			count = MaxStrLenP - oldDotCol
			if deleted > count {
				deleted = count
			}
			line := frame.Dot.Line
			MarksSqueeze(line, oldDotCol, line, oldDotCol+deleted)
			MarksShift(
				line,
				oldDotCol+deleted,
				MaxStrLenP-(oldDotCol+deleted)+1,
				line,
				oldDotCol,
			)
			frame.TextModified = true
			MarkCreate(line, frame.Dot.Col, &frame.Marks[MarkModified])
			if frame.Marks[MarkEquals] != nil {
				MarkDestroy(&frame.Marks[MarkEquals])
			}
		}
		result = cmdStatus || !fromSpan
	}()

	for {
		cmdValid := true
		dotCol := frame.Dot.Col
		switch rept {
		case LeadParamNone, LeadParamPlus, LeadParamPInt:
			if count > MaxStrLenP-dotCol {
				cmdValid = false
			}
		case LeadParamPIndef:
			count = MaxStrLenP - dotCol
		case LeadParamMinus, LeadParamNInt:
			count = -count
			if count < dotCol {
				frame.Dot.Col -= count
			} else if !fromSpan && count == 1 && dotCol == 1 && joinLines(frame) {
				MarkDestroy(&frame.Marks[MarkEquals])
				return
			} else {
				cmdValid = false
			}
		case LeadParamNIndef:
			count = frame.Dot.Col - 1
			frame.Dot.Col = 1
		}

		if cmdValid {
			// Update the text of the line
			oldUsed := frame.Dot.Line.Used
			length := (frame.Dot.Line.Used + 1) - (frame.Dot.Col + count)
			if length > 0 {
				l := frame.Dot.Line
				dotCol := frame.Dot.Col
				l.Str.Erase(count, dotCol)
				l.Str.FillN(' ', count, l.Used+1-count)
				l.Used -= count
			} else if frame.Dot.Col <= frame.Dot.Line.Used {
				d := frame.Dot
				d.Line.Str.FillN(' ', d.Line.Used+1-d.Col, d.Col)
				d.Line.Used = d.Line.Str.TrimmedLen(' ', d.Col)
			}

			// Update the screen
			scrCol := frame.Dot.Col - frame.ScrOffset
			if (frame.Dot.Line.ScrRowNr != 0) && (count != 0) &&
				(frame.Dot.Col <= oldUsed) && (scrCol <= frame.ScrWidth) {
				if scrCol <= 0 {
					scrCol = 1
				}
				VduMoveCurs(scrCol, frame.Dot.Line.ScrRowNr)
				length = frame.ScrWidth + 1 - scrCol
				if count < length {
					length = count
					VduDeleteChars(count)
				} else {
					VduClearEOL()
				}
				firstCol := frame.ScrOffset + frame.ScrWidth + 1 - length
				if firstCol <= frame.Dot.Line.Used {
					VduMoveCurs(
						frame.ScrWidth+1-length, frame.Dot.Line.ScrRowNr,
					)
					if length > frame.Dot.Line.Used+1-firstCol {
						length = frame.Dot.Line.Used + 1 - firstCol
					}
					VduDisplayStr(frame.Dot.Line.Str.Slice(firstCol, length), true)
				}
			}
			deleted += count
			cmdStatus = true
		}

		if fromSpan {
			return
		}
		if cmdValid {
			Screen.Fixup(frame)
		} else {
			VduBeep()
		}
		key = VduGetKey()
		if TtControlC {
			return
		}
		rept = LeadParamNone
		count = 1
		var cmd Commands
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
	return
}

// CharcmdRubout handles rubout commands
func CharcmdRubout(frame *FrameObject, rept LeadParam, count int, fromSpan bool) (result bool) {
	var cmdStatus bool
	if EditMode == ModeInsert {
		if rept == LeadParamPIndef {
			rept = LeadParamNIndef
		} else {
			rept = LeadParamNInt
		}
		cmdStatus = CharcmdDelete(frame, rept, -count, fromSpan)
		return cmdStatus || !fromSpan
	} else {
		oldDotCol := frame.Dot.Col
		dotUsed := frame.Dot.Line.Used
		oldStr := NewStrObjectCopy(frame.Dot.Line.Str, 1, dotUsed, dotUsed)
		var key int
		var eqlCol int

		defer func() {
			if TtControlC {
				cmdStatus = false
				frame.Dot.Col = 1
				TextOvertype(false, 1, oldStr, dotUsed, frame.Dot)
				frame.Dot.Col = oldDotCol
			} else if cmdStatus {
				frame.TextModified = true
				MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
				MarkCreate(frame.Dot.Line, eqlCol, &frame.Marks[MarkEquals])
			}
			result = cmdStatus || !fromSpan
		}()

		for {
			if rept == LeadParamPIndef {
				count = frame.Dot.Col - 1
			}
			cmdValid := (count <= frame.Dot.Col-1)
			if cmdValid {
				eqlCol = frame.Dot.Col
				frame.Dot.Col -= count
				if !TextOvertype(true, 1, BlankString, count, frame.Dot) {
					return
				}
				frame.Dot.Col -= count
				cmdStatus = true
			}
			if fromSpan {
				return
			}
			if cmdValid {
				Screen.Fixup(frame)
			} else {
				VduBeep()
			}
			key = VduGetKey()
			if TtControlC {
				return
			}
			rept = LeadParamNone
			count = 1
			var cmd Commands
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
		return
	}
}
