/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         SPAN
//
// Description:  Creation/destruction, manipulation of spans.

package ludwig

// SpanFind finds a span of the specified name.
// Returns a pointer to the entry in the span table if found.
// If the span is a frame then reset the frame's two marks.
func SpanFind(spanName string, ptr **SpanObject, oldp **SpanObject) bool {
	*oldp = nil
	*ptr = FirstSpan
	if *ptr == nil {
		return false
	}
	for (*ptr).Name < spanName {
		*oldp = *ptr
		*ptr = (*ptr).FLink
		if *ptr == nil {
			return false
		}
	}
	if (*ptr).Name == spanName {
		if (*ptr).Frame != nil {
			if !MarkCreate((*ptr).Frame.FirstGroup.FirstLine, 1, &(*ptr).MarkOne) ||
				!MarkCreate((*ptr).Frame.LastGroup.LastLine, 1, &(*ptr).MarkTwo) {
				return false
			}
		}
		return true
	}
	return false
}

// SpanCreate creates a span of the specified name over the specified range of lines.
// Checks first that a span of this name doesn't already exist. If it does, it re-defines it.
// Fails if span already exists and is a frame.
func SpanCreate(spanName string, firstMark *MarkObject, lastMark *MarkObject) bool {
	var p *SpanObject
	var ptr *SpanObject
	var oldp *SpanObject
	var mrk1 *MarkObject
	var mrk2 *MarkObject

	if SpanFind(spanName, &p, &oldp) {
		if p.Frame != nil {
			ScreenMessage(MsgFrameOfThatNameExists)
			return false
		}
		ptr = p
		if ptr.Code != nil {
			CodeDiscard(&ptr.Code)
		}
		mrk1 = ptr.MarkOne
		mrk2 = ptr.MarkTwo
	} else {
		mrk1 = nil
		mrk2 = nil
		ptr = &SpanObject{}
		ptr.Name = spanName
		ptr.Code = nil
		// Now hook the span into the span structure
		if p == nil {
			ptr.FLink = nil
		} else {
			ptr.FLink = p
			p.BLink = ptr
		}
		if oldp == nil {
			ptr.BLink = nil
			FirstSpan = ptr
		} else {
			ptr.BLink = oldp
			oldp.FLink = ptr
		}
	}

	if !MarkCreate(firstMark.Line, firstMark.Col, &mrk1) {
		return false
	}
	if !MarkCreate(lastMark.Line, lastMark.Col, &mrk2) {
		return false
	}

	var lineNrFirst int
	var lineNrLast int
	if LineToNumber(mrk1.Line, &lineNrFirst) && LineToNumber(mrk2.Line, &lineNrLast) {
		ptr.Frame = nil
		if (lineNrFirst < lineNrLast) ||
			((lineNrFirst == lineNrLast) && (mrk1.Col < mrk2.Col)) {
			// Marks are in the right order
			ptr.MarkOne = mrk1
			ptr.MarkTwo = mrk2
		} else {
			// Marks are in reverse order
			ptr.MarkOne = mrk2
			ptr.MarkTwo = mrk1
		}
		return true
	}
	return false
}

// SpanDestroy destroys the specified span.
// Fails if span is a frame or if span is not destroyed.
func SpanDestroy(span **SpanObject) bool {
	if (*span).Frame != nil {
		ScreenMessage(MsgCantKillFrame)
		return false
	}
	if (*span).Code != nil {
		CodeDiscard(&(*span).Code)
	}
	if (*span).BLink != nil {
		(*span).BLink.FLink = (*span).FLink
	} else {
		FirstSpan = (*span).FLink
	}
	if (*span).FLink != nil {
		(*span).FLink.BLink = (*span).BLink
	}
	if MarkDestroy(&(*span).MarkOne) {
		if MarkDestroy(&(*span).MarkTwo) {
			*span = nil
			return true
		}
	}
	return false
}

// SpanIndex displays the list of spans. This is the \SI command.
func SpanIndex() bool {
	ScreenUnload()
	ScreenHome(true)
	p := FirstSpan
	haveSpan := false
	ScreenWriteln()
	ScreenWriteStr(0, "Spans")
	ScreenWriteln()
	ScreenWriteStr(0, "=====")
	ScreenWriteln()
	lineCount := 3
	for p != nil {
		if p.Frame == nil {
			haveSpan = true
			if lineCount > TerminalInfo.Height-2 {
				ScreenPause()
				ScreenHome(true)
				ScreenWriteln()
				ScreenWriteStr(0, "Spans")
				ScreenWriteln()
				ScreenWriteStr(0, "=====")
				ScreenWriteln()
				lineCount = 3
			}
			ScreenWriteNameStr(0, p.Name, NameLen)
			ScreenWriteStr(0, " : ")
			continu := p.MarkOne.Line != p.MarkTwo.Line
			var spanStart string
			if p.MarkOne.Col <= p.MarkOne.Line.Used {
				if !continu {
					continu = p.MarkTwo.Col-p.MarkOne.Col > NameLen
					toCopy := NameLen
					if p.MarkTwo.Col-p.MarkOne.Col < toCopy {
						toCopy = p.MarkTwo.Col - p.MarkOne.Col
					}
					spanStart = p.MarkOne.Line.Str.Slice(p.MarkOne.Col, p.MarkOne.Col+toCopy)
				} else {
					toCopy := min(p.MarkOne.Line.Used+1-p.MarkOne.Col, NameLen)
					spanStart = p.MarkOne.Line.Str.Slice(p.MarkOne.Col, p.MarkOne.Col+toCopy)
				}
			}
			ScreenWriteNameStr(0, spanStart, NameLen)
			if continu {
				ScreenWriteStr(1, "...")
			}
			ScreenWriteln()
			lineCount++
		}
		p = p.FLink
	}
	if !haveSpan {
		ScreenWriteStr(10, "<none>")
		ScreenWriteln()
		lineCount++
	}
	firstTime := true
	oldCount := lineCount
	p = FirstSpan
	lineCount = TerminalInfo.Height
	for p != nil {
		if p.Frame != nil {
			if lineCount > TerminalInfo.Height-2 {
				if !firstTime {
					ScreenPause()
					ScreenHome(true)
				}
				ScreenWriteln()
				ScreenWriteStr(0, "Frames")
				ScreenWriteln()
				ScreenWriteStr(0, "======")
				ScreenWriteln()
				if firstTime {
					lineCount = oldCount + 3
					firstTime = false
				} else {
					lineCount = 3
				}
			}
			ScreenWriteNameStr(0, p.Name, NameLen)
			ScreenWriteln()
			lineCount++
			var fylNam string
			if p.Frame.InputFile != 0 {
				ScreenWriteStr(0, "  Input:  ")
				FileName(Files[p.Frame.InputFile], 70, &fylNam)
				ScreenWriteFileNameStr(0, fylNam, len(fylNam))
				ScreenWriteln()
				lineCount++
			}
			if p.Frame.OutputFile != 0 {
				ScreenWriteStr(0, "  Output: ")
				FileName(Files[p.Frame.OutputFile], 70, &fylNam)
				ScreenWriteFileNameStr(0, fylNam, len(fylNam))
				ScreenWriteln()
				lineCount++
			}
		}
		p = p.FLink
	}
	ScreenPause()
	return true
}
