/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         EXEC
//
// Description:  The primitive LUDWIG commands.

package ludwig

import (
	"math"
)

func iabs(n int) int {
	if n < 0 {
		if n == math.MinInt {
			// Going to assume this is close enough
			return math.MaxInt
		}
		return -n
	}
	return n
}

// ExecComputeLineRange returns the range of lines specified by the REPT/COUNT pair.
// It returns false if the range does not exist.
// It returns firstLine as nil if the range is empty.
// The range returned WILL NOT include the null line.
// It is assumed that the mark (if any) has been checked for validity.
func ExecComputeLineRange(
	frame *FrameObject,
	rept LeadParam,
	count int,
	firstLine **LineHdrObject,
	lastLine **LineHdrObject,
) bool {
	*firstLine = frame.Dot.Line
	*lastLine = frame.Dot.Line

	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		if count == 0 {
			*firstLine = nil
		} else if count <= 20 { // TRY TO OPTIMIZE COMMON CASE
			for lineNr := 1; lineNr < count; lineNr++ {
				*lastLine = (*lastLine).FLink
				if *lastLine == nil {
					return false
				}
			}
			if (*lastLine).FLink == nil {
				return false
			}
		} else {
			lineNr := LineToNumber(*firstLine)
			*lastLine = LineFromNumber(frame, lineNr+count-1)
			if *lastLine == nil {
				return false
			}
			if (*lastLine).FLink == nil {
				return false
			}
		}

	case LeadParamMinus, LeadParamNInt:
		count = -count
		*lastLine = frame.Dot.Line.BLink
		if *lastLine == nil {
			return false
		}
		if count <= 20 {
			for lineNr := 1; lineNr <= count; lineNr++ {
				*firstLine = (*firstLine).BLink
				if *firstLine == nil {
					return false
				}
			}
		} else {
			lineNr := LineToNumber(*lastLine)
			if count > lineNr {
				return false
			}
			lineNr = lineNr - count + 1
			*firstLine = LineFromNumber(frame, lineNr)
		}

	case LeadParamPIndef:
		if frame.Dot.Line.FLink == nil {
			*firstLine = nil
		} else {
			*lastLine = frame.LastGroup.LastLine.BLink
		}

	case LeadParamNIndef:
		*lastLine = frame.Dot.Line.BLink
		if *lastLine == nil {
			*firstLine = nil
		} else {
			*firstLine = frame.FirstGroup.FirstLine
		}

	case LeadParamMarker:
		markLine := frame.Marks[count].Line
		if markLine == *firstLine { // TRY TO OPTIMIZE MOST COMMON CASES
			*firstLine = nil
		} else if markLine.FLink == *firstLine {
			*firstLine = markLine
			*lastLine = markLine
		} else if markLine.BLink == *firstLine {
			*lastLine = *firstLine
		} else { // DO IT THE HARD WAY!
			markLineNr := LineToNumber(markLine)
			lineNr := LineToNumber(frame.Dot.Line)
			if markLineNr < lineNr {
				*firstLine = markLine
				*lastLine = (*lastLine).BLink
			} else {
				*lastLine = markLine.BLink
			}
		}
	}

	return true
}

