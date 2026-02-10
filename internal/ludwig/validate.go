/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         VALIDATE
//
// Description:  Validation of entire Ludwig data structure.

package ludwig

// ValidateCommand validates the data structure
func ValidateCommand() bool {
	/*
	  Purpose  : Validate the data structure.
	  Inputs   : none.
	  Outputs  : none.
	  Bugchecks: .lots and lots of them!
	*/
	// In production builds, validation is typically disabled
	// For now, we always validate. Can be controlled via build tags if needed.

	const (
		oops = 0x0001
		cmd  = 0x0002
		heap = 0x0004
	)

	if CurrentFrame == nil || FrameOops == nil || FrameCmd == nil || FrameHeap == nil {
		ScreenMessage(DbgInvalidFramePtr)
		return false
	}

	if FirstSpan == nil {
		ScreenMessage(DbgInvalidSpanPtr)
		return false
	}

	// Validate the data structure.
	frameList := 0 // Bit mask OOPS, CMD, HEAP
	scrRow := 0
	thisSpan := FirstSpan
	var prevSpan *SpanObject

	for thisSpan != nil {
		if thisSpan.BLink != prevSpan {
			ScreenMessage(DbgInvalidBlink)
			return false
		}
		if thisSpan.MarkOne == nil || thisSpan.MarkTwo == nil {
			ScreenMessage(DbgMarkPtrIsNil)
			return false
		}
		if thisSpan.Code != nil {
			if thisSpan.Code.Ref == 0 {
				ScreenMessage(DbgRefCountIsZero)
				return false
			}
		}
		if thisSpan.Frame != nil {
			thisFrame := thisSpan.Frame
			if thisFrame == FrameCmd {
				frameList |= cmd
			} else if thisFrame == FrameOops {
				frameList |= oops
			}
			if thisFrame == FrameHeap {
				frameList |= heap
			}

			if thisFrame.FirstGroup == nil || thisFrame.LastGroup == nil {
				ScreenMessage(DbgInvalidGroupPtr)
				return false
			}
			if thisFrame.FirstGroup.BLink != nil {
				ScreenMessage(DbgFirstNotAtTop)
				return false
			}
			endGroup := thisFrame.LastGroup.FLink
			if endGroup != nil {
				ScreenMessage(DbgLastNotAtEnd)
				return false
			}
			thisGroup := thisFrame.FirstGroup
			var prevGroup *GroupObject
			thisLine := thisFrame.FirstGroup.FirstLine
			var prevLine *LineHdrObject
			var endLine *LineHdrObject
			lineNr := 1

			for thisGroup != endGroup {
				if thisGroup.BLink != prevGroup {
					ScreenMessage(DbgInvalidBlink)
					return false
				}
				if thisGroup.Frame != thisFrame {
					ScreenMessage(DbgInvalidFramePtr)
					return false
				}
				if thisGroup.FirstLine == nil || thisGroup.LastLine == nil {
					ScreenMessage(DbgLinePtrIsNil)
					return false
				}
				if thisGroup.FirstLine != thisLine {
					ScreenMessage(DbgFirstNotAtTop)
					return false
				}
				lineCount := 0
				endLine = thisGroup.LastLine.FLink

				for thisLine != endLine {
					if thisLine.BLink != prevLine {
						ScreenMessage(DbgInvalidBlink)
						return false
					}
					if thisLine.Group != thisGroup {
						ScreenMessage(DbgInvalidGroupPtr)
						return false
					}
					if thisLine.OffsetNr != lineCount {
						ScreenMessage(DbgInvalidOffsetNr)
						return false
					}
					for _, thisMark := range thisLine.Marks {
						if thisMark.Line != thisLine {
							ScreenMessage(DbgInvalidLinePtr)
							return false
						}
					}
					if thisLine.Str == nil && thisLine.Len() != 0 {
						ScreenMessage(DbgInvalidLineLength)
						return false
					}
					if thisLine.Used > thisLine.Len() {
						ScreenMessage(DbgInvalidLineUsedLength)
						return false
					}
					if thisLine.ScrRowNr != scrRow {
						if thisLine == ScrTopLine {
							scrRow = thisLine.ScrRowNr
						} else {
							ScreenMessage(DbgInvalidScrRowNr)
							return false
						}
					}
					if scrRow != 0 {
						if thisLine != ScrBotLine {
							scrRow++
						} else {
							scrRow = 0
						}
					}
					lineCount++
					prevLine = thisLine
					thisLine = thisLine.FLink
				}

				if thisGroup.LastLine != prevLine {
					ScreenMessage(DbgLastNotAtEnd)
					return false
				}
				if thisGroup.FirstLineNr != lineNr {
					ScreenMessage(DbgInvalidLineNr)
					return false
				}
				if thisGroup.NrLines != lineCount {
					ScreenMessage(DbgInvalidNrLines)
					return false
				}
				lineNr = lineNr + thisGroup.NrLines
				prevGroup = thisGroup
				thisGroup = thisGroup.FLink
			}

			if thisFrame.FirstGroup.FirstLine.BLink != nil {
				ScreenMessage(DbgFirstNotAtTop)
				return false
			}
			if endLine != nil {
				ScreenMessage(DbgLastNotAtEnd)
				return false
			}
			if thisFrame.Dot == nil {
				ScreenMessage(DbgMarkPtrIsNil)
				return false
			}
			if thisFrame.Dot.Line.Group.Frame != thisFrame {
				ScreenMessage(DbgMarkInWrongFrame)
				return false
			}
			for markNr := MinMarkNumber; markNr <= MaxMarkNumber; markNr++ {
				if thisFrame.Marks[markNr-MinMarkNumber] != nil {
					if thisFrame.Marks[markNr-MinMarkNumber].Line.Group.Frame != thisFrame {
						ScreenMessage(DbgMarkInWrongFrame)
						return false
					}
				}
			}
			if thisFrame.ScrHeight == 0 || thisFrame.ScrHeight > TerminalInfo.Height {
				ScreenMessage(DbgInvalidScrParam)
				return false
			}
			if thisFrame.ScrWidth == 0 || thisFrame.ScrWidth > TerminalInfo.Width {
				ScreenMessage(DbgInvalidScrParam)
				return false
			}
			if thisFrame.Span != thisSpan {
				ScreenMessage(DbgInvalidSpanPtr)
				return false
			}
			if thisFrame.MarginLeft >= thisFrame.MarginRight {
				ScreenMessage(MsgLeftMarginGeRight)
				return false
			}
			if thisSpan.MarkOne.Line.Group.Frame != thisFrame ||
				thisSpan.MarkTwo.Line.Group.Frame != thisFrame {
				ScreenMessage(DbgMarkInWrongFrame)
				return false
			}
		} else if thisSpan.MarkOne.Line.Group.Frame != thisSpan.MarkTwo.Line.Group.Frame {
			ScreenMessage(DbgMarksFromDiffFrames)
			return false
		}
		prevSpan = thisSpan
		thisSpan = thisSpan.FLink
	}

	if frameList != (cmd | oops | heap) {
		ScreenMessage(DbgNeededFrameNotFound)
		return false
	}

	return true
}
