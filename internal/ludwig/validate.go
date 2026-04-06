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
func ValidateCommand(currentFrame *FrameObject, specialFrames *SpecialFrames) bool {
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

	if currentFrame == nil || specialFrames == nil {
		Screen.Message(DbgInvalidFramePtr)
		return false
	}
	if specialFrames.Oops() == nil || specialFrames.Cmd() == nil || specialFrames.Heap() == nil {
		Screen.Message(DbgInvalidFramePtr)
		return false
	}

	if FirstSpan == nil {
		Screen.Message(DbgInvalidSpanPtr)
		return false
	}

	// Validate the data structure.
	frameList := 0 // Bit mask OOPS, CMD, HEAP
	scrRow := 0
	thisSpan := FirstSpan
	var prevSpan *SpanObject

	for thisSpan != nil {
		if thisSpan.BLink != prevSpan {
			Screen.Message(DbgInvalidBlink)
			return false
		}
		if thisSpan.MarkOne == nil || thisSpan.MarkTwo == nil {
			Screen.Message(DbgMarkPtrIsNil)
			return false
		}
		if thisSpan.Code != nil {
			if thisSpan.Code.Ref == 0 {
				Screen.Message(DbgRefCountIsZero)
				return false
			}
		}
		if thisSpan.Frame != nil {
			thisFrame := thisSpan.Frame
			switch thisFrame {
			case specialFrames.Cmd():
				frameList |= cmd
			case specialFrames.Oops():
				frameList |= oops
			case specialFrames.Heap():
				frameList |= heap
			}

			if thisFrame.FirstGroup == nil || thisFrame.LastGroup == nil {
				Screen.Message(DbgInvalidGroupPtr)
				return false
			}
			if thisFrame.FirstGroup.BLink != nil {
				Screen.Message(DbgFirstNotAtTop)
				return false
			}
			endGroup := thisFrame.LastGroup.FLink
			if endGroup != nil {
				Screen.Message(DbgLastNotAtEnd)
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
					Screen.Message(DbgInvalidBlink)
					return false
				}
				if thisGroup.Frame != thisFrame {
					Screen.Message(DbgInvalidFramePtr)
					return false
				}
				if thisGroup.FirstLine == nil || thisGroup.LastLine == nil {
					Screen.Message(DbgLinePtrIsNil)
					return false
				}
				if thisGroup.FirstLine != thisLine {
					Screen.Message(DbgFirstNotAtTop)
					return false
				}
				lineCount := 0
				endLine = thisGroup.LastLine.FLink

				for thisLine != endLine {
					if thisLine.BLink != prevLine {
						Screen.Message(DbgInvalidBlink)
						return false
					}
					if thisLine.Group != thisGroup {
						Screen.Message(DbgInvalidGroupPtr)
						return false
					}
					if thisLine.OffsetNr != lineCount {
						Screen.Message(DbgInvalidOffsetNr)
						return false
					}
					for _, thisMark := range thisLine.Marks {
						if thisMark.Line != thisLine {
							Screen.Message(DbgInvalidLinePtr)
							return false
						}
					}
					if thisLine.Str == nil && thisLine.Len() != 0 {
						Screen.Message(DbgInvalidLineLength)
						return false
					}
					if thisLine.Used > thisLine.Len() {
						Screen.Message(DbgInvalidLineUsedLength)
						return false
					}
					if thisLine.ScrRowNr != scrRow {
						if thisLine == Screen.TopLine {
							scrRow = thisLine.ScrRowNr
						} else {
							Screen.Message(DbgInvalidScrRowNr)
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
					Screen.Message(DbgLastNotAtEnd)
					return false
				}
				if thisGroup.FirstLineNr != lineNr {
					Screen.Message(DbgInvalidLineNr)
					return false
				}
				if thisGroup.NrLines != lineCount {
					Screen.Message(DbgInvalidNrLines)
					return false
				}
				lineNr = lineNr + thisGroup.NrLines
				prevGroup = thisGroup
				thisGroup = thisGroup.FLink
			}

			if thisFrame.FirstGroup.FirstLine.BLink != nil {
				Screen.Message(DbgFirstNotAtTop)
				return false
			}
			if endLine != nil {
				Screen.Message(DbgLastNotAtEnd)
				return false
			}
			if thisFrame.Dot == nil {
				Screen.Message(DbgMarkPtrIsNil)
				return false
			}
			if thisFrame.Dot.Line.Group.Frame != thisFrame {
				Screen.Message(DbgMarkInWrongFrame)
				return false
			}
			for markNr := 0; markNr <= MaxMarkNumber; markNr++ {
				if thisFrame.Marks[markNr] != nil {
					if thisFrame.Marks[markNr].Line.Group.Frame != thisFrame {
						Screen.Message(DbgMarkInWrongFrame)
						return false
					}
				}
			}
			if thisFrame.ScrHeight == 0 || thisFrame.ScrHeight > TerminalInfo.Height {
				Screen.Message(DbgInvalidScrParam)
				return false
			}
			if thisFrame.ScrWidth == 0 || thisFrame.ScrWidth > TerminalInfo.Width {
				Screen.Message(DbgInvalidScrParam)
				return false
			}
			if thisFrame.Span != thisSpan {
				Screen.Message(DbgInvalidSpanPtr)
				return false
			}
			if thisFrame.MarginLeft >= thisFrame.MarginRight {
				Screen.Message(MsgLeftMarginGeRight)
				return false
			}
			if thisSpan.MarkOne.Line.Group.Frame != thisFrame ||
				thisSpan.MarkTwo.Line.Group.Frame != thisFrame {
				Screen.Message(DbgMarkInWrongFrame)
				return false
			}
		} else if thisSpan.MarkOne.Line.Group.Frame != thisSpan.MarkTwo.Line.Group.Frame {
			Screen.Message(DbgMarksFromDiffFrames)
			return false
		}
		prevSpan = thisSpan
		thisSpan = thisSpan.FLink
	}

	if frameList != (cmd | oops | heap) {
		Screen.Message(DbgNeededFrameNotFound)
		return false
	}

	return true
}
