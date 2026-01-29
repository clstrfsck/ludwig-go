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

const lettersS = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var letters map[byte]bool

func init() {
	letters = make(map[byte]bool)
	for i := range lettersS {
		letters[lettersS[i]] = true
	}
}

func keyIsLower(key int) bool {
	if key < 0 {
		return false
	}
	if key > MaxSetRange {
		return false
	}
	return LowerSet.Bit(key) != 0
}

// CaseDittoCommand handles case change and ditto commands
func CaseDittoCommand(command Commands, rept LeadParam, count int, fromSpan bool) bool {
	cmdStatus := false
	insert := (command == CmdDittoUp || command == CmdDittoDown) &&
		((EditMode == ModeInsert) ||
			((EditMode == ModeCommand) && (PreviousMode == ModeInsert)))

	// Remember current line
	oldDotCol := CurrentFrame.Dot.Col

	var oldStr StrObject
	ChFillCopy(
		CurrentFrame.Dot.Line.Str,
		1,
		CurrentFrame.Dot.Line.Used,
		&oldStr,
		1,
		MaxStrLen,
		' ',
	)

	commandSet := make(map[Commands]bool)
	var otherLine *LineHdrObject

	switch command {
	case CmdCaseUp, CmdCaseLow, CmdCaseEdit:
		commandSet[CmdCaseUp] = true
		commandSet[CmdCaseLow] = true
		commandSet[CmdCaseEdit] = true
		otherLine = CurrentFrame.Dot.Line
	case CmdDittoUp, CmdDittoDown:
		if insert && (rept == LeadParamMinus || rept == LeadParamNInt ||
			rept == LeadParamNIndef) {
			ScreenMessage(MsgNotAllowedInInsertMode)
			return false
		}
		commandSet[CmdDittoUp] = true
		commandSet[CmdDittoDown] = true
	}

	var firstCol int
	var newCol int
	var key int

	for {
		switch command {
		case CmdDittoUp:
			otherLine = CurrentFrame.Dot.Line.BLink
		case CmdDittoDown:
			otherLine = CurrentFrame.Dot.Line.FLink
		}

		cmdValid := true
		if otherLine != nil {
			switch rept {
			case LeadParamNone, LeadParamPlus, LeadParamPInt:
				if (count != 0) && (CurrentFrame.Dot.Col+count > otherLine.Used+1) {
					cmdValid = false
				}
				firstCol = CurrentFrame.Dot.Col
				newCol = CurrentFrame.Dot.Col + count

			case LeadParamPIndef:
				count = otherLine.Used + 1 - CurrentFrame.Dot.Col
				if count < 0 {
					cmdValid = false
				}
				firstCol = CurrentFrame.Dot.Col
				newCol = otherLine.Used + 1

			case LeadParamMinus, LeadParamNInt:
				count = -count
				if count >= CurrentFrame.Dot.Col {
					cmdValid = false
				} else {
					firstCol = CurrentFrame.Dot.Col - count
				}
				newCol = firstCol

			case LeadParamNIndef:
				count = CurrentFrame.Dot.Col - 1
				firstCol = 1
				newCol = 1
			}
		} else {
			cmdValid = false
		}

		// Carry out the command
		if cmdValid {
			i := otherLine.Used + 1 - firstCol
			var newStr StrObject
			if i <= 0 {
				newStr.FillN(' ', count, 1)
			} else {
				ChFillCopy(otherLine.Str, firstCol, i, &newStr, 1, count, ' ')
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
					if letters[ch] {
						ch = ChToLower(newStr.Get(j))
					} else {
						ch = ChToUpper(newStr.Get(j))
					}
					newStr.Set(j, ch)
				}

			case CmdDittoUp, CmdDittoDown:
				// No massaging required
			}

			CurrentFrame.Dot.Col = firstCol
			if insert {
				if !TextInsert(true, 1, newStr, count, CurrentFrame.Dot) {
					goto l9
				}
			} else {
				if !TextOvertype(true, 1, newStr, count, CurrentFrame.Dot) {
					goto l9
				}
			}
			// Reposition dot
			CurrentFrame.Dot.Col = newCol
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

		var keyUp int
		if keyIsLower(key) {
			keyUp = key - 32 // Uppercase it!
		} else {
			keyUp = key
		}

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

		if !commandSet[command] {
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
		CurrentFrame.TextModified = true
		MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkModified-MinMarkNumber],
		)
		MarkCreate(
			CurrentFrame.Dot.Line,
			oldDotCol,
			&CurrentFrame.Marks[MarkEquals-MinMarkNumber],
		)
	}
	return cmdStatus || !fromSpan
}
