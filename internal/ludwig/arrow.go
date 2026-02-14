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
func ArrowCommand(command Commands, rept LeadParam, count int, fromSpan bool) bool {
	cmdStatus := false
	var newEql MarkObject
	oldDot := *CurrentFrame.Dot
	eopLineNr := CurrentFrame.LastGroup.FirstLineNr + CurrentFrame.LastGroup.LastLine.OffsetNr

	var key int
	for {
		cmdValid := false
		switch command {
		case CmdReturn:
			cmdValid = doCmdReturn(count, &newEql, &eopLineNr)

		case CmdHome:
			cmdValid = doCmdHome(&newEql)

		case CmdTab:
			cmdValid = doCmdTabBacktab(1, count, &newEql)

		case CmdBacktab:
			cmdValid = doCmdTabBacktab(-1, count, &newEql)

		case CmdLeft:
			cmdValid = doCmdLeft(rept, count, &newEql)

		case CmdRight:
			cmdValid = doCmdRight(rept, count, &newEql)

		case CmdDown:
			cmdValid = doCmdDown(rept, count, &newEql, eopLineNr)

		case CmdUp:
			cmdValid = doCmdUp(rept, count, &newEql)
		}

		if cmdValid {
			cmdStatus = true
		}
		if fromSpan {
			break
		}
		ScreenFixup()
		if !cmdValid || ((command == CmdDown) && (rept != LeadParamPIndef) &&
			(CurrentFrame.Dot.Line.FLink == nil)) {
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
		MarkCreate(oldDot.Line, oldDot.Col, &CurrentFrame.Dot)
	} else {
		// Define Equals.
		if cmdStatus {
			MarkCreate(newEql.Line, newEql.Col, &CurrentFrame.Marks[MarkEquals])
			if (command == CmdDown) && (rept != LeadParamPIndef) &&
				(CurrentFrame.Dot.Line.FLink == nil) {
				cmdStatus = false
			}
		}
	}
	return cmdStatus || !fromSpan
}

func doCmdDown(rept LeadParam, count int, newEql *MarkObject, eopLineNr int) bool {
	*newEql = *CurrentFrame.Dot
	dotLine := CurrentFrame.Dot.Line
	var lineNr int
	if !LineToNumber(dotLine, &lineNr) {
		return false
	}
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		if lineNr+count <= eopLineNr {
			if count < MaxGroupLines/2 {
				for counter := 1; counter <= count; counter++ {
					dotLine = dotLine.FLink
				}
			} else {
				if !LineFromNumber(CurrentFrame, lineNr+count, &dotLine) {
					return false
				}
			}
		}
	case LeadParamPIndef:
		dotLine = CurrentFrame.LastGroup.LastLine
	}
	if !MarkCreate(dotLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot) {
		return false
	}
	return true
}

func doCmdHome(newEql *MarkObject) bool {
	*newEql = *CurrentFrame.Dot
	if CurrentFrame == ScrFrame {
		if !MarkCreate(ScrTopLine, CurrentFrame.ScrOffset+1, &CurrentFrame.Dot) {
			return false
		}
	}
	return true
}

func doCmdLeft(rept LeadParam, count int, newEql *MarkObject) bool {
	*newEql = *CurrentFrame.Dot
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		if CurrentFrame.Dot.Col-count >= 1 {
			CurrentFrame.Dot.Col -= count
			return true
		}
	case LeadParamPIndef:
		if CurrentFrame.Dot.Col >= CurrentFrame.MarginLeft {
			CurrentFrame.Dot.Col = CurrentFrame.MarginLeft
			return true
		}
	}
	return false
}

func doCmdRight(rept LeadParam, count int, newEql *MarkObject) bool {
	*newEql = *CurrentFrame.Dot
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		if CurrentFrame.Dot.Col+count <= MaxStrLenP {
			CurrentFrame.Dot.Col += count
			return true
		}
	case LeadParamPIndef:
		if CurrentFrame.Dot.Col <= CurrentFrame.MarginRight {
			CurrentFrame.Dot.Col = CurrentFrame.MarginRight
			return true
		}
	}
	return false
}

func doCmdTabBacktab(step, count int, newEql *MarkObject) bool {
	*newEql = *CurrentFrame.Dot
	newCol := CurrentFrame.Dot.Col
	for counter := 1; counter <= count; counter++ {
		for {
			newCol += step
			if newCol <= 0 || newCol >= MaxStrLenP ||
				CurrentFrame.TabStops[newCol] ||
				(newCol == CurrentFrame.MarginLeft) ||
				(newCol == CurrentFrame.MarginRight) {
				break
			}
		}
		if (newCol <= 0) || (newCol >= MaxStrLenP) {
			return false
		}
	}
	CurrentFrame.Dot.Col = newCol
	return true
}

func doCmdUp(rept LeadParam, count int, newEql *MarkObject) bool {
	*newEql = *CurrentFrame.Dot
	dotLine := CurrentFrame.Dot.Line
	var lineNr int
	if !LineToNumber(dotLine, &lineNr) {
		return false
	}
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		if lineNr-count > 0 {
			if count < MaxGroupLines/2 {
				for counter := 1; counter <= count; counter++ {
					dotLine = dotLine.BLink
				}
			} else {
				if !LineFromNumber(CurrentFrame, lineNr-count, &dotLine) {
					return false
				}
			}
		} else {
			return false
		}
	case LeadParamPIndef:
		dotLine = CurrentFrame.FirstGroup.FirstLine
	}
	if !MarkCreate(dotLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot) {
		return false
	}
	return true
}

func doCmdReturn(count int, newEql *MarkObject, eopLineNr *int) bool {
	*newEql = *CurrentFrame.Dot
	dotLine := CurrentFrame.Dot.Line
	dotCol := CurrentFrame.Dot.Col
	for counter := 1; counter <= count; counter++ {
		if TtControlC {
			return false
		}
		if dotLine.FLink == nil {
			if !TextRealizeNull(dotLine) {
				return false
			}
			*eopLineNr++
			dotLine = dotLine.BLink
			if counter == 1 {
				newEql.Line = dotLine
			}
		}
		dotCol = TextReturnCol(dotLine, dotCol, false)
		dotLine = dotLine.FLink
	}
	if !MarkCreate(dotLine, dotCol, &CurrentFrame.Dot) {
		return false
	}
	return true
}