// Execute executes a command with the specified parameters
func Execute(
	frame *FrameObject,
	specialFrames *SpecialFrames,
	command Commands,
	rept LeadParam,
	count int,
	tparam *TParObject,
	fromSpan bool,
) (currentFrame *FrameObject, cmdSuccess bool) {
	currentFrame = frame
	ExecLevel++
	defer func() {
		ExecLevel--
	}()

	if TtControlC {
		return
	}
	if ExecLevel == MaxExecRecursion {
		Screen.Message(MsgCommandRecursionLimit)
		return
	}

	// Fix commands which use marks without using @ in the syntax
	switch command {
	case CmdMark:
		if count == 0 || iabs(count) > MaxUserMarkNumber {
			Screen.Message(MsgIllegalMarkNumber)
			return
		}
	case CmdSpanDefine:
		if rept == LeadParamNone || rept == LeadParamPInt {
			if count == 0 || count > MaxUserMarkNumber {
				Screen.Message(MsgIllegalMarkNumber)
				return
			}
			rept = LeadParamMarker
		}
	}

	// Check the mark, assign theMark to the mark
	var theMark *MarkObject
	if rept == LeadParamMarker {
		theMark = currentFrame.Marks[count]
		if theMark == nil {
			Screen.Message(MsgMarkNotDefined)
			return
		}
	}

	// Save the current value of DOT and current frame for use by equals
	oldDot := *currentFrame.Dot
	oldFrame := currentFrame

	// Execute the command
	switch command {
	case CmdAdvance:
		// Establish which line to advance to
		cmdSuccess = (rept == LeadParamPIndef || rept == LeadParamNIndef || rept == LeadParamMarker)
		newLine := currentFrame.Dot.Line
		switch rept {
		case LeadParamNone, LeadParamPlus, LeadParamPInt:
			if count < 20 {
				for count > 0 {
					count--
					newLine = newLine.FLink
					if newLine == nil {
						return
					}
				}
			} else {
				lineNr := LineToNumber(newLine)
				newLine = LineFromNumber(currentFrame, lineNr+count)
				if newLine == nil {
					return
				}
			}
			// if flink is nil we are on eop-line, so fail
			if newLine.FLink == nil {
				return
			}
			cmdSuccess = true

		case LeadParamMinus, LeadParamNInt:
			count = -count
			if count < 20 {
				for count > 0 {
					count--
					newLine = newLine.BLink
					if newLine == nil {
						return
					}
				}
			} else {
				lineNr := LineToNumber(newLine)
				if count >= lineNr {
					return
				}
				newLine = LineFromNumber(currentFrame, lineNr-count)
				if newLine == nil {
					return
				}
			}
			cmdSuccess = true

		case LeadParamPIndef:
			newLine = currentFrame.LastGroup.LastLine

		case LeadParamNIndef:
			newLine = currentFrame.FirstGroup.FirstLine

		case LeadParamMarker:
			newLine = theMark.Line
		}

		MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkEquals])
		MarkCreate(newLine, 1, &currentFrame.Dot)

	case CmdBridge, CmdNext:
		var request TParObject
		if TparGet1(currentFrame, tparam, command, &request) {
			cmdSuccess = NextbridgeCommand(currentFrame, count, &request, command == CmdBridge)
		}

	case CmdCaseEdit, CmdCaseLow, CmdCaseUp, CmdDittoDown, CmdDittoUp:
		cmdSuccess = CaseDittoCommand(currentFrame, command, rept, count, fromSpan)

	case CmdDeleteChar:
		if rept != LeadParamMarker {
			cmdSuccess = CharcmdDelete(currentFrame, rept, count, fromSpan)
		} else {
			theOtherMark := currentFrame.Dot
			lineNr := LineToNumber(currentFrame.Dot.Line)
			line2Nr := LineToNumber(theMark.Line)
			if lineNr > line2Nr || (lineNr == line2Nr && currentFrame.Dot.Col > theMark.Col) {
				// Reverse mark pointers to get theOtherMark first
				anotherMark := theMark
				theMark = theOtherMark
				theOtherMark = anotherMark
			}
			if currentFrame != specialFrames.Oops() {
				// Make sure oops_span is okay
				MarkCreate(specialFrames.Oops().LastGroup.LastLine, 1, &specialFrames.Oops().Span.MarkTwo)
				cmdSuccess = TextMove(
					false,                                   // Don't copy, transfer
					1,                                       // One instance of
					theOtherMark,                            // starting pos
					theMark,                                 // ending pos
					specialFrames.Oops().Span.MarkTwo,       // destination
					&specialFrames.Oops().Marks[MarkEquals], // leave at start
					&specialFrames.Oops().Dot,               // leave at end
				)
				specialFrames.Oops().TextModified = true
				MarkCreate(
					specialFrames.Oops().Dot.Line,
					specialFrames.Oops().Dot.Col,
					&specialFrames.Oops().Marks[MarkModified],
				)
			} else {
				cmdSuccess = TextRemove(theOtherMark, theMark)
			}
			currentFrame.TextModified = true
			MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkModified])
		}

	case CmdDeleteLine:
		// Establish which lines to kill, this is common to K and FW cmds
		var firstLine, lastLine *LineHdrObject
		if !ExecComputeLineRange(currentFrame, rept, count, &firstLine, &lastLine) {
			return
		}
		if firstLine != nil {
			dotCol := currentFrame.Dot.Col
			if lastLine.FLink == nil {
				return
			}
			MarksSqueeze(firstLine, 1, lastLine.FLink, 1)
			LinesExtract(firstLine, lastLine)
			if currentFrame != specialFrames.Oops() {
				LinesInject(firstLine, lastLine, specialFrames.Oops().LastGroup.LastLine)
				MarkCreate(firstLine, 1, &specialFrames.Oops().Marks[MarkEquals])
				MarkCreate(specialFrames.Oops().LastGroup.LastLine, 1, &specialFrames.Oops().Dot)
				specialFrames.Oops().TextModified = true
				MarkCreate(
					specialFrames.Oops().Dot.Line,
					specialFrames.Oops().Dot.Col,
					&specialFrames.Oops().Marks[MarkModified],
				)
			}
			currentFrame.Dot.Col = dotCol
			currentFrame.TextModified = true
			MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkModified])
		}
		cmdSuccess = true

	case CmdBacktab, CmdDown, CmdHome, CmdLeft, CmdReturn, CmdRight, CmdTab, CmdUp:
		if command == CmdReturn && EditMode == ModeInsert &&
			currentFrame.Options.Has(OptNewLine) {
			if currentFrame.Dot.Line.FLink == nil {
				TextRealizeNull(currentFrame.Dot.Line)
				cmdSuccess = ArrowCommand(currentFrame, command, rept, count, fromSpan)
			} else {
				currentFrame, cmdSuccess = Execute(
					currentFrame,
					specialFrames,
					CmdSplitLine,
					rept,
					count,
					tparam,
					fromSpan,
				)
			}
		} else {
			cmdSuccess = ArrowCommand(currentFrame, command, rept, count, fromSpan)
		}

	case CmdDump:
		// TODO: Implement this in a sensible way for golang
		cmdSuccess = false

	case CmdEqualColumn:
		var request TParObject
		i := 1 // Start of column number, j receives column number
		if TparGet1(currentFrame, tparam, command, &request) {
			if j, found := TparToInt(&request, &i); found {
				switch rept {
				case LeadParamNone, LeadParamPlus:
					cmdSuccess = (currentFrame.Dot.Col == j)
				case LeadParamMinus:
					cmdSuccess = (currentFrame.Dot.Col != j)
				case LeadParamPIndef:
					cmdSuccess = (currentFrame.Dot.Col >= j)
				case LeadParamNIndef:
					cmdSuccess = (currentFrame.Dot.Col <= j)
				}
			}
		}

	case CmdEqualEol:
		switch rept {
		case LeadParamNone, LeadParamPlus:
			cmdSuccess = (currentFrame.Dot.Col == currentFrame.Dot.Line.Used+1)
		case LeadParamMinus:
			cmdSuccess = (currentFrame.Dot.Col != currentFrame.Dot.Line.Used+1)
		case LeadParamPIndef:
			cmdSuccess = (currentFrame.Dot.Col >= currentFrame.Dot.Line.Used+1)
		case LeadParamNIndef:
			cmdSuccess = (currentFrame.Dot.Col <= currentFrame.Dot.Line.Used+1)
		}

	case CmdEqualEop, CmdEqualEof:
		cmdSuccess = (currentFrame.Dot.Line.FLink == nil)
		if command == CmdEqualEof {
			if currentFrame.InputFile != 0 {
				if !Files[currentFrame.InputFile].Eof {
					cmdSuccess = false
				}
			}
		}
		if rept == LeadParamMinus {
			cmdSuccess = !cmdSuccess
		}

	case CmdEqualMark:
		var request TParObject
		if !TparGet1(currentFrame, tparam, command, &request) {
			return
		}
		var found bool
		var j int
		if j, found = TparToMark(&request); !found {
			return
		}
		if currentFrame.Marks[j] != nil {
			switch rept {
			case LeadParamNone, LeadParamPlus, LeadParamMinus:
				if currentFrame.Marks[j].Line == currentFrame.Dot.Line &&
					currentFrame.Marks[j].Col == currentFrame.Dot.Col {
					cmdSuccess = true
				}
				if rept == LeadParamMinus {
					cmdSuccess = !cmdSuccess
				}
			case LeadParamPIndef, LeadParamNIndef:
				if currentFrame.Marks[j].Line == currentFrame.Dot.Line {
					if rept == LeadParamPIndef {
						cmdSuccess = (currentFrame.Dot.Col >= currentFrame.Marks[j].Col)
					} else {
						cmdSuccess = (currentFrame.Dot.Col <= currentFrame.Marks[j].Col)
					}
				} else {
					lineNr := LineToNumber(currentFrame.Dot.Line)
					line2Nr := LineToNumber(currentFrame.Marks[j].Line)
					if rept == LeadParamPIndef {
						cmdSuccess = (lineNr >= line2Nr)
					} else {
						cmdSuccess = (lineNr <= line2Nr)
					}
				}
			}
		}

	case CmdEqualString:
		var request TParObject
		if TparGet1(currentFrame, tparam, command, &request) {
			if request.Len == 0 {
				// If didn't specify, use default
				request = currentFrame.EqsTpar
				if request.Len == 0 {
					Screen.Message(MsgNoDefaultStr)
					return
				}
			} else {
				currentFrame.EqsTpar = request // If did specify, save for next time
			}
		}
		cmdSuccess = EqsGetRepEqs(currentFrame, rept, request)

	case CmdDoLastCommand, CmdExecuteString:
		if currentFrame == specialFrames.Cmd() {
			Screen.Message(MsgNotWhileEditingCmd)
			return
		}
		if command == CmdExecuteString {
			var request TParObject
			if !TparGet1(currentFrame, tparam, command, &request) {
				return
			}

			specialFrames.Cmd().ReturnFrame = currentFrame
			currentFrame = specialFrames.Cmd()

			// Zap frame COMMAND's current contents
			firstLine := specialFrames.Cmd().FirstGroup.FirstLine
			lastLine := specialFrames.Cmd().LastGroup.LastLine.BLink
			if lastLine != nil {
				MarksSqueeze(firstLine, 1, lastLine.FLink, 1)
				LinesExtract(firstLine, lastLine)
			}

			// Insert the new tpar into frame COMMAND
			if !TextInsertTpar(&request, specialFrames.Cmd().Dot, &specialFrames.Cmd().Marks[MarkEquals]) {
				return
			}

			currentFrame = currentFrame.ReturnFrame
		}

		// Recompile and execute frame COMMAND
		var newSpan, oldSpan *SpanObject
		if SpanFind(specialFrames.Cmd().Span.Name, &newSpan, &oldSpan) {
			var ok bool
			currentFrame, ok = CodeCompile(currentFrame, specialFrames.Cmd().Span, true)
			if !ok {
				return
			}
			currentFrame, cmdSuccess = CodeInterpret(
				currentFrame,
				specialFrames,
				rept,
				count,
				specialFrames.Cmd().Span.Code,
				true,
			)
		}

	case CmdFileInput, CmdFileOutput, CmdFileEdit, CmdFileRead, CmdFileWrite,
		CmdFileRewind, CmdFileKill, CmdFileSave,
		CmdFileGlobalInput, CmdFileGlobalOutput, CmdFileGlobalRewind, CmdFileGlobalKill:
		cmdSuccess = FileCommand(currentFrame, command, rept, count, tparam, fromSpan)

	case CmdFileExecute:
		if currentFrame == specialFrames.Cmd() {
			Screen.Message(MsgNotWhileEditingCmd)
			return
		}
		var request TParObject
		if TparGet1(currentFrame, tparam, command, &request) {
			newTparam := request
			specialFrames.Cmd().ReturnFrame = currentFrame
			currentFrame = specialFrames.Cmd()
			// Zap frame COMMAND's current contents
			firstLine := specialFrames.Cmd().FirstGroup.FirstLine
			lastLine := specialFrames.Cmd().LastGroup.LastLine.BLink
			if lastLine != nil {
				MarksSqueeze(firstLine, 1, lastLine.FLink, 1)
				LinesExtract(firstLine, lastLine)
			}
			if FileCommand(currentFrame, CmdFileExecute, LeadParamNone, 0, &newTparam, false) {
				currentFrame = currentFrame.ReturnFrame
				// Recompile and execute frame COMMAND
				var newSpan, oldSpan *SpanObject
				if SpanFind(specialFrames.Cmd().Span.Name, &newSpan, &oldSpan) {
					var ok bool
					currentFrame, ok = CodeCompile(currentFrame, specialFrames.Cmd().Span, true)
					if ok {
						currentFrame, cmdSuccess = CodeInterpret(
							currentFrame,
							specialFrames,
							rept,
							count,
							specialFrames.Cmd().Span.Code,
							true,
						)
					}
				}
			} else {
				currentFrame = currentFrame.ReturnFrame
			}
		}

	case CmdFileTable:
		FileTable()
		cmdSuccess = true

	case CmdFrameEdit:
		var request TParObject
		if TparGet1(currentFrame, tparam, command, &request) {
			newName := request.Str.Slice(1, request.Len)
			if newFrame, ok := FrameEdit(currentFrame, newName); ok {
				currentFrame = newFrame
				cmdSuccess = true
			} else {
				cmdSuccess = false
			}
		}

	case CmdFrameKill:
		var request TParObject
		if TparGet1(currentFrame, tparam, command, &request) {
			newName := request.Str.Slice(1, request.Len)
			cmdSuccess = FrameKill(currentFrame, newName)
		}

	case CmdFrameParameters:
		cmdSuccess = FrameParameter(currentFrame, tparam)

	case CmdFrameReturn:
		for i := 1; i <= count; i++ {
			if currentFrame.ReturnFrame == nil {
				currentFrame = oldFrame
				return
			}
			currentFrame = currentFrame.ReturnFrame
		}
		cmdSuccess = true

	case CmdGet:
		var request TParObject
		if TparGet1(currentFrame, tparam, command, &request) {
			if request.Len == 0 {
				// If didn't specify, use default
				request = currentFrame.GetTpar
				if request.Len == 0 {
					Screen.Message(MsgNoDefaultStr)
					return
				}
			} else {
				currentFrame.GetTpar = request // If did specify, save for next time
			}
			cmdSuccess = EqsGetRepGet(currentFrame, count, request, fromSpan)
		}

	case CmdHelp:
		if LudwigMode == LudwigBatch {
			Screen.Message(MsgInteractiveModeOnly)
			return
		}
		var request TParObject
		if TparGet1(currentFrame, tparam, command, &request) {
			HelpHelp(string(request.Str.Slice(1, request.Len)))
			cmdSuccess = true // Never fails
		}

	case CmdInsertChar:
		cmdSuccess = CharcmdInsert(currentFrame, rept, count, fromSpan)

	case CmdInsertLine:
		if count != 0 {
			firstLine, lastLine := LinesCreate(iabs(count))
			LinesInject(firstLine, lastLine, currentFrame.Dot.Line)
			if count > 0 {
				MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkEquals])
				MarkCreate(firstLine, currentFrame.Dot.Col, &currentFrame.Dot)
			} else {
				MarkCreate(firstLine, currentFrame.Dot.Col, &currentFrame.Marks[MarkEquals])
			}
			currentFrame.TextModified = true
			MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkModified])
		} else {
			MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkEquals])
		}
		cmdSuccess = true

	case CmdInsertMode:
		EditMode = ModeInsert
		cmdSuccess = true

	case CmdInsertText:
		if FileData.OldCmds && !fromSpan {
			if rept == LeadParamNone {
				EditMode = ModeInsert
				cmdSuccess = true
			} else {
				Screen.Message(MsgSyntaxError)
			}
		} else {
			var request TParObject
			if TparGet1(currentFrame, tparam, command, &request) {
				if request.Con == nil {
					cmdSuccess = TextInsert(true, count, request.Str, request.Len, currentFrame.Dot)
					if cmdSuccess && (count*request.Len != 0) {
						currentFrame.TextModified = true
						MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkModified])
					}
				} else {
					for i := 1; i <= count; i++ {
						if !TextInsertTpar(&request, currentFrame.Dot, &currentFrame.Marks[MarkEquals]) {
							return
						}
					}
					currentFrame.TextModified = true
					MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkModified])
					cmdSuccess = true
				}
			}
		}

	case CmdInsertInvisible:
		if LudwigMode != LudwigScreen {
			return
		}
		var i int
		if currentFrame.Dot.Col > currentFrame.Dot.Line.Used {
			i = MaxStrLenP - currentFrame.Dot.Col
		} else {
			i = MaxStrLen - currentFrame.Dot.Line.Used
		}
		if rept == LeadParamPIndef {
			count = i
		}
		if count > i {
			return
		}
		newStr := BlankString.Clone()
		i = 0
		for i < count {
			key := VduGetKey()
			if TtControlC {
				return
			}
			if ChIsPrintable(rune(key)) {
				i++
				newStr.Set(i, byte(key))
			} else if key == 13 {
				if rept == LeadParamPIndef {
					count = i
				} else {
					i = count
				}
			} else {
				VduBeep()
			}
		}
		cmdSuccess = TextInsert(true, 1, newStr, count, currentFrame.Dot)
		if cmdSuccess && count != 0 {
			currentFrame.TextModified = true
			MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkModified])
		}

	case CmdJump:
		switch rept {
		case LeadParamNone, LeadParamPlus, LeadParamPInt:
			if currentFrame.Dot.Col+count > MaxStrLenP {
				return
			}
		case LeadParamMinus, LeadParamNInt:
			if currentFrame.Dot.Col <= -count {
				return
			}
		case LeadParamPIndef:
			if currentFrame.Dot.Col > currentFrame.Dot.Line.Used+1 {
				return
			}
			count = 1 + currentFrame.Dot.Line.Used - currentFrame.Dot.Col
		case LeadParamNIndef:
			count = 1 - currentFrame.Dot.Col
		case LeadParamMarker:
			MarkCreate(theMark.Line, theMark.Col, &currentFrame.Dot)
			count = 0
		}
		currentFrame.Dot.Col += count
		cmdSuccess = true

	case CmdLineCentre:
		cmdSuccess = WordCentre(currentFrame, rept, count)

	case CmdLineFill:
		cmdSuccess = WordFill(currentFrame, rept, count)

	case CmdLineJustify:
		cmdSuccess = WordJustify(currentFrame, rept, count)

	case CmdLineSquash:
		cmdSuccess = WordSqueeze(currentFrame, rept, count)

	case CmdLineLeft:
		cmdSuccess = WordLeft(currentFrame, rept, count)

	case CmdLineRight:
		cmdSuccess = WordRight(currentFrame, rept, count)

	case CmdWordAdvance:
		if FileData.OldCmds {
			cmdSuccess = WordAdvanceWord(currentFrame, rept, count)
		} else {
			cmdSuccess = NewwordAdvanceWord(currentFrame, rept, count)
		}

	case CmdWordDelete:
		if FileData.OldCmds {
			cmdSuccess = WordDeleteWord(currentFrame, specialFrames.Oops(), rept, count)
		} else {
			cmdSuccess = NewwordDeleteWord(currentFrame, specialFrames.Oops(), rept, count)
		}

	case CmdAdvanceParagraph:
		cmdSuccess = NewwordAdvanceParagraph(currentFrame, rept, count)

	case CmdDeleteParagraph:
		cmdSuccess = NewwordDeleteParagraph(currentFrame, specialFrames.Oops(), rept, count)

	case CmdMark:
		cmdSuccess = true
		if count < 0 {
			MarkDestroy(&currentFrame.Marks[-count])
		} else {
			MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[count])
		}

	case CmdNoop:
		// Nothing to do, as one might expect

	case CmdCommand:
		if rept == LeadParamMinus {
			if EditMode != ModeCommand {
				PreviousMode = EditMode
				EditMode = ModeCommand
			} else {
				return
			}
		} else {
			if EditMode == ModeCommand {
				EditMode = PreviousMode
			} else {
				if LudwigMode != LudwigScreen {
					Screen.Message(MsgScreenModeOnly)
					return
				}
				if !UserCommandIntroducer(currentFrame) {
					return
				}
			}
		}
		cmdSuccess = true

	case CmdOvertypeMode:
		EditMode = ModeOvertype
		cmdSuccess = true

	case CmdOvertypeText:
		if FileData.OldCmds && !fromSpan {
			if rept == LeadParamNone {
				EditMode = ModeOvertype
				cmdSuccess = true
			} else {
				Screen.Message(MsgSyntaxError)
			}
		} else {
			var request TParObject
			if TparGet1(currentFrame, tparam, command, &request) {
				cmdSuccess = TextOvertype(true, count, request.Str, request.Len, currentFrame.Dot)
				if cmdSuccess && (count*request.Len != 0) {
					currentFrame.TextModified = true
					MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkModified])
				}
			}
		}

	case CmdPage:
		if !fromSpan {
			Screen.Message(MsgPaging)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		cmdSuccess = FilePage(currentFrame, &ExitAbort)
		// Clean up the PAGING message
		if !fromSpan {
			Screen.ClearMsgs(false)
		}

	case CmdOpSysCommand:
		var request TParObject
		if TparGet1(currentFrame, tparam, command, &request) {
			var firstLine, lastLine *LineHdrObject
			var ok bool
			if firstLine, lastLine, _, ok = OpsysCommand(&request); !ok {
				return
			}
			if firstLine != nil {
				LinesInject(firstLine, lastLine, currentFrame.Dot.Line)
				MarkCreate(firstLine, 1, &currentFrame.Marks[MarkEquals])
				currentFrame.TextModified = true
				MarkCreate(lastLine.FLink, 1, &currentFrame.Marks[MarkModified])
				MarkCreate(lastLine.FLink, 1, &currentFrame.Dot)
				cmdSuccess = true
			}
		}

	case CmdPositionColumn:
		if count > MaxStrLen {
			return
		}
		currentFrame.Dot.Col = count
		cmdSuccess = true

	case CmdPositionLine:
		newLine := LineFromNumber(currentFrame, count)
		if newLine == nil {
			return
		}
		cmdSuccess = true
		MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkEquals])
		MarkCreate(newLine, 1, &currentFrame.Dot)

	case CmdQuit:
		currentFrame, cmdSuccess = QuitCommand(currentFrame)

	case CmdReplace:
		var request, request2 TParObject
		if TparGet2(currentFrame, tparam, command, &request, &request2) {
			if request.Len == 0 { // If didn't specify, use default
				if currentFrame.Rep1Tpar.Len == 0 {
					Screen.Message(MsgNoDefaultStr)
					return
				}
			} else {
				currentFrame.Rep1Tpar = request // If did specify, save for next time
				currentFrame.Rep2Tpar = request2
				request2.Con = nil
			}
			cmdSuccess = EqsGetRepRep(
				currentFrame,
				rept,
				count,
				currentFrame.Rep1Tpar,
				currentFrame.Rep2Tpar,
				fromSpan,
			)
		}

	case CmdRubout:
		cmdSuccess = CharcmdRubout(currentFrame, rept, count, fromSpan)

	case CmdSetMarginLeft:
		if rept == LeadParamMinus {
			currentFrame.MarginLeft = InitialMarginLeft
		} else {
			if currentFrame.Dot.Col >= currentFrame.MarginRight {
				Screen.Message(MsgLeftMarginGeRight)
				return
			}
			currentFrame.MarginLeft = currentFrame.Dot.Col
		}
		cmdSuccess = true

	case CmdSetMarginRight:
		if rept == LeadParamMinus {
			currentFrame.MarginRight = InitialMarginRight
		} else {
			if currentFrame.Dot.Col <= currentFrame.MarginLeft {
				Screen.Message(MsgLeftMarginGeRight)
				return
			}
			currentFrame.MarginRight = currentFrame.Dot.Col
		}
		cmdSuccess = true

	case CmdSpanJump, CmdSpanCompile, CmdSpanCopy, CmdSpanDefine,
		CmdSpanExecute, CmdSpanExecuteNoRecompile, CmdSpanTransfer:
		var request TParObject
		if TparGet1(currentFrame, tparam, command, &request) {
			newName := request.Str.Slice(1, request.Len)
			switch command {
			case CmdSpanDefine:
				if rept == LeadParamMinus {
					var newSpan, oldSpan *SpanObject
					if SpanFind(newName, &newSpan, &oldSpan) {
						cmdSuccess = SpanDestroy(&newSpan)
					} else {
						Screen.Message(MsgNoSuchSpan)
					}
				} else {
					cmdSuccess = SpanCreate(newName, theMark, currentFrame.Dot)
				}

			case CmdSpanJump:
				var newSpan, oldSpan *SpanObject
				if SpanFind(newName, &newSpan, &oldSpan) {
					var newCol int
					var newLine *LineHdrObject
					if rept == LeadParamMinus {
						newCol = newSpan.MarkOne.Col
						newLine = newSpan.MarkOne.Line
					} else {
						newCol = newSpan.MarkTwo.Col
						newLine = newSpan.MarkTwo.Line
					}
					if newLine.Group.Frame == currentFrame {
						MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &currentFrame.Marks[MarkEquals])
						MarkCreate(newLine, newCol, &currentFrame.Dot)
						cmdSuccess = true
					} else {
						fr := newLine.Group.Frame

						if newFrame, ok := FrameEdit(currentFrame, fr.Span.Name); ok {
							currentFrame = newFrame
							if fr.Marks[MarkEquals] != nil {
								MarkDestroy(&fr.Marks[MarkEquals])
							}
							MarkCreate(newLine, newCol, &fr.Dot)
							cmdSuccess = true
						}
					}
				} else {
					Screen.Message(MsgNoSuchSpan)
				}

			case CmdSpanCopy, CmdSpanTransfer:
				var newSpan, oldSpan *SpanObject
				if SpanFind(newName, &newSpan, &oldSpan) {
					cmdSuccess = TextMove(
						command == CmdSpanCopy,
						count,
						newSpan.MarkOne,
						newSpan.MarkTwo,
						currentFrame.Dot,                // Dest
						&currentFrame.Marks[MarkEquals], // New_Start
						&currentFrame.Dot,               // New_End
					)
					if command == CmdSpanTransfer && newSpan.Frame == nil && cmdSuccess {
						MarkCreate(
							currentFrame.Marks[MarkEquals].Line,
							currentFrame.Marks[MarkEquals].Col,
							&newSpan.MarkOne,
						)
						MarkCreate(currentFrame.Dot.Line, currentFrame.Dot.Col, &newSpan.MarkTwo)
					}
				} else {
					Screen.Message(MsgNoSuchSpan)
				}

			case CmdSpanCompile, CmdSpanExecute, CmdSpanExecuteNoRecompile:
				var newSpan, oldSpan *SpanObject
				if SpanFind(newName, &newSpan, &oldSpan) {
					if newSpan.Code == nil || command != CmdSpanExecuteNoRecompile {
						var ok bool
						currentFrame, ok = CodeCompile(currentFrame, newSpan, true)
						if !ok {
							return
						}
					}
					if command == CmdSpanCompile {
						cmdSuccess = true
					} else {
						currentFrame, cmdSuccess = CodeInterpret(
							currentFrame,
							specialFrames,
							rept,
							count,
							newSpan.Code,
							true,
						)
					}
				} else {
					Screen.Message(MsgNoSuchSpan)
				}
			}
		}

	case CmdSpanIndex:
		cmdSuccess = SpanIndex()

	case CmdSpanAssign:
		var request, request2 TParObject
		if !TparGet2(currentFrame, tparam, command, &request, &request2) {
			return
		}
		if request.Len == 0 {
			return
		}
		newName := request.Str.Slice(1, request.Len)
		var newSpan, oldSpan *SpanObject
		if SpanFind(newName, &newSpan, &oldSpan) {
			// Grunge the old one
			if newSpan == specialFrames.Oops().Span {
				if !TextRemove(specialFrames.Oops().Span.MarkOne, specialFrames.Oops().Span.MarkTwo) {
					return
				}
			} else {
				// Make sure oops_span is okay
				MarkCreate(specialFrames.Oops().LastGroup.LastLine, 1, &specialFrames.Oops().Span.MarkTwo)
				if !TextMove(
					false, // Don't copy, transfer
					1,     // One instance of
					newSpan.MarkOne,
					newSpan.MarkTwo,
					specialFrames.Oops().Span.MarkTwo,       // destination
					&specialFrames.Oops().Marks[MarkEquals], // leave at start
					&specialFrames.Oops().Dot,               // leave at end
				) {
					return
				}
			}
		} else {
			// Create a span in frame "HEAP"
			MarkCreate(specialFrames.Heap().LastGroup.LastLine, 1, &specialFrames.Heap().Span.MarkTwo)
			if !SpanCreate(newName, specialFrames.Heap().Span.MarkTwo, specialFrames.Heap().Span.MarkTwo) {
				return
			}
			if !SpanFind(newName, &newSpan, &oldSpan) {
				return
			}
		}
		// Now copy the tpar into the span
		if !TextInsertTpar(&request2, newSpan.MarkTwo, &newSpan.MarkOne) {
			return
		}
		fr := newSpan.MarkTwo.Line.Group.Frame
		fr.TextModified = true
		MarkCreate(newSpan.MarkTwo.Line, newSpan.MarkTwo.Col, &fr.Marks[MarkModified])
		cmdSuccess = true

	case CmdSplitLine:
		if currentFrame.Dot.Line.FLink == nil {
			TextRealizeNull(currentFrame.Dot.Line)
		}
		cmdSuccess = TextSplitLine(currentFrame.Dot, 0, &currentFrame.Marks[MarkEquals])

	case CmdSwapLine:
		cmdSuccess = SwapLine(currentFrame, rept, count)

	case CmdUserCommandIntroducer:
		if LudwigMode != LudwigScreen {
			Screen.Message(MsgScreenModeOnly)
			return
		}
		cmdSuccess = UserCommandIntroducer(currentFrame)

	case CmdUserKey:
		if LudwigMode != LudwigScreen {
			Screen.Message(MsgScreenModeOnly)
			return
		}
		var request, request2 TParObject
		if TparGet2(currentFrame, tparam, command, &request, &request2) {
			if request.Len == 0 {
				cmdSuccess = false
			} else {
				currentFrame, cmdSuccess = UserKey(currentFrame, specialFrames.Heap(), &request, &request2)
			}
		}

	case CmdUserParent:
		if LudwigMode == LudwigBatch {
			Screen.Message(MsgInteractiveModeOnly)
			return
		}
		cmdSuccess = UserParent()

	case CmdUserSubprocess:
		if LudwigMode == LudwigBatch {
			Screen.Message(MsgInteractiveModeOnly)
			return
		}
		cmdSuccess = UserSubprocess()

	case CmdUserUndo:
		cmdSuccess = UserUndo()

	case CmdWindowBackward, CmdWindowEnd, CmdWindowForward, CmdWindowLeft,
		CmdWindowMiddle, CmdWindowNew, CmdWindowRight, CmdWindowScroll,
		CmdWindowSetHeight, CmdWindowTop, CmdWindowUpdate:
		cmdSuccess = WindowCommand(currentFrame, command, rept, count, fromSpan)

	case CmdResizeWindow:
		Screen.Resize(currentFrame)
		cmdSuccess = true

	case CmdValidate:
		cmdSuccess = ValidateCommand(currentFrame, specialFrames)

	case CmdBlockDefine, CmdBlockTransfer, CmdBlockCopy:
		Screen.Message(MsgNotImplemented)

	default:
		Screen.Message(DbgInternalLogicError)
	}

	if cmdSuccess {
		eqSet := false
		switch CmdAttrib[command].EqAction {
		case EqOld:
			eqSet = true
			MarkCreate(oldDot.Line, oldDot.Col, &oldFrame.Marks[MarkEquals])
		case EqDel:
			eqSet = true
			MarkDestroy(&oldFrame.Marks[MarkEquals])
		case EqNil:
			eqSet = true
		}
		if !eqSet {
			Screen.Message(MsgEqualsNotSet)
		}
	}
	return
}
