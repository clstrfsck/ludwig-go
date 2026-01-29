/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         WINDOW
//
// Description:  Implement the window commands.

package ludwig

// WindowCommand implements all window-related commands
func WindowCommand(command Commands, rept LeadParam, count int, fromSpan bool) bool {
	cmdSuccess := false

	var lineNr int
	var line2Nr int
	var line3Nr int

	switch command {
	case CmdWindowBackward:
		if !LineToNumber(CurrentFrame.Dot.Line, &lineNr) {
			goto cleanup
		}
		if lineNr <= CurrentFrame.ScrHeight*count {
			MarkCreate(
				CurrentFrame.FirstGroup.FirstLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot,
			)
		} else {
			newLine := CurrentFrame.Dot.Line
			for i := 1; i <= CurrentFrame.ScrHeight*count; i++ {
				newLine = newLine.BLink
			}
			if count == 1 {
				line := CurrentFrame.Dot.Line
				if line.ScrRowNr != 0 {
					if line.ScrRowNr > CurrentFrame.ScrHeight-CurrentFrame.MarginBottom {
						ScreenScroll(
							-2*CurrentFrame.ScrHeight+line.ScrRowNr+CurrentFrame.MarginBottom,
							true,
						)
					} else {
						ScreenScroll(-CurrentFrame.ScrHeight, true)
					}
				}
			} else {
				ScreenUnload()
			}
			MarkCreate(newLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot)
		}
		cmdSuccess = true

	case CmdWindowEnd:
		cmdSuccess = MarkCreate(
			CurrentFrame.LastGroup.LastLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot,
		)

	case CmdWindowForward:
		if !LineToNumber(CurrentFrame.Dot.Line, &lineNr) {
			goto cleanup
		}
		lastGroup := CurrentFrame.LastGroup
		dot := CurrentFrame.Dot
		if lineNr+CurrentFrame.ScrHeight*count >
			lastGroup.FirstLineNr+lastGroup.LastLine.OffsetNr {
			MarkCreate(lastGroup.LastLine, dot.Col, &dot)
		} else {
			newLine := dot.Line
			for i := 1; i <= CurrentFrame.ScrHeight*count; i++ {
				newLine = newLine.FLink
			}
			if count == 1 {
				line := dot.Line
				if line.ScrRowNr != 0 {
					if line.ScrRowNr <= CurrentFrame.MarginTop {
						ScreenScroll(
							CurrentFrame.ScrHeight+line.ScrRowNr-CurrentFrame.MarginTop-1,
							true,
						)
					} else {
						ScreenScroll(CurrentFrame.ScrHeight, true)
					}
				}
			} else {
				ScreenUnload()
			}
			MarkCreate(newLine, dot.Col, &CurrentFrame.Dot)
		}
		cmdSuccess = true

	case CmdWindowLeft:
		cmdSuccess = true
		if ScrFrame == CurrentFrame {
			if rept == LeadParamNone {
				count = CurrentFrame.ScrWidth / 2
			}
			if CurrentFrame.ScrOffset < count {
				count = CurrentFrame.ScrOffset
			}
			ScreenSlide(-count)
			if CurrentFrame.ScrOffset+CurrentFrame.ScrWidth < CurrentFrame.Dot.Col {
				CurrentFrame.Dot.Col = CurrentFrame.ScrOffset + CurrentFrame.ScrWidth
			}
		}

	case CmdWindowMiddle:
		cmdSuccess = true
		if ScrFrame == CurrentFrame {
			if LineToNumber(CurrentFrame.Dot.Line, &lineNr) &&
				LineToNumber(ScrTopLine, &line2Nr) && LineToNumber(ScrBotLine, &line3Nr) {
				ScreenScroll(lineNr-((line2Nr+line3Nr)/2), true)
			}
		}

	case CmdWindowNew:
		cmdSuccess = true
		ScreenRedraw()

	case CmdWindowRight:
		cmdSuccess = true
		if ScrFrame == CurrentFrame {
			if rept == LeadParamNone {
				count = CurrentFrame.ScrWidth / 2
			}
			if MaxStrLenP < (CurrentFrame.ScrOffset+CurrentFrame.ScrWidth)+count {
				count = MaxStrLenP - (CurrentFrame.ScrOffset + CurrentFrame.ScrWidth)
			}
			ScreenSlide(count)
			if CurrentFrame.Dot.Col <= CurrentFrame.ScrOffset {
				CurrentFrame.Dot.Col = CurrentFrame.ScrOffset + 1
			}
		}

	case CmdWindowScroll:
		cmdSuccess = true
		if CurrentFrame == ScrFrame {
			var key int
			for {
				if rept == LeadParamPIndef {
					count = CurrentFrame.Dot.Line.ScrRowNr - 1
					if count < 0 {
						count = 0
					}
				} else if rept == LeadParamNIndef {
					count = CurrentFrame.Dot.Line.ScrRowNr - CurrentFrame.ScrHeight
				}
				if rept != LeadParamNone {
					ScreenScroll(count, true)
				}
				key = 0

				// If the dot is still visible and the command is interactive
				// then support stay-behind mode
				if !fromSpan && (CurrentFrame.Dot.Line.ScrRowNr != 0) &&
					(CurrentFrame.ScrOffset < CurrentFrame.Dot.Col) &&
					(CurrentFrame.Dot.Col <= CurrentFrame.ScrOffset+CurrentFrame.ScrWidth) {
					if !cmdSuccess {
						VduBeep()
						cmdSuccess = true
					}
					VduMoveCurs(
						CurrentFrame.Dot.Col-CurrentFrame.ScrOffset,
						CurrentFrame.Dot.Line.ScrRowNr,
					)
					key = VduGetKey()
					if TtControlC {
						key = 0
					} else if Lookup[key].Command == CmdUp {
						rept = LeadParamPInt
						count = 1
					} else if Lookup[key].Command == CmdDown {
						rept = LeadParamNInt
						count = -1
					} else {
						VduTakeBackKey(key)
						key = 0
					}
				}
				if key == 0 {
					break
				}
			}
		}

	case CmdWindowSetHeight:
		if rept == LeadParamNone {
			count = TerminalInfo.Height
		}
		cmdSuccess = FrameSetHeight(count, false)

	case CmdWindowTop:
		cmdSuccess = MarkCreate(
			CurrentFrame.FirstGroup.FirstLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot,
		)

	case CmdWindowUpdate:
		cmdSuccess = true
		if LudwigMode == LudwigScreen {
			ScreenFixup()
		}

	default:
		// All other commands ignored
	}

cleanup:
	return cmdSuccess
}
