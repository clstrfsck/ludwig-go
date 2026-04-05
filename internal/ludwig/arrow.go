/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         ARROW
//
// Description:  The arrow key, TAB, and BACKTAB commands.

package ludwig

func isArrowCommand(command Commands) bool {
	switch command {
	case CmdReturn,
		CmdHome,
		CmdTab,
		CmdBacktab,
		CmdLeft,
		CmdRight,
		CmdDown,
		CmdUp:
		return true
	}
	return false
}

// ArrowCommand handles arrow key, TAB, and BACKTAB commands
func ArrowCommand(frame *FrameObject, command Commands, rept LeadParam, count int, fromSpan bool) bool {
	cmdStatus := false
	var newEql MarkObject
	oldDot := *frame.Dot
	eopLineNr := frame.LastGroup.FirstLineNr + frame.LastGroup.LastLine.OffsetNr

	var key int
	for {
		cmdValid := false
		switch command {
		case CmdReturn:
			cmdValid = doCmdReturn(frame, count, &newEql, &eopLineNr)

		case CmdHome:
			cmdValid = doCmdHome(frame, &newEql)

		case CmdTab:
			cmdValid = doCmdTabBacktab(frame, 1, count, &newEql)

		case CmdBacktab:
			cmdValid = doCmdTabBacktab(frame, -1, count, &newEql)

		case CmdLeft:
			cmdValid = doCmdLeft(frame, rept, count, &newEql)

		case CmdRight:
			cmdValid = doCmdRight(frame, rept, count, &newEql)

		case CmdDown:
			cmdValid = doCmdDown(frame, rept, count, &newEql, eopLineNr)

		case CmdUp:
			cmdValid = doCmdUp(frame, rept, count, &newEql)
		}

		if cmdValid {
			cmdStatus = true
		}
		if fromSpan {
			break
		}
		Screen.Fixup(frame)
		if !cmdValid || ((command == CmdDown) && (rept != LeadParamPIndef) &&
			(frame.Dot.Line.FLink == nil)) {
			VduBeep()
		}
		key = VduGetKey()
		if TtControlC {
			break
		}
		rept = LeadParamNone
		count = 1
		command = Lookup[key].Command
		if (command == CmdReturn) && (EditMode == ModeInsert) {
			command = CmdSplitLine
		}
		if !isArrowCommand(command) {
			VduTakeBackKey(key)
			break
		}
	}

	if TtControlC {
		MarkCreate(oldDot.Line, oldDot.Col, &frame.Dot)
	} else {
		// Define Equals.
		if cmdStatus {
			MarkCreate(newEql.Line, newEql.Col, &frame.Marks[MarkEquals])
			if (command == CmdDown) && (rept != LeadParamPIndef) &&
				(frame.Dot.Line.FLink == nil) {
				cmdStatus = false
			}
		}
	}
	return cmdStatus || !fromSpan
}

func doCmdDown(frame *FrameObject, rept LeadParam, count int, newEql *MarkObject, eopLineNr int) bool {
	dotLine := frame.Dot.Line
	lineNr := LineToNumber(dotLine)
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		if lineNr+count <= eopLineNr {
			if count < MaxGroupLines/2 {
				for counter := 1; dotLine != nil && counter <= count; counter++ {
					dotLine = dotLine.FLink
				}
			} else {
				dotLine = LineFromNumber(frame, lineNr+count)
			}
		}
	case LeadParamPIndef:
		dotLine = frame.LastGroup.LastLine
	}
	if dotLine == nil {
		return false
	}
	*newEql = *frame.Dot
	MarkCreate(dotLine, frame.Dot.Col, &frame.Dot)
	return true
}

func doCmdHome(frame *FrameObject, newEql *MarkObject) bool {
	*newEql = *frame.Dot
	if frame == Screen.Frame {
		MarkCreate(Screen.TopLine, frame.ScrOffset+1, &frame.Dot)
	}
	return true
}

func doCmdLeft(frame *FrameObject, rept LeadParam, count int, newEql *MarkObject) bool {
	*newEql = *frame.Dot
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		if frame.Dot.Col-count >= 1 {
			frame.Dot.Col -= count
			return true
		}
	case LeadParamPIndef:
		if frame.Dot.Col >= frame.MarginLeft {
			frame.Dot.Col = frame.MarginLeft
			return true
		}
	}
	return false
}

func doCmdRight(frame *FrameObject, rept LeadParam, count int, newEql *MarkObject) bool {
	*newEql = *frame.Dot
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		if frame.Dot.Col+count <= MaxStrLenP {
			frame.Dot.Col += count
			return true
		}
	case LeadParamPIndef:
		if frame.Dot.Col <= frame.MarginRight {
			frame.Dot.Col = frame.MarginRight
			return true
		}
	}
	return false
}

func doCmdTabBacktab(frame *FrameObject, step, count int, newEql *MarkObject) bool {
	*newEql = *frame.Dot
	newCol := frame.Dot.Col
	for counter := 1; counter <= count; counter++ {
		for {
			newCol += step
			if newCol <= 0 || newCol >= MaxStrLenP ||
				frame.TabStops[newCol] ||
				(newCol == frame.MarginLeft) ||
				(newCol == frame.MarginRight) {
				break
			}
		}
		if (newCol <= 0) || (newCol >= MaxStrLenP) {
			return false
		}
	}
	frame.Dot.Col = newCol
	return true
}

func doCmdUp(frame *FrameObject, rept LeadParam, count int, newEql *MarkObject) bool {
	dotLine := frame.Dot.Line
	lineNr := LineToNumber(dotLine)
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		if lineNr-count > 0 {
			if count < MaxGroupLines/2 {
				for counter := 1; counter <= count; counter++ {
					dotLine = dotLine.BLink
				}
			} else {
				dotLine = LineFromNumber(frame, lineNr-count)
			}
		} else {
			return false
		}
	case LeadParamPIndef:
		dotLine = frame.FirstGroup.FirstLine
	}
	if dotLine == nil {
		return false
	}
	*newEql = *frame.Dot
	MarkCreate(dotLine, frame.Dot.Col, &frame.Dot)
	return true
}

func doCmdReturn(frame *FrameObject, count int, newEql *MarkObject, eopLineNr *int) bool {
	*newEql = *frame.Dot
	dotLine := frame.Dot.Line
	dotCol := frame.Dot.Col
	for counter := 1; counter <= count; counter++ {
		if TtControlC {
			return false
		}
		if dotLine.FLink == nil {
			TextRealizeNull(dotLine)
			(*eopLineNr)++
			dotLine = dotLine.BLink
			if counter == 1 {
				newEql.Line = dotLine
			}
		}
		dotCol = TextReturnCol(dotLine, dotCol, false)
		dotLine = dotLine.FLink
	}
	MarkCreate(dotLine, dotCol, &frame.Dot)
	return true
}
