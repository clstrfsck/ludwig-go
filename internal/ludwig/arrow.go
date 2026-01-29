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

var arrowCommands = map[Commands]bool{
	CmdReturn:  true,
	CmdHome:    true,
	CmdTab:     true,
	CmdBacktab: true,
	CmdLeft:    true,
	CmdRight:   true,
	CmdDown:    true,
	CmdUp:      true,
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
		if arrowCommands[command] {
			switch command {
			case CmdReturn:
				{
					cmdValid = true
					newEql = *CurrentFrame.Dot
					dotLine := CurrentFrame.Dot.Line
					dotCol := CurrentFrame.Dot.Col
					for counter := 1; counter <= count; counter++ {
						if TtControlC {
							goto l9
						}
						if dotLine.FLink == nil {
							if !TextRealizeNull(dotLine) {
								goto l9
							}
							eopLineNr++
							dotLine = dotLine.BLink
							if counter == 1 {
								newEql.Line = dotLine
							}
						}
						dotCol = TextReturnCol(dotLine, dotCol, false)
						dotLine = dotLine.FLink
					}
					if !MarkCreate(dotLine, dotCol, &CurrentFrame.Dot) {
						goto l9
					}
				}

			case CmdHome:
				{
					cmdValid = true
					newEql = *CurrentFrame.Dot
					if CurrentFrame == ScrFrame {
						if !MarkCreate(
							ScrTopLine, CurrentFrame.ScrOffset+1, &CurrentFrame.Dot,
						) {
							goto l9
						}
					}
				}

			case CmdTab, CmdBacktab:
				{
					newCol := CurrentFrame.Dot.Col
					var step int
					if command == CmdTab {
						if newCol == MaxStrLenP {
							goto l1
						}
						step = 1
					} else {
						step = -1
					}
					for counter := 1; counter <= count; counter++ {
						for {
							newCol += step
							if CurrentFrame.TabStops[newCol] ||
								(newCol == CurrentFrame.MarginLeft) ||
								(newCol == CurrentFrame.MarginRight) {
								break
							}
						}
						if (newCol == 0) || (newCol == MaxStrLenP) {
							goto l1
						}
					}
					cmdValid = true
					newEql = *CurrentFrame.Dot
					CurrentFrame.Dot.Col = newCol
				l1:
				}

			case CmdLeft:
				{
					switch rept {
					case LeadParamNone, LeadParamPlus, LeadParamPInt:
						if CurrentFrame.Dot.Col-count >= 1 {
							cmdValid = true
							newEql = *CurrentFrame.Dot
							CurrentFrame.Dot.Col -= count
						}
					case LeadParamPIndef:
						if CurrentFrame.Dot.Col >= CurrentFrame.MarginLeft {
							cmdValid = true
							newEql = *CurrentFrame.Dot
							CurrentFrame.Dot.Col = CurrentFrame.MarginLeft
						}
					}
				}

			case CmdRight:
				{
					switch rept {
					case LeadParamNone, LeadParamPlus, LeadParamPInt:
						if CurrentFrame.Dot.Col+count <= MaxStrLenP {
							cmdValid = true
							newEql = *CurrentFrame.Dot
							CurrentFrame.Dot.Col += count
						}
					case LeadParamPIndef:
						if CurrentFrame.Dot.Col <= CurrentFrame.MarginRight {
							cmdValid = true
							newEql = *CurrentFrame.Dot
							CurrentFrame.Dot.Col = CurrentFrame.MarginRight
						}
					}
				}

			case CmdDown:
				{
					dotLine := CurrentFrame.Dot.Line
					var lineNr int
					if !LineToNumber(dotLine, &lineNr) {
						goto l9
					}
					switch rept {
					case LeadParamNone, LeadParamPlus, LeadParamPInt:
						if lineNr+count <= eopLineNr {
							cmdValid = true
							newEql = *CurrentFrame.Dot
							if count < MaxGroupLines/2 {
								for counter := 1; counter <= count; counter++ {
									dotLine = dotLine.FLink
								}
							} else {
								if !LineFromNumber(CurrentFrame, lineNr+count, &dotLine) {
									goto l9
								}
							}
						}
					case LeadParamPIndef:
						{
							cmdValid = true
							newEql = *CurrentFrame.Dot
							dotLine = CurrentFrame.LastGroup.LastLine
						}
					}
					if !MarkCreate(dotLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot) {
						goto l9
					}
				}

			case CmdUp:
				{
					dotLine := CurrentFrame.Dot.Line
					var lineNr int
					if !LineToNumber(dotLine, &lineNr) {
						goto l9
					}
					switch rept {
					case LeadParamNone, LeadParamPlus, LeadParamPInt:
						if lineNr-count > 0 {
							cmdValid = true
							newEql = *CurrentFrame.Dot
							if count < MaxGroupLines/2 {
								for counter := 1; counter <= count; counter++ {
									dotLine = dotLine.BLink
								}
							} else {
								if !LineFromNumber(CurrentFrame, lineNr-count, &dotLine) {
									goto l9
								}
							}
						}
					case LeadParamPIndef:
						{
							cmdValid = true
							newEql = *CurrentFrame.Dot
							dotLine = CurrentFrame.FirstGroup.FirstLine
						}
					}
					if !MarkCreate(dotLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot) {
						goto l9
					}
				}
			}
		} else {
			VduTakeBackKey(key)
			goto l9
		}

		if cmdValid {
			cmdStatus = true
		}
		if fromSpan {
			goto l9
		}
		ScreenFixup()
		if !cmdValid || ((command == CmdDown) && (rept != LeadParamPIndef) &&
			(CurrentFrame.Dot.Line.FLink == nil)) {
			VduBeep()
		}
		key = VduGetKey()
		if TtControlC {
			goto l9
		}
		rept = LeadParamNone
		count = 1
		command = Lookup[key].Command
		if (command == CmdReturn) && (EditMode == ModeInsert) {
			command = CmdSplitLine
		}
	}

l9:
	if TtControlC {
		MarkCreate(oldDot.Line, oldDot.Col, &CurrentFrame.Dot)
	} else {
		// Define Equals.
		if cmdStatus {
			MarkCreate(newEql.Line, newEql.Col, &CurrentFrame.Marks[MarkEquals-MinMarkNumber])
			if (command == CmdDown) && (rept != LeadParamPIndef) &&
				(CurrentFrame.Dot.Line.FLink == nil) {
				cmdStatus = false
			}
		}
	}
	return cmdStatus || !fromSpan
}
