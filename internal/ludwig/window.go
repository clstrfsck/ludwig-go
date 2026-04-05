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
func WindowCommand(frame *FrameObject, command Commands, rept LeadParam, count int, fromSpan bool) bool {
	cmdSuccess := false

	switch command {
	case CmdWindowBackward:
		lineNr := LineToNumber(frame.Dot.Line)
		if lineNr <= frame.ScrHeight*count {
			MarkCreate(
				frame.FirstGroup.FirstLine, frame.Dot.Col, &frame.Dot,
			)
		} else {
			newLine := frame.Dot.Line
			for i := 1; i <= frame.ScrHeight*count; i++ {
				newLine = newLine.BLink
			}
			if count == 1 {
				line := frame.Dot.Line
				if line.ScrRowNr != 0 {
					if line.ScrRowNr > frame.ScrHeight-frame.MarginBottom {
						Screen.Scroll(
							-2*frame.ScrHeight+line.ScrRowNr+frame.MarginBottom,
							true,
						)
					} else {
						Screen.Scroll(-frame.ScrHeight, true)
					}
				}
			} else {
				Screen.Unload()
			}
			MarkCreate(newLine, frame.Dot.Col, &frame.Dot)
		}
		cmdSuccess = true

	case CmdWindowEnd:
		MarkCreate(frame.LastGroup.LastLine, frame.Dot.Col, &frame.Dot)
		cmdSuccess = true

	case CmdWindowForward:
		lineNr := LineToNumber(frame.Dot.Line)
		lastGroup := frame.LastGroup
		dot := frame.Dot
		if lineNr+frame.ScrHeight*count >
			lastGroup.FirstLineNr+lastGroup.LastLine.OffsetNr {
			MarkCreate(lastGroup.LastLine, dot.Col, &frame.Dot)
		} else {
			newLine := dot.Line
			for i := 1; i <= frame.ScrHeight*count; i++ {
				newLine = newLine.FLink
			}
			if count == 1 {
				line := dot.Line
				if line.ScrRowNr != 0 {
					if line.ScrRowNr <= frame.MarginTop {
						Screen.Scroll(
							frame.ScrHeight+line.ScrRowNr-frame.MarginTop-1,
							true,
						)
					} else {
						Screen.Scroll(frame.ScrHeight, true)
					}
				}
			} else {
				Screen.Unload()
			}
			MarkCreate(newLine, dot.Col, &frame.Dot)
		}
		cmdSuccess = true

	case CmdWindowLeft:
		cmdSuccess = true
		if Screen.Frame == frame {
			if rept == LeadParamNone {
				count = frame.ScrWidth / 2
			}
			if frame.ScrOffset < count {
				count = frame.ScrOffset
			}
			Screen.Slide(-count)
			if frame.ScrOffset+frame.ScrWidth < frame.Dot.Col {
				frame.Dot.Col = frame.ScrOffset + frame.ScrWidth
			}
		}

	case CmdWindowMiddle:
		cmdSuccess = true
		if Screen.Frame == frame {
			lineNr := LineToNumber(frame.Dot.Line)
			line2Nr := LineToNumber(Screen.TopLine)
			line3Nr := LineToNumber(Screen.BotLine)
			Screen.Scroll(lineNr-((line2Nr+line3Nr)/2), true)
		}

	case CmdWindowNew:
		cmdSuccess = true
		Screen.Redraw()

	case CmdWindowRight:
		cmdSuccess = true
		if Screen.Frame == frame {
			if rept == LeadParamNone {
				count = frame.ScrWidth / 2
			}
			if MaxStrLenP < (frame.ScrOffset+frame.ScrWidth)+count {
				count = MaxStrLenP - (frame.ScrOffset + frame.ScrWidth)
			}
			Screen.Slide(count)
			if frame.Dot.Col <= frame.ScrOffset {
				frame.Dot.Col = frame.ScrOffset + 1
			}
		}

	case CmdWindowScroll:
		cmdSuccess = true
		if frame == Screen.Frame {
			var key int
			for {
				switch rept {
				case LeadParamPIndef:
					count = max(frame.Dot.Line.ScrRowNr-1, 0)
				case LeadParamNIndef:
					count = frame.Dot.Line.ScrRowNr - frame.ScrHeight
				}
				if rept != LeadParamNone {
					Screen.Scroll(count, true)
				}
				key = 0

				// If the dot is still visible and the command is interactive
				// then support stay-behind mode
				if !fromSpan && (frame.Dot.Line.ScrRowNr != 0) &&
					(frame.ScrOffset < frame.Dot.Col) &&
					(frame.Dot.Col <= frame.ScrOffset+frame.ScrWidth) {
					if !cmdSuccess {
						VduBeep()
						cmdSuccess = true
					}
					VduMoveCurs(
						frame.Dot.Col-frame.ScrOffset,
						frame.Dot.Line.ScrRowNr,
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
		cmdSuccess = FrameSetHeight(frame, count, false)

	case CmdWindowTop:
		MarkCreate(frame.FirstGroup.FirstLine, frame.Dot.Col, &frame.Dot)
		cmdSuccess = true

	case CmdWindowUpdate:
		cmdSuccess = true
		if LudwigMode == LudwigScreen {
			Screen.Fixup(frame)
		}

	default:
		// All other commands ignored
	}

	return cmdSuccess
}
