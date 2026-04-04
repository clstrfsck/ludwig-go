/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         CASEDITTO
//
// Description:  The Case change and Ditto commands.

package ludwig

type commandType int

const (
	caseCommandType commandType = iota
	dittoCommandType
	unknownCommandType
)

func getCommandType(command Commands) commandType {
	switch command {
	case CmdCaseUp, CmdCaseLow, CmdCaseEdit:
		return caseCommandType
	case CmdDittoUp, CmdDittoDown:
		return dittoCommandType
	default:
		return unknownCommandType
	}
}

// CaseDittoCommand handles case change and ditto commands
func CaseDittoCommand(frame *FrameObject, command Commands, rept LeadParam, count int, fromSpan bool) bool {
	cmdStatus := false
	insert := (command == CmdDittoUp || command == CmdDittoDown) &&
		((EditMode == ModeInsert) ||
			((EditMode == ModeCommand) && (PreviousMode == ModeInsert)))

	// Remember current line
	oldDotCol := frame.Dot.Col

	oldStr := NewStrObjectCopy(
		frame.Dot.Line.Str,
		1,
		frame.Dot.Line.Used,
		frame.Dot.Line.Used,
	)

	cmdType := getCommandType(command)
	var otherLine *LineHdrObject
	switch command {
	case CmdCaseUp, CmdCaseLow, CmdCaseEdit:
		otherLine = frame.Dot.Line
	case CmdDittoUp, CmdDittoDown:
		if insert && (rept == LeadParamMinus || rept == LeadParamNInt ||
			rept == LeadParamNIndef) {
			ScreenMessage(&Screen, MsgNotAllowedInInsertMode)
			return false
		}
	}

	for {
		switch command {
		case CmdDittoUp:
			otherLine = frame.Dot.Line.BLink
		case CmdDittoDown:
			otherLine = frame.Dot.Line.FLink
		}

		cmdValid := true
		var firstCol int
		var newCol int
		if otherLine != nil {
			switch rept {
			case LeadParamNone, LeadParamPlus, LeadParamPInt:
				if (count != 0) && (frame.Dot.Col+count > otherLine.Used+1) {
					cmdValid = false
				}
				firstCol = frame.Dot.Col
				newCol = frame.Dot.Col + count

			case LeadParamPIndef:
				count = otherLine.Used + 1 - frame.Dot.Col
				if count < 0 {
					cmdValid = false
				}
				firstCol = frame.Dot.Col
				newCol = otherLine.Used + 1

			case LeadParamMinus, LeadParamNInt:
				count = -count
				if count >= frame.Dot.Col {
					cmdValid = false
				} else {
					firstCol = frame.Dot.Col - count
				}
				newCol = firstCol

			case LeadParamNIndef:
				count = frame.Dot.Col - 1
				firstCol = 1
				newCol = 1
			}
		} else {
			cmdValid = false
		}

		// Carry out the command
		if cmdValid {
			i := otherLine.Used + 1 - firstCol
			var newStr *StrObject
			if i > 0 {
				newStr = NewStrObjectCopy(otherLine.Str, firstCol, i, count)
			} else {
				newStr = NewBlankStrObject(count)
			}

			switch command {
			case CmdCaseUp:
				newStr.ApplyN(ChToUpper, count, 1)

			case CmdCaseLow:
				newStr.ApplyN(ChToLower, count, 1)

			case CmdCaseEdit:
				var ch byte
				if (1 < firstCol) && (firstCol <= otherLine.Used) {
					ch = otherLine.Str.Get(firstCol - 1)
				} else {
					ch = ' '
				}
				for j := 1; j <= count; j++ {
					if ChIsLetter(rune(ch)) {
						ch = ChToLower(newStr.Get(j))
					} else {
						ch = ChToUpper(newStr.Get(j))
					}
					newStr.Set(j, ch)
				}

			case CmdDittoUp, CmdDittoDown:
				// No massaging required
			}

			frame.Dot.Col = firstCol
			if insert {
				if !TextInsert(true, 1, newStr, count, frame.Dot) {
					break
				}
			} else {
				if !TextOvertype(true, 1, newStr, count, frame.Dot) {
					break
				}
			}
			// Reposition dot
			frame.Dot.Col = newCol
			cmdStatus = true
		}

		if fromSpan {
			break
		}
		if cmdValid {
			ScreenFixup(&Screen, frame)
		} else {
			VduBeep()
		}
		key := VduGetKey()
		if TtControlC {
			break
		}

		keyUp := ChKeyToUpper(key)

		switch rept {
		case LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef:
			rept = LeadParamPlus
			count = +1
		case LeadParamMinus, LeadParamNInt, LeadParamNIndef:
			rept = LeadParamMinus
			count = -1
		}

		if command == CmdDittoUp || command == CmdDittoDown {
			command = Lookup[keyUp].Command
		} else if keyUp == 'E' {
			command = CmdCaseEdit
		} else if keyUp == 'L' {
			command = CmdCaseLow
		} else if keyUp == 'U' {
			command = CmdCaseUp
		} else {
			command = CmdNoop
		}

		if cmdType == unknownCommandType || cmdType != getCommandType(command) {
			VduTakeBackKey(key)
			break
		}
	}

	if TtControlC {
		cmdStatus = false
		frame.Dot.Col = 1
		TextOvertype(false, 1, oldStr, oldStr.Len(), frame.Dot)
		frame.Dot.Col = oldDotCol
	} else if cmdStatus {
		frame.TextModified = true
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
		MarkCreate(
			frame.Dot.Line,
			oldDotCol,
			&frame.Marks[MarkEquals],
		)
	}
	return cmdStatus || !fromSpan
}
