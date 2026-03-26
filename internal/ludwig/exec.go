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
func Execute(command Commands, rept LeadParam, count int, tparam *TParObject, fromSpan bool) (cmdSuccess bool) {
	ExecLevel++
	defer func() {
		ExecLevel--
	}()

	if TtControlC {
		return
	}
	if ExecLevel == MaxExecRecursion {
		ScreenMessage(MsgCommandRecursionLimit)
		return
	}

	// Fix commands which use marks without using @ in the syntax
	switch command {
	case CmdMark:
		if count == 0 || iabs(count) > MaxUserMarkNumber {
			ScreenMessage(MsgIllegalMarkNumber)
			return
		}
	case CmdSpanDefine:
		if rept == LeadParamNone || rept == LeadParamPInt {
			if count == 0 || count > MaxUserMarkNumber {
				ScreenMessage(MsgIllegalMarkNumber)
				return
			}
			rept = LeadParamMarker
		}
	}

	// Check the mark, assign theMark to the mark
	var theMark *MarkObject
	if rept == LeadParamMarker {
		theMark = CurrentFrame.Marks[count]
		if theMark == nil {
			ScreenMessage(MsgMarkNotDefined)
			return
		}
	}

	// Save the current value of DOT and CURRENT_FRAME for use by equals
	oldDot := *CurrentFrame.Dot
	oldFrame := CurrentFrame

	// Execute the command
	switch command {
	case CmdAdvance:
		// Establish which line to advance to
		cmdSuccess = (rept == LeadParamPIndef || rept == LeadParamNIndef || rept == LeadParamMarker)
		newLine := CurrentFrame.Dot.Line
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
				newLine = LineFromNumber(CurrentFrame, lineNr+count)
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
				newLine = LineFromNumber(CurrentFrame, lineNr-count)
				if newLine == nil {
					return
				}
			}
			cmdSuccess = true

		case LeadParamPIndef:
			newLine = CurrentFrame.LastGroup.LastLine

		case LeadParamNIndef:
			newLine = CurrentFrame.FirstGroup.FirstLine

		case LeadParamMarker:
			newLine = theMark.Line
		}

		MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkEquals])
		MarkCreate(newLine, 1, &CurrentFrame.Dot)

	case CmdBridge, CmdNext:
		var request TParObject
		if TparGet1(tparam, command, &request) {
			cmdSuccess = NextbridgeCommand(CurrentFrame, count, &request, command == CmdBridge)
		}

	case CmdCaseEdit, CmdCaseLow, CmdCaseUp, CmdDittoDown, CmdDittoUp:
		cmdSuccess = CaseDittoCommand(CurrentFrame, command, rept, count, fromSpan)

	case CmdDeleteChar:
		if rept != LeadParamMarker {
			cmdSuccess = CharcmdDelete(CurrentFrame, rept, count, fromSpan)
		} else {
			theOtherMark := CurrentFrame.Dot
			lineNr := LineToNumber(CurrentFrame.Dot.Line)
			line2Nr := LineToNumber(theMark.Line)
			if lineNr > line2Nr || (lineNr == line2Nr && CurrentFrame.Dot.Col > theMark.Col) {
				// Reverse mark pointers to get theOtherMark first
				anotherMark := theMark
				theMark = theOtherMark
				theOtherMark = anotherMark
			}
			if CurrentFrame != FrameOops {
				// Make sure oops_span is okay
				MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Span.MarkTwo)
				cmdSuccess = TextMove(
					false,                        // Don't copy, transfer
					1,                            // One instance of
					theOtherMark,                 // starting pos
					theMark,                      // ending pos
					FrameOops.Span.MarkTwo,       // destination
					&FrameOops.Marks[MarkEquals], // leave at start
					&FrameOops.Dot,               // leave at end
				)
				FrameOops.TextModified = true
				MarkCreate(FrameOops.Dot.Line, FrameOops.Dot.Col, &FrameOops.Marks[MarkModified])
			} else {
				cmdSuccess = TextRemove(theOtherMark, theMark)
			}
			CurrentFrame.TextModified = true
			MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified])
		}

	case CmdDeleteLine:
		// Establish which lines to kill, this is common to K and FW cmds
		var firstLine, lastLine *LineHdrObject
		if !ExecComputeLineRange(CurrentFrame, rept, count, &firstLine, &lastLine) {
			return
		}
		if firstLine != nil {
			dotCol := CurrentFrame.Dot.Col
			if lastLine.FLink == nil {
				return
			}
			MarksSqueeze(firstLine, 1, lastLine.FLink, 1)
			LinesExtract(firstLine, lastLine)
			if CurrentFrame != FrameOops {
				LinesInject(firstLine, lastLine, FrameOops.LastGroup.LastLine)
				MarkCreate(firstLine, 1, &FrameOops.Marks[MarkEquals])
				MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Dot)
				FrameOops.TextModified = true
				MarkCreate(FrameOops.Dot.Line, FrameOops.Dot.Col, &FrameOops.Marks[MarkModified])
			}
			CurrentFrame.Dot.Col = dotCol
			CurrentFrame.TextModified = true
			MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified])
		}
		cmdSuccess = true

	case CmdBacktab, CmdDown, CmdHome, CmdLeft, CmdReturn, CmdRight, CmdTab, CmdUp:
		if command == CmdReturn && EditMode == ModeInsert &&
			CurrentFrame.Options.Has(OptNewLine) {
			if CurrentFrame.Dot.Line.FLink == nil {
				TextRealizeNull(CurrentFrame.Dot.Line)
				cmdSuccess = ArrowCommand(CurrentFrame, command, rept, count, fromSpan)
			} else {
				cmdSuccess = Execute(CmdSplitLine, rept, count, tparam, fromSpan)
			}
		} else {
			cmdSuccess = ArrowCommand(CurrentFrame, command, rept, count, fromSpan)
		}

	case CmdDump:
		// DEBUG command - skip in release build

	case CmdEqualColumn:
		var request TParObject
		i := 1 // Start of column number, j receives column number
		if TparGet1(tparam, command, &request) {
			if j, found := TparToInt(&request, &i); found {
				switch rept {
				case LeadParamNone, LeadParamPlus:
					cmdSuccess = (CurrentFrame.Dot.Col == j)
				case LeadParamMinus:
					cmdSuccess = (CurrentFrame.Dot.Col != j)
				case LeadParamPIndef:
					cmdSuccess = (CurrentFrame.Dot.Col >= j)
				case LeadParamNIndef:
					cmdSuccess = (CurrentFrame.Dot.Col <= j)
				}
			}
		}

	case CmdEqualEol:
		switch rept {
		case LeadParamNone, LeadParamPlus:
			cmdSuccess = (CurrentFrame.Dot.Col == CurrentFrame.Dot.Line.Used+1)
		case LeadParamMinus:
			cmdSuccess = (CurrentFrame.Dot.Col != CurrentFrame.Dot.Line.Used+1)
		case LeadParamPIndef:
			cmdSuccess = (CurrentFrame.Dot.Col >= CurrentFrame.Dot.Line.Used+1)
		case LeadParamNIndef:
			cmdSuccess = (CurrentFrame.Dot.Col <= CurrentFrame.Dot.Line.Used+1)
		}

	case CmdEqualEop, CmdEqualEof:
		cmdSuccess = (CurrentFrame.Dot.Line.FLink == nil)
		if command == CmdEqualEof {
			if CurrentFrame.InputFile != 0 {
				if !Files[CurrentFrame.InputFile].Eof {
					cmdSuccess = false
				}
			}
		}
		if rept == LeadParamMinus {
			cmdSuccess = !cmdSuccess
		}

	case CmdEqualMark:
		var request TParObject
		if !TparGet1(tparam, command, &request) {
			return
		}
		var found bool
		var j int
		if j, found = TparToMark(&request); !found {
			return
		}
		if CurrentFrame.Marks[j] != nil {
			switch rept {
			case LeadParamNone, LeadParamPlus, LeadParamMinus:
				if CurrentFrame.Marks[j].Line == CurrentFrame.Dot.Line &&
					CurrentFrame.Marks[j].Col == CurrentFrame.Dot.Col {
					cmdSuccess = true
				}
				if rept == LeadParamMinus {
					cmdSuccess = !cmdSuccess
				}
			case LeadParamPIndef, LeadParamNIndef:
				if CurrentFrame.Marks[j].Line == CurrentFrame.Dot.Line {
					if rept == LeadParamPIndef {
						cmdSuccess = (CurrentFrame.Dot.Col >= CurrentFrame.Marks[j].Col)
					} else {
						cmdSuccess = (CurrentFrame.Dot.Col <= CurrentFrame.Marks[j].Col)
					}
				} else {
					lineNr := LineToNumber(CurrentFrame.Dot.Line)
					line2Nr := LineToNumber(CurrentFrame.Marks[j].Line)
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
		if TparGet1(tparam, command, &request) {
			if request.Len == 0 {
				// If didn't specify, use default
				request = CurrentFrame.EqsTpar
				if request.Len == 0 {
					ScreenMessage(MsgNoDefaultStr)
					return
				}
			} else {
				CurrentFrame.EqsTpar = request // If did specify, save for next time
			}
		}
		cmdSuccess = EqsGetRepEqs(rept, request)

	case CmdDoLastCommand, CmdExecuteString:
		if CurrentFrame == FrameCmd {
			ScreenMessage(MsgNotWhileEditingCmd)
			return
		}
		if command == CmdExecuteString {
			var request TParObject
			if !TparGet1(tparam, command, &request) {
				return
			}

			FrameCmd.ReturnFrame = CurrentFrame
			CurrentFrame = FrameCmd

			// Zap frame COMMAND's current contents
			firstLine := FrameCmd.FirstGroup.FirstLine
			lastLine := FrameCmd.LastGroup.LastLine.BLink
			if lastLine != nil {
				MarksSqueeze(firstLine, 1, lastLine.FLink, 1)
				LinesExtract(firstLine, lastLine)
			}

			// Insert the new tpar into frame COMMAND
			if !TextInsertTpar(&request, FrameCmd.Dot, &FrameCmd.Marks[MarkEquals]) {
				return
			}

			CurrentFrame = CurrentFrame.ReturnFrame
		}

		// Recompile and execute frame COMMAND
		var newSpan, oldSpan *SpanObject
		if SpanFind(FrameCmd.Span.Name, &newSpan, &oldSpan) {
			if !CodeCompile(FrameCmd.Span, true) {
				return
			}
			cmdSuccess = CodeInterpret(rept, count, FrameCmd.Span.Code, true)
		}

	case CmdFileInput, CmdFileOutput, CmdFileEdit, CmdFileRead, CmdFileWrite,
		CmdFileRewind, CmdFileKill, CmdFileSave,
		CmdFileGlobalInput, CmdFileGlobalOutput, CmdFileGlobalRewind, CmdFileGlobalKill:
		cmdSuccess = FileCommand(CurrentFrame, command, rept, count, tparam, fromSpan)

	case CmdFileExecute:
		if CurrentFrame == FrameCmd {
			ScreenMessage(MsgNotWhileEditingCmd)
			return
		}
		var request TParObject
		if TparGet1(tparam, command, &request) {
			newTparam := request
			FrameCmd.ReturnFrame = CurrentFrame
			CurrentFrame = FrameCmd
			// Zap frame COMMAND's current contents
			firstLine := FrameCmd.FirstGroup.FirstLine
			lastLine := FrameCmd.LastGroup.LastLine.BLink
			if lastLine != nil {
				MarksSqueeze(firstLine, 1, lastLine.FLink, 1)
				LinesExtract(firstLine, lastLine)
			}
			if FileCommand(CurrentFrame, CmdFileExecute, LeadParamNone, 0, &newTparam, false) {
				CurrentFrame = CurrentFrame.ReturnFrame
				// Recompile and execute frame COMMAND
				var newSpan, oldSpan *SpanObject
				if SpanFind(FrameCmd.Span.Name, &newSpan, &oldSpan) {
					if CodeCompile(FrameCmd.Span, true) {
						cmdSuccess = CodeInterpret(rept, count, FrameCmd.Span.Code, true)
					}
				}
			} else {
				CurrentFrame = CurrentFrame.ReturnFrame
			}
		}

	case CmdFileTable:
		FileTable()
		cmdSuccess = true

	case CmdFrameEdit:
		var request TParObject
		if TparGet1(tparam, command, &request) {
			newName := request.Str.Slice(1, request.Len)
			if newFrame, ok := FrameEdit(CurrentFrame, newName); ok {
				CurrentFrame = newFrame
				cmdSuccess = true
			} else {
				cmdSuccess = false
			}
		}

	case CmdFrameKill:
		var request TParObject
		if TparGet1(tparam, command, &request) {
			newName := request.Str.Slice(1, request.Len)
			cmdSuccess = FrameKill(CurrentFrame, newName)
		}

	case CmdFrameParameters:
		cmdSuccess = FrameParameter(CurrentFrame, tparam)

	case CmdFrameReturn:
		for i := 1; i <= count; i++ {
			if CurrentFrame.ReturnFrame == nil {
				CurrentFrame = oldFrame
				return
			}
			CurrentFrame = CurrentFrame.ReturnFrame
		}
		cmdSuccess = true

	case CmdGet:
		var request TParObject
		if TparGet1(tparam, command, &request) {
			if request.Len == 0 {
				// If didn't specify, use default
				request = CurrentFrame.GetTpar
				if request.Len == 0 {
					ScreenMessage(MsgNoDefaultStr)
					return
				}
			} else {
				CurrentFrame.GetTpar = request // If did specify, save for next time
			}
			cmdSuccess = EqsGetRepGet(count, request, fromSpan)
		}

	case CmdHelp:
		if LudwigMode == LudwigBatch {
			ScreenMessage(MsgInteractiveModeOnly)
			return
		}
		var request TParObject
		if TparGet1(tparam, command, &request) {
			HelpHelp(string(request.Str.Slice(1, request.Len)))
			cmdSuccess = true // Never fails
		}

	case CmdInsertChar:
		cmdSuccess = CharcmdInsert(CurrentFrame, rept, count, fromSpan)

	case CmdInsertLine:
		if count != 0 {
			firstLine, lastLine := LinesCreate(iabs(count))
			LinesInject(firstLine, lastLine, CurrentFrame.Dot.Line)
			if count > 0 {
				MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkEquals])
				MarkCreate(firstLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot)
			} else {
				MarkCreate(firstLine, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkEquals])
			}
			CurrentFrame.TextModified = true
			MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified])
		} else {
			MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkEquals])
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
				ScreenMessage(MsgSyntaxError)
			}
		} else {
			var request TParObject
			if TparGet1(tparam, command, &request) {
				if request.Con == nil {
					cmdSuccess = TextInsert(true, count, request.Str, request.Len, CurrentFrame.Dot)
					if cmdSuccess && (count*request.Len != 0) {
						CurrentFrame.TextModified = true
						MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified])
					}
				} else {
					for i := 1; i <= count; i++ {
						if !TextInsertTpar(&request, CurrentFrame.Dot, &CurrentFrame.Marks[MarkEquals]) {
							return
						}
					}
					CurrentFrame.TextModified = true
					MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified])
					cmdSuccess = true
				}
			}
		}

	case CmdInsertInvisible:
		if LudwigMode != LudwigScreen {
			return
		}
		var i int
		if CurrentFrame.Dot.Col > CurrentFrame.Dot.Line.Used {
			i = MaxStrLenP - CurrentFrame.Dot.Col
		} else {
			i = MaxStrLen - CurrentFrame.Dot.Line.Used
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
		cmdSuccess = TextInsert(true, 1, newStr, count, CurrentFrame.Dot)
		if cmdSuccess && count != 0 {
			CurrentFrame.TextModified = true
			MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified])
		}

	case CmdJump:
		switch rept {
		case LeadParamNone, LeadParamPlus, LeadParamPInt:
			if CurrentFrame.Dot.Col+count > MaxStrLenP {
				return
			}
		case LeadParamMinus, LeadParamNInt:
			if CurrentFrame.Dot.Col <= -count {
				return
			}
		case LeadParamPIndef:
			if CurrentFrame.Dot.Col > CurrentFrame.Dot.Line.Used+1 {
				return
			}
			count = 1 + CurrentFrame.Dot.Line.Used - CurrentFrame.Dot.Col
		case LeadParamNIndef:
			count = 1 - CurrentFrame.Dot.Col
		case LeadParamMarker:
			MarkCreate(theMark.Line, theMark.Col, &CurrentFrame.Dot)
			count = 0
		}
		CurrentFrame.Dot.Col += count
		cmdSuccess = true

	case CmdLineCentre:
		cmdSuccess = WordCentre(CurrentFrame, rept, count)

	case CmdLineFill:
		cmdSuccess = WordFill(CurrentFrame, rept, count)

	case CmdLineJustify:
		cmdSuccess = WordJustify(CurrentFrame, rept, count)

	case CmdLineSquash:
		cmdSuccess = WordSqueeze(CurrentFrame, rept, count)

	case CmdLineLeft:
		cmdSuccess = WordLeft(CurrentFrame, rept, count)

	case CmdLineRight:
		cmdSuccess = WordRight(CurrentFrame, rept, count)

	case CmdWordAdvance:
		if FileData.OldCmds {
			cmdSuccess = WordAdvanceWord(CurrentFrame, rept, count)
		} else {
			cmdSuccess = NewwordAdvanceWord(CurrentFrame, rept, count)
		}

	case CmdWordDelete:
		if FileData.OldCmds {
			cmdSuccess = WordDeleteWord(CurrentFrame, rept, count)
		} else {
			cmdSuccess = NewwordDeleteWord(CurrentFrame, rept, count)
		}

	case CmdAdvanceParagraph:
		cmdSuccess = NewwordAdvanceParagraph(CurrentFrame, rept, count)

	case CmdDeleteParagraph:
		cmdSuccess = NewwordDeleteParagraph(CurrentFrame, rept, count)

	case CmdMark:
		cmdSuccess = true
		if count < 0 {
			MarkDestroy(&CurrentFrame.Marks[-count])
		} else {
			MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[count])
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
					ScreenMessage(MsgScreenModeOnly)
					return
				}
				if !UserCommandIntroducer(CurrentFrame) {
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
				ScreenMessage(MsgSyntaxError)
			}
		} else {
			var request TParObject
			if TparGet1(tparam, command, &request) {
				cmdSuccess = TextOvertype(true, count, request.Str, request.Len, CurrentFrame.Dot)
				if cmdSuccess && (count*request.Len != 0) {
					CurrentFrame.TextModified = true
					MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified])
				}
			}
		}

	case CmdPage:
		if !fromSpan {
			ScreenMessage(MsgPaging)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		cmdSuccess = FilePage(CurrentFrame, &ExitAbort)
		// Clean up the PAGING message
		if !fromSpan {
			ScreenClearMsgs(false)
		}

	case CmdOpSysCommand:
		var request TParObject
		if TparGet1(tparam, command, &request) {
			var firstLine, lastLine *LineHdrObject
			var ok bool
			if firstLine, lastLine, _, ok = OpsysCommand(&request); !ok {
				return
			}
			if firstLine != nil {
				LinesInject(firstLine, lastLine, CurrentFrame.Dot.Line)
				MarkCreate(firstLine, 1, &CurrentFrame.Marks[MarkEquals])
				CurrentFrame.TextModified = true
				MarkCreate(lastLine.FLink, 1, &CurrentFrame.Marks[MarkModified])
				MarkCreate(lastLine.FLink, 1, &CurrentFrame.Dot)
				cmdSuccess = true
			}
		}

	case CmdPositionColumn:
		if count > MaxStrLen {
			return
		}
		CurrentFrame.Dot.Col = count
		cmdSuccess = true

	case CmdPositionLine:
		newLine := LineFromNumber(CurrentFrame, count)
		if newLine == nil {
			return
		}
		cmdSuccess = true
		MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkEquals])
		MarkCreate(newLine, 1, &CurrentFrame.Dot)

	case CmdQuit:
		cmdSuccess = QuitCommand()

	case CmdReplace:
		var request, request2 TParObject
		if TparGet2(tparam, command, &request, &request2) {
			if request.Len == 0 { // If didn't specify, use default
				if CurrentFrame.Rep1Tpar.Len == 0 {
					ScreenMessage(MsgNoDefaultStr)
					return
				}
			} else {
				CurrentFrame.Rep1Tpar = request // If did specify, save for next time
				CurrentFrame.Rep2Tpar = request2
				request2.Con = nil
			}
			cmdSuccess = EqsGetRepRep(rept, count, CurrentFrame.Rep1Tpar,
				CurrentFrame.Rep2Tpar, fromSpan)
		}

	case CmdRubout:
		cmdSuccess = CharcmdRubout(CurrentFrame, rept, count, fromSpan)

	case CmdSetMarginLeft:
		if rept == LeadParamMinus {
			CurrentFrame.MarginLeft = InitialMarginLeft
		} else {
			if CurrentFrame.Dot.Col >= CurrentFrame.MarginRight {
				ScreenMessage(MsgLeftMarginGeRight)
				return
			}
			CurrentFrame.MarginLeft = CurrentFrame.Dot.Col
		}
		cmdSuccess = true

	case CmdSetMarginRight:
		if rept == LeadParamMinus {
			CurrentFrame.MarginRight = InitialMarginRight
		} else {
			if CurrentFrame.Dot.Col <= CurrentFrame.MarginLeft {
				ScreenMessage(MsgLeftMarginGeRight)
				return
			}
			CurrentFrame.MarginRight = CurrentFrame.Dot.Col
		}
		cmdSuccess = true

	case CmdSpanJump, CmdSpanCompile, CmdSpanCopy, CmdSpanDefine,
		CmdSpanExecute, CmdSpanExecuteNoRecompile, CmdSpanTransfer:
		var request TParObject
		if TparGet1(tparam, command, &request) {
			newName := request.Str.Slice(1, request.Len)
			switch command {
			case CmdSpanDefine:
				if rept == LeadParamMinus {
					var newSpan, oldSpan *SpanObject
					if SpanFind(newName, &newSpan, &oldSpan) {
						cmdSuccess = SpanDestroy(&newSpan)
					} else {
						ScreenMessage(MsgNoSuchSpan)
					}
				} else {
					cmdSuccess = SpanCreate(newName, theMark, CurrentFrame.Dot)
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
					if newLine.Group.Frame == CurrentFrame {
						MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkEquals])
						MarkCreate(newLine, newCol, &CurrentFrame.Dot)
						cmdSuccess = true
					} else {
						fr := newLine.Group.Frame

						if newFrame, ok := FrameEdit(CurrentFrame, fr.Span.Name); ok {
							CurrentFrame = newFrame
							if fr.Marks[MarkEquals] != nil {
								MarkDestroy(&fr.Marks[MarkEquals])
							}
							MarkCreate(newLine, newCol, &fr.Dot)
							cmdSuccess = true
						}
					}
				} else {
					ScreenMessage(MsgNoSuchSpan)
				}

			case CmdSpanCopy, CmdSpanTransfer:
				var newSpan, oldSpan *SpanObject
				if SpanFind(newName, &newSpan, &oldSpan) {
					cmdSuccess = TextMove(
						command == CmdSpanCopy,
						count,
						newSpan.MarkOne,
						newSpan.MarkTwo,
						CurrentFrame.Dot,                // Dest
						&CurrentFrame.Marks[MarkEquals], // New_Start
						&CurrentFrame.Dot,               // New_End
					)
					if command == CmdSpanTransfer && newSpan.Frame == nil && cmdSuccess {
						MarkCreate(
							CurrentFrame.Marks[MarkEquals].Line,
							CurrentFrame.Marks[MarkEquals].Col,
							&newSpan.MarkOne,
						)
						MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &newSpan.MarkTwo)
					}
				} else {
					ScreenMessage(MsgNoSuchSpan)
				}

			case CmdSpanCompile, CmdSpanExecute, CmdSpanExecuteNoRecompile:
				var newSpan, oldSpan *SpanObject
				if SpanFind(newName, &newSpan, &oldSpan) {
					if newSpan.Code == nil || command != CmdSpanExecuteNoRecompile {
						if !CodeCompile(newSpan, true) {
							return
						}
					}
					if command == CmdSpanCompile {
						cmdSuccess = true
					} else {
						cmdSuccess = CodeInterpret(rept, count, newSpan.Code, true)
					}
				} else {
					ScreenMessage(MsgNoSuchSpan)
				}
			}
		}

	case CmdSpanIndex:
		cmdSuccess = SpanIndex()

	case CmdSpanAssign:
		var request, request2 TParObject
		if !TparGet2(tparam, command, &request, &request2) {
			return
		}
		if request.Len == 0 {
			return
		}
		newName := request.Str.Slice(1, request.Len)
		var newSpan, oldSpan *SpanObject
		if SpanFind(newName, &newSpan, &oldSpan) {
			// Grunge the old one
			if newSpan == FrameOops.Span {
				if !TextRemove(FrameOops.Span.MarkOne, FrameOops.Span.MarkTwo) {
					return
				}
			} else {
				// Make sure oops_span is okay
				MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Span.MarkTwo)
				if !TextMove(
					false, // Don't copy, transfer
					1,     // One instance of
					newSpan.MarkOne,
					newSpan.MarkTwo,
					FrameOops.Span.MarkTwo,       // destination
					&FrameOops.Marks[MarkEquals], // leave at start
					&FrameOops.Dot,               // leave at end
				) {
					return
				}
			}
		} else {
			// Create a span in frame "HEAP"
			MarkCreate(FrameHeap.LastGroup.LastLine, 1, &FrameHeap.Span.MarkTwo)
			if !SpanCreate(newName, FrameHeap.Span.MarkTwo, FrameHeap.Span.MarkTwo) {
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
		if CurrentFrame.Dot.Line.FLink == nil {
			TextRealizeNull(CurrentFrame.Dot.Line)
		}
		cmdSuccess = TextSplitLine(CurrentFrame.Dot, 0, &CurrentFrame.Marks[MarkEquals])

	case CmdSwapLine:
		cmdSuccess = SwapLine(CurrentFrame, rept, count)

	case CmdUserCommandIntroducer:
		if LudwigMode != LudwigScreen {
			ScreenMessage(MsgScreenModeOnly)
			return
		}
		cmdSuccess = UserCommandIntroducer(CurrentFrame)

	case CmdUserKey:
		if LudwigMode != LudwigScreen {
			ScreenMessage(MsgScreenModeOnly)
			return
		}
		var request, request2 TParObject
		if TparGet2(tparam, command, &request, &request2) {
			if request.Len == 0 {
				cmdSuccess = false
			} else {
				cmdSuccess = UserKey(&request, &request2)
			}
		}

	case CmdUserParent:
		if LudwigMode == LudwigBatch {
			ScreenMessage(MsgInteractiveModeOnly)
			return
		}
		cmdSuccess = UserParent()

	case CmdUserSubprocess:
		if LudwigMode == LudwigBatch {
			ScreenMessage(MsgInteractiveModeOnly)
			return
		}
		cmdSuccess = UserSubprocess()

	case CmdUserUndo:
		cmdSuccess = UserUndo()

	case CmdWindowBackward, CmdWindowEnd, CmdWindowForward, CmdWindowLeft,
		CmdWindowMiddle, CmdWindowNew, CmdWindowRight, CmdWindowScroll,
		CmdWindowSetHeight, CmdWindowTop, CmdWindowUpdate:
		cmdSuccess = WindowCommand(CurrentFrame, command, rept, count, fromSpan)

	case CmdResizeWindow:
		ScreenResize()
		cmdSuccess = true

	case CmdValidate:
		// DEBUG command - skip in release build

	case CmdBlockDefine, CmdBlockTransfer, CmdBlockCopy:
		ScreenMessage(MsgNotImplemented)

	default:
		ScreenMessage(DbgInternalLogicError)
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
			ScreenMessage(MsgEqualsNotSet)
		}
	}
	return
}
