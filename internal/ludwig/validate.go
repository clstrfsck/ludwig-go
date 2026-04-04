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
func ValidateCommand(currentFrame, frameOops, frameCmd, frameHeap *FrameObject) bool {
	/*
	  Purpose  : Validate the data structure.
	  Inputs   : none.
	  Outputs  : none.
	  Bugchecks: .lots and lots of them!
	*/

	const (
		oops = 0x0001
		cmd  = 0x0002
		heap = 0x0004
	)

	if currentFrame == nil || frameOops == nil || frameCmd == nil || frameHeap == nil {
		ScreenMessage(&Screen, DbgInvalidFramePtr)
		return false
	}

	if FirstSpan == nil {
		ScreenMessage(&Screen, DbgInvalidSpanPtr)
		return false
	}

	// Validate the data structure.
	frameList := 0 // Bit mask OOPS, CMD, HEAP
	scrRow := 0
	thisSpan := FirstSpan
	var prevSpan *SpanObject

	for thisSpan != nil {
		if thisSpan.BLink != prevSpan {
			ScreenMessage(&Screen, DbgInvalidBlink)
			return false
		}
		if thisSpan.MarkOne == nil || thisSpan.MarkTwo == nil {
			ScreenMessage(&Screen, DbgMarkPtrIsNil)
			return false
		}
		if thisSpan.Code != nil {
			if thisSpan.Code.Ref == 0 {
				ScreenMessage(&Screen, DbgRefCountIsZero)
				return false
			}
		}
		if thisSpan.Frame != nil {
			thisFrame := thisSpan.Frame
			switch thisFrame {
			case frameCmd:
				frameList |= cmd
			case frameOops:
				frameList |= oops
			case frameHeap:
				frameList |= heap
			}

			if thisFrame.FirstGroup == nil || thisFrame.LastGroup == nil {
				ScreenMessage(&Screen, DbgInvalidGroupPtr)
				return false
			}
			if thisFrame.FirstGroup.BLink != nil {
				ScreenMessage(&Screen, DbgFirstNotAtTop)
				return false
			}
			endGroup := thisFrame.LastGroup.FLink
			if endGroup != nil {
				ScreenMessage(&Screen, DbgLastNotAtEnd)
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
					ScreenMessage(&Screen, DbgInvalidBlink)
					return false
				}
				if thisGroup.Frame != thisFrame {
					ScreenMessage(&Screen, DbgInvalidFramePtr)
					return false
				}
				if thisGroup.FirstLine == nil || thisGroup.LastLine == nil {
					ScreenMessage(&Screen, DbgLinePtrIsNil)
					return false
				}
				if thisGroup.FirstLine != thisLine {
					ScreenMessage(&Screen, DbgFirstNotAtTop)
					return false
				}
				lineCount := 0
				endLine = thisGroup.LastLine.FLink

				for thisLine != endLine {
					if thisLine.BLink != prevLine {
						ScreenMessage(&Screen, DbgInvalidBlink)
						return false
					}
					if thisLine.Group != thisGroup {
						ScreenMessage(&Screen, DbgInvalidGroupPtr)
						return false
					}
					if thisLine.OffsetNr != lineCount {
						ScreenMessage(&Screen, DbgInvalidOffsetNr)
						return false
					}
					for _, thisMark := range thisLine.Marks {
						if thisMark.Line != thisLine {
							ScreenMessage(&Screen, DbgInvalidLinePtr)
							return false
						}
					}
					if thisLine.Str == nil && thisLine.Len() != 0 {
						ScreenMessage(&Screen, DbgInvalidLineLength)
						return false
					}
					if thisLine.Used > thisLine.Len() {
						ScreenMessage(&Screen, DbgInvalidLineUsedLength)
						return false
					}
					if thisLine.ScrRowNr != scrRow {
						if thisLine == Screen.TopLine {
							scrRow = thisLine.ScrRowNr
						} else {
							ScreenMessage(&Screen, DbgInvalidScrRowNr)
							return false
						}
					}
					if scrRow != 0 {
						if thisLine != Screen.BotLine {
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
					ScreenMessage(&Screen, DbgLastNotAtEnd)
					return false
				}
				if thisGroup.FirstLineNr != lineNr {
					ScreenMessage(&Screen, DbgInvalidLineNr)
					return false
				}
				if thisGroup.NrLines != lineCount {
					ScreenMessage(&Screen, DbgInvalidNrLines)
					return false
				}
				lineNr = lineNr + thisGroup.NrLines
				prevGroup = thisGroup
				thisGroup = thisGroup.FLink
			}

			if thisFrame.FirstGroup.FirstLine.BLink != nil {
				ScreenMessage(&Screen, DbgFirstNotAtTop)
				return false
			}
			if endLine != nil {
				ScreenMessage(&Screen, DbgLastNotAtEnd)
				return false
			}
			if thisFrame.Dot == nil {
				ScreenMessage(&Screen, DbgMarkPtrIsNil)
				return false
			}
			if thisFrame.Dot.Line.Group.Frame != thisFrame {
				ScreenMessage(&Screen, DbgMarkInWrongFrame)
				return false
			}
			for markNr := 0; markNr <= MaxMarkNumber; markNr++ {
				if thisFrame.Marks[markNr] != nil {
					if thisFrame.Marks[markNr].Line.Group.Frame != thisFrame {
						ScreenMessage(&Screen, DbgMarkInWrongFrame)
						return false
					}
				}
			}
			if thisFrame.ScrHeight == 0 || thisFrame.ScrHeight > TerminalInfo.Height {
				ScreenMessage(&Screen, DbgInvalidScrParam)
				return false
			}
			if thisFrame.ScrWidth == 0 || thisFrame.ScrWidth > TerminalInfo.Width {
				ScreenMessage(&Screen, DbgInvalidScrParam)
				return false
			}
			if thisFrame.Span != thisSpan {
				ScreenMessage(&Screen, DbgInvalidSpanPtr)
				return false
			}
			if thisFrame.MarginLeft >= thisFrame.MarginRight {
				ScreenMessage(&Screen, MsgLeftMarginGeRight)
				return false
			}
			if thisSpan.MarkOne.Line.Group.Frame != thisFrame ||
				thisSpan.MarkTwo.Line.Group.Frame != thisFrame {
				ScreenMessage(&Screen, DbgMarkInWrongFrame)
				return false
			}
		} else if thisSpan.MarkOne.Line.Group.Frame != thisSpan.MarkTwo.Line.Group.Frame {
			ScreenMessage(&Screen, DbgMarksFromDiffFrames)
			return false
		}
		prevSpan = thisSpan
		thisSpan = thisSpan.FLink
	}

	if frameList != (cmd | oops | heap) {
		ScreenMessage(&Screen, DbgNeededFrameNotFound)
		return false
	}

	return true
}
