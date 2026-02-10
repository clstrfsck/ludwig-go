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
	result := false
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
					goto l99
				}
			}
			if (*lastLine).FLink == nil {
				goto l99
			}
		} else {
			var lineNr int
			if !LineToNumber(*firstLine, &lineNr) {
				goto l99
			}
			// FIXME: Should this be "frame" rather than "CurrentFrame"
			if !LineFromNumber(CurrentFrame, lineNr+count-1, lastLine) {
				goto l99
			}
			if *lastLine == nil {
				goto l99
			}
			if (*lastLine).FLink == nil {
				goto l99
			}
		}

	case LeadParamMinus, LeadParamNInt:
		count = -count
		*lastLine = frame.Dot.Line.BLink
		if *lastLine == nil {
			goto l99
		}
		if count <= 20 {
			for lineNr := 1; lineNr <= count; lineNr++ {
				*firstLine = (*firstLine).BLink
				if *firstLine == nil {
					goto l99
				}
			}
		} else {
			var lineNr int
			if !LineToNumber(*lastLine, &lineNr) {
				goto l99
			}
			if count > lineNr {
				goto l99
			}
			lineNr = lineNr - count + 1
			// FIXME: Not sure if "CurrentFrame" should be "frame"
			if !LineFromNumber(CurrentFrame, lineNr, firstLine) {
				goto l99
			}
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
		markLine := frame.Marks[count-MinMarkNumber].Line
		if markLine == *firstLine { // TRY TO OPTIMIZE MOST COMMON CASES
			*firstLine = nil
		} else if markLine.FLink == *firstLine {
			*firstLine = markLine
			*lastLine = markLine
		} else if markLine.BLink == *firstLine {
			*lastLine = *firstLine
		} else { // DO IT THE HARD WAY!
			var markLineNr int
			if !LineToNumber(markLine, &markLineNr) {
				goto l99
			}
			var lineNr int
			if !LineToNumber(frame.Dot.Line, &lineNr) {
				goto l99
			}
			if markLineNr < lineNr {
				*firstLine = markLine
				*lastLine = (*lastLine).BLink
			} else {
				*lastLine = markLine.BLink
			}
		}
	}

	result = true
l99:
	return result
}

// Execute executes a command with the specified parameters
func Execute(command Commands, rept LeadParam, count int, tparam *TParObject, fromSpan bool) bool {
	var cmdSuccess bool
	var newCol int
	var dotCol int
	var newLine *LineHdrObject
	var firstLine *LineHdrObject
	var lastLine *LineHdrObject
	var key int
	var i, j int
	var lineNr, line2Nr int
	var newName string
	var newSpan, oldSpan *SpanObject
	var request, request2 TParObject
	var theMark *MarkObject
	var theOtherMark *MarkObject
	var anotherMark *MarkObject
	var eqSet bool            // These 3 are used for
	var oldFrame *FrameObject // the setting up of
	var oldDot MarkObject     // the commands = behaviour
	var newStr *StrObject

	cmdSuccess = false
	request.Nxt = nil
	request.Con = nil
	request2.Nxt = nil
	request2.Con = nil
	ExecLevel++
	if TtControlC {
		goto l99
	}
	if ExecLevel == MaxExecRecursion {
		ScreenMessage(MsgCommandRecursionLimit)
		goto l99
	}

	// Fix commands which use marks without using @ in the syntax
	switch command {
	case CmdMark:
		if count == 0 || int(math.Abs(float64(count))) > MaxUserMarkNumber {
			ScreenMessage(MsgIllegalMarkNumber)
			goto l99
		}
	case CmdSpanDefine:
		if rept == LeadParamNone || rept == LeadParamPInt {
			if count == 0 || count > MaxUserMarkNumber {
				ScreenMessage(MsgIllegalMarkNumber)
				goto l99
			}
			rept = LeadParamMarker
		}
	}

	// Check the mark, assign theMark to the mark
	if rept == LeadParamMarker {
		theMark = CurrentFrame.Marks[count]
		if theMark == nil {
			ScreenMessage(MsgMarkNotDefined)
			goto l99
		}
	}

	// Save the current value of DOT and CURRENT_FRAME for use by equals
	oldDot = *CurrentFrame.Dot
	oldFrame = CurrentFrame

	// Execute the command
	switch command {
	case CmdAdvance:
		// Establish which line to advance to
		cmdSuccess = (rept == LeadParamPIndef || rept == LeadParamNIndef || rept == LeadParamMarker)
		newLine = CurrentFrame.Dot.Line
		switch rept {
		case LeadParamNone, LeadParamPlus, LeadParamPInt:
			if count < 20 {
				for count > 0 {
					count--
					newLine = newLine.FLink
					if newLine == nil {
						goto l99
					}
				}
			} else {
				if !LineToNumber(newLine, &lineNr) {
					goto l99
				}
				if !LineFromNumber(CurrentFrame, lineNr+count, &newLine) {
					goto l99
				}
				if newLine == nil {
					goto l99
				}
			}
			// if flink is nil we are on eop-line, so fail
			if newLine.FLink == nil {
				goto l99
			}
			cmdSuccess = true

		case LeadParamMinus, LeadParamNInt:
			count = -count
			if count < 20 {
				for count > 0 {
					count--
					newLine = newLine.BLink
					if newLine == nil {
						goto l99
					}
				}
			} else {
				if !LineToNumber(newLine, &lineNr) {
					goto l99
				}
				if count >= lineNr {
					goto l99
				}
				if !LineFromNumber(CurrentFrame, lineNr-count, &newLine) {
					goto l99
				}
				if newLine == nil {
					goto l99
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

		if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
			goto l99
		}
		MarkCreate(newLine, 1, &CurrentFrame.Dot)

	case CmdBridge, CmdNext:
		if TparGet1(tparam, command, &request) {
			cmdSuccess = NextbridgeCommand(count, &request, command == CmdBridge)
		}

	case CmdCaseEdit, CmdCaseLow, CmdCaseUp, CmdDittoDown, CmdDittoUp:
		cmdSuccess = CaseDittoCommand(command, rept, count, fromSpan)

	case CmdDeleteChar:
		if rept != LeadParamMarker {
			cmdSuccess = CharcmdDelete(command, rept, count, fromSpan)
		} else {
			theOtherMark = CurrentFrame.Dot
			if !LineToNumber(CurrentFrame.Dot.Line, &lineNr) {
				goto l99
			}
			if !LineToNumber(theMark.Line, &line2Nr) {
				goto l99
			}
			if lineNr > line2Nr || (lineNr == line2Nr && CurrentFrame.Dot.Col > theMark.Col) {
				// Reverse mark pointers to get theOtherMark first
				anotherMark = theMark
				theMark = theOtherMark
				theOtherMark = anotherMark
			}
			if CurrentFrame != FrameOops {
				// Make sure oops_span is okay
				if !MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Span.MarkTwo) {
					goto l99
				}
				cmdSuccess = TextMove(
					false,                  // Don't copy, transfer
					1,                      // One instance of
					theOtherMark,           // starting pos
					theMark,                // ending pos
					FrameOops.Span.MarkTwo, // destination
					&FrameOops.Marks[MarkEquals-MinMarkNumber], // leave at start
					&FrameOops.Dot, // leave at end
				)
				FrameOops.TextModified = true
				MarkCreate(FrameOops.Dot.Line, FrameOops.Dot.Col, &FrameOops.Marks[MarkModified-MinMarkNumber])
			} else {
				cmdSuccess = TextRemove(theOtherMark, theMark)
			}
			CurrentFrame.TextModified = true
			MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified-MinMarkNumber])
		}

	case CmdDeleteLine:
		// Establish which lines to kill, this is common to K and FW cmds
		if !ExecComputeLineRange(CurrentFrame, rept, count, &firstLine, &lastLine) {
			goto l99
		}
		if firstLine != nil {
			dotCol = CurrentFrame.Dot.Col
			if lastLine.FLink == nil {
				goto l99
			}
			if !MarksSqueeze(firstLine, 1, lastLine.FLink, 1) {
				goto l99
			}
			if !LinesExtract(firstLine, lastLine) {
				goto l99
			}
			if CurrentFrame != FrameOops {
				if !LinesInject(firstLine, lastLine, FrameOops.LastGroup.LastLine) {
					goto l99
				}
				if !MarkCreate(firstLine, 1, &FrameOops.Marks[MarkEquals-MinMarkNumber]) {
					goto l99
				}
				if !MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Dot) {
					goto l99
				}
				FrameOops.TextModified = true
				MarkCreate(FrameOops.Dot.Line, FrameOops.Dot.Col, &FrameOops.Marks[MarkModified-MinMarkNumber])
			} else if !LinesDestroy(&firstLine, &lastLine) {
				goto l99
			}
			CurrentFrame.Dot.Col = dotCol
			CurrentFrame.TextModified = true
			MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified-MinMarkNumber])
		}
		cmdSuccess = true

	case CmdBacktab, CmdDown, CmdHome, CmdLeft, CmdReturn, CmdRight, CmdTab, CmdUp:
		if command == CmdReturn && EditMode == ModeInsert &&
			CurrentFrame.Options.Has(OptNewLine) {
			if CurrentFrame.Dot.Line.FLink == nil {
				cmdSuccess = TextRealizeNull(CurrentFrame.Dot.Line)
				cmdSuccess = ArrowCommand(command, rept, count, fromSpan)
			} else {
				cmdSuccess = Execute(CmdSplitLine, rept, count, tparam, fromSpan)
			}
		} else {
			cmdSuccess = ArrowCommand(command, rept, count, fromSpan)
		}

	case CmdDump:
		// DEBUG command - skip in release build

	case CmdEqualColumn:
		i = 1 // Start of column number, j receives column number
		if TparGet1(tparam, command, &request) {
			if TparToInt(&request, &i, &j) {
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
		if !TparGet1(tparam, command, &request) {
			goto l99
		}
		if !TparToMark(&request, &j) {
			goto l99
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
				} else if LineToNumber(CurrentFrame.Dot.Line, &lineNr) &&
					LineToNumber(CurrentFrame.Marks[j].Line, &line2Nr) {
					if rept == LeadParamPIndef {
						cmdSuccess = (lineNr >= line2Nr)
					} else {
						cmdSuccess = (lineNr <= line2Nr)
					}
				}
			}
		}

	case CmdEqualString:
		if TparGet1(tparam, command, &request) {
			if request.Len == 0 {
				// If didn't specify, use default
				request = CurrentFrame.EqsTpar
				if request.Len == 0 {
					ScreenMessage(MsgNoDefaultStr)
					goto l99
				}
			} else {
				CurrentFrame.EqsTpar = request // If did specify, save for next time
			}
		}
		cmdSuccess = EqsGetRepEqs(rept, request)

	case CmdDoLastCommand, CmdExecuteString:
		if CurrentFrame == FrameCmd {
			ScreenMessage(MsgNotWhileEditingCmd)
			goto l99
		}
		if command == CmdExecuteString {
			if !TparGet1(tparam, command, &request) {
				goto l99
			}

			FrameCmd.ReturnFrame = CurrentFrame
			CurrentFrame = FrameCmd

			// Zap frame COMMAND's current contents
			firstLine = FrameCmd.FirstGroup.FirstLine
			lastLine = FrameCmd.LastGroup.LastLine.BLink
			if lastLine != nil {
				if !MarksSqueeze(firstLine, 1, lastLine.FLink, 1) {
					goto l99
				}
				if !LinesExtract(firstLine, lastLine) {
					goto l99
				}
				if !LinesDestroy(&firstLine, &lastLine) {
					goto l99
				}
			}

			// Insert the new tpar into frame COMMAND
			if !TextInsertTpar(&request, FrameCmd.Dot, &FrameCmd.Marks[MarkEquals-MinMarkNumber]) {
				goto l99
			}

			CurrentFrame = CurrentFrame.ReturnFrame
		}

		// Recompile and execute frame COMMAND
		if SpanFind(FrameCmd.Span.Name, &newSpan, &oldSpan) {
			if !CodeCompile(FrameCmd.Span, true) {
				goto l99
			}
			cmdSuccess = CodeInterpret(rept, count, FrameCmd.Span.Code, true)
		}

	case CmdFileInput, CmdFileOutput, CmdFileEdit, CmdFileRead, CmdFileWrite,
		CmdFileRewind, CmdFileKill, CmdFileSave,
		CmdFileGlobalInput, CmdFileGlobalOutput, CmdFileGlobalRewind, CmdFileGlobalKill:
		cmdSuccess = FileCommand(command, rept, count, tparam, fromSpan)

	case CmdFileExecute:
		if CurrentFrame == FrameCmd {
			ScreenMessage(MsgNotWhileEditingCmd)
			goto l99
		}
		if TparGet1(tparam, command, &request) {
			newTparam := request
			FrameCmd.ReturnFrame = CurrentFrame
			CurrentFrame = FrameCmd
			// Zap frame COMMAND's current contents
			firstLine = FrameCmd.FirstGroup.FirstLine
			lastLine = FrameCmd.LastGroup.LastLine.BLink
			if lastLine != nil {
				if !MarksSqueeze(firstLine, 1, lastLine.FLink, 1) {
					goto l99
				}
				if !LinesExtract(firstLine, lastLine) {
					goto l99
				}
				if !LinesDestroy(&firstLine, &lastLine) {
					goto l99
				}
			}
			if FileCommand(CmdFileExecute, LeadParamNone, 0, &newTparam, false) {
				CurrentFrame = CurrentFrame.ReturnFrame
				// Recompile and execute frame COMMAND
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
		if TparGet1(tparam, command, &request) {
			newName = request.Str.Slice(1, request.Len)
			cmdSuccess = FrameEdit(newName)
		}

	case CmdFrameKill:
		if TparGet1(tparam, command, &request) {
			newName = request.Str.Slice(1, request.Len)
			cmdSuccess = FrameKill(newName)
		}

	case CmdFrameParameters:
		cmdSuccess = FrameParameter(tparam)

	case CmdFrameReturn:
		for i = 1; i <= count; i++ {
			if CurrentFrame.ReturnFrame == nil {
				CurrentFrame = oldFrame
				goto l99
			}
			CurrentFrame = CurrentFrame.ReturnFrame
		}
		cmdSuccess = true

	case CmdGet:
		if TparGet1(tparam, command, &request) {
			if request.Len == 0 {
				// If didn't specify, use default
				request = CurrentFrame.GetTpar
				if request.Len == 0 {
					ScreenMessage(MsgNoDefaultStr)
					goto l99
				}
			} else {
				CurrentFrame.GetTpar = request // If did specify, save for next time
			}
			cmdSuccess = EqsGetRepGet(count, request, fromSpan)
		}

	case CmdHelp:
		if LudwigMode == LudwigBatch {
			ScreenMessage(MsgInteractiveModeOnly)
			goto l99
		}
		if TparGet1(tparam, command, &request) {
			HelpHelp(string(request.Str.Slice(1, request.Len)))
			cmdSuccess = true // Never fails
		}

	case CmdInsertChar:
		cmdSuccess = CharcmdInsert(command, rept, count, fromSpan)

	case CmdInsertLine:
		if count != 0 {
			cmdSuccess = LinesCreate(int(math.Abs(float64(count))), &firstLine, &lastLine)
			if cmdSuccess {
				cmdSuccess = LinesInject(firstLine, lastLine, CurrentFrame.Dot.Line)
			}
			if cmdSuccess {
				if count > 0 {
					cmdSuccess = MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
						&CurrentFrame.Marks[MarkEquals-MinMarkNumber])
					cmdSuccess = MarkCreate(firstLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot)
				} else {
					cmdSuccess = MarkCreate(firstLine, CurrentFrame.Dot.Col,
						&CurrentFrame.Marks[MarkEquals-MinMarkNumber])
				}
				CurrentFrame.TextModified = true
				MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
					&CurrentFrame.Marks[MarkModified-MinMarkNumber])
			}
		} else {
			cmdSuccess = MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
				&CurrentFrame.Marks[MarkEquals-MinMarkNumber])
		}

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
		} else if TparGet1(tparam, command, &request) {
			if request.Con == nil {
				cmdSuccess = TextInsert(true, count, request.Str, request.Len, CurrentFrame.Dot)
				if cmdSuccess && (count*request.Len != 0) {
					CurrentFrame.TextModified = true
					cmdSuccess = MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
						&CurrentFrame.Marks[MarkModified-MinMarkNumber])
				}
			} else {
				for i = 1; i <= count; i++ {
					if !TextInsertTpar(&request, CurrentFrame.Dot, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
						goto l99
					}
				}
				CurrentFrame.TextModified = true
				cmdSuccess = MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
					&CurrentFrame.Marks[MarkModified-MinMarkNumber])
			}
		}

	case CmdInsertInvisible:
		if LudwigMode != LudwigScreen {
			goto l99
		}
		if CurrentFrame.Dot.Col > CurrentFrame.Dot.Line.Used {
			i = MaxStrLenP - CurrentFrame.Dot.Col
		} else {
			i = MaxStrLen - CurrentFrame.Dot.Line.Used
		}
		if rept == LeadParamPIndef {
			count = i
		}
		if count > i {
			goto l99
		}
		newStr = BlankString.Clone()
		i = 0
		for i < count {
			key = VduGetKey()
			if TtControlC {
				goto l99
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
			if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
				&CurrentFrame.Marks[MarkModified-MinMarkNumber]) {
				cmdSuccess = false
			}
		}

	case CmdJump:
		switch rept {
		case LeadParamNone, LeadParamPlus, LeadParamPInt:
			if CurrentFrame.Dot.Col+count > MaxStrLenP {
				goto l99
			}
		case LeadParamMinus, LeadParamNInt:
			if CurrentFrame.Dot.Col <= -count {
				goto l99
			}
		case LeadParamPIndef:
			if CurrentFrame.Dot.Col > CurrentFrame.Dot.Line.Used+1 {
				goto l99
			}
			count = 1 + CurrentFrame.Dot.Line.Used - CurrentFrame.Dot.Col
		case LeadParamNIndef:
			count = 1 - CurrentFrame.Dot.Col
		case LeadParamMarker:
			if !MarkCreate(theMark.Line, theMark.Col, &CurrentFrame.Dot) {
				goto l99
			}
			count = 0
		}
		CurrentFrame.Dot.Col += count
		cmdSuccess = true

	case CmdLineCentre:
		cmdSuccess = WordCentre(rept, count)

	case CmdLineFill:
		cmdSuccess = WordFill(rept, count)

	case CmdLineJustify:
		cmdSuccess = WordJustify(rept, count)

	case CmdLineSquash:
		cmdSuccess = WordSqueeze(rept, count)

	case CmdLineLeft:
		cmdSuccess = WordLeft(rept, count)

	case CmdLineRight:
		cmdSuccess = WordRight(rept, count)

	case CmdWordAdvance:
		if FileData.OldCmds {
			cmdSuccess = WordAdvanceWord(rept, count)
		} else {
			cmdSuccess = NewwordAdvanceWord(rept, count)
		}

	case CmdWordDelete:
		if FileData.OldCmds {
			cmdSuccess = WordDeleteWord(rept, count)
		} else {
			cmdSuccess = NewwordDeleteWord(rept, count)
		}

	case CmdAdvanceParagraph:
		cmdSuccess = NewwordAdvanceParagraph(rept, count)

	case CmdDeleteParagraph:
		cmdSuccess = NewwordDeleteParagraph(rept, count)

	case CmdMark:
		if count < 0 {
			if CurrentFrame.Marks[-count] != nil {
				cmdSuccess = MarkDestroy(&CurrentFrame.Marks[-count])
			} else {
				cmdSuccess = true
			}
		} else {
			cmdSuccess = MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
				&CurrentFrame.Marks[count])
		}

	case CmdNoop:
		// Nothing to do, as one might expect

	case CmdCommand:
		if rept == LeadParamMinus {
			if EditMode != ModeCommand {
				PreviousMode = EditMode
				EditMode = ModeCommand
			} else {
				goto l99
			}
		} else {
			if EditMode == ModeCommand {
				EditMode = PreviousMode
			} else {
				if LudwigMode != LudwigScreen {
					ScreenMessage(MsgScreenModeOnly)
					goto l99
				}
				cmdSuccess = UserCommandIntroducer()
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
		} else if TparGet1(tparam, command, &request) {
			cmdSuccess = TextOvertype(true, count, request.Str, request.Len, CurrentFrame.Dot)
			if cmdSuccess && (count*request.Len != 0) {
				CurrentFrame.TextModified = true
				if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
					&CurrentFrame.Marks[MarkModified-MinMarkNumber]) {
					cmdSuccess = false
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
		if TparGet1(tparam, command, &request) {
			if !OpsysCommand(&request, &firstLine, &lastLine, &i) {
				goto l99
			}
			if firstLine != nil {
				if !LinesInject(firstLine, lastLine, CurrentFrame.Dot.Line) {
					goto l99
				}
				if !MarkCreate(firstLine, 1, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
					goto l99
				}
				CurrentFrame.TextModified = true
				if !MarkCreate(lastLine.FLink, 1, &CurrentFrame.Marks[MarkModified-MinMarkNumber]) {
					goto l99
				}
				if !MarkCreate(lastLine.FLink, 1, &CurrentFrame.Dot) {
					goto l99
				}
				cmdSuccess = true
			}
		}

	case CmdPositionColumn:
		if count > MaxStrLen {
			goto l99
		}
		CurrentFrame.Dot.Col = count
		cmdSuccess = true

	case CmdPositionLine:
		if !LineFromNumber(CurrentFrame, count, &newLine) {
			goto l99
		}
		if newLine == nil {
			goto l99
		}
		cmdSuccess = true
		if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
			goto l99
		}
		MarkCreate(newLine, 1, &CurrentFrame.Dot)

	case CmdQuit:
		cmdSuccess = QuitCommand()

	case CmdReplace:
		if TparGet2(tparam, command, &request, &request2) {
			if request.Len == 0 { // If didn't specify, use default
				if CurrentFrame.Rep1Tpar.Len == 0 {
					ScreenMessage(MsgNoDefaultStr)
					goto l99
				}
			} else {
				CurrentFrame.Rep1Tpar = request // If did specify, save for next time
				TparCleanObject(&CurrentFrame.Rep2Tpar)
				CurrentFrame.Rep2Tpar = request2
				request2.Con = nil
			}
			cmdSuccess = EqsGetRepRep(rept, count, CurrentFrame.Rep1Tpar,
				CurrentFrame.Rep2Tpar, fromSpan)
		}

	case CmdRubout:
		cmdSuccess = CharcmdRubout(command, rept, count, fromSpan)

	case CmdSetMarginLeft:
		if rept == LeadParamMinus {
			CurrentFrame.MarginLeft = InitialMarginLeft
		} else {
			if CurrentFrame.Dot.Col >= CurrentFrame.MarginRight {
				ScreenMessage(MsgLeftMarginGeRight)
				goto l99
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
				goto l99
			}
			CurrentFrame.MarginRight = CurrentFrame.Dot.Col
		}
		cmdSuccess = true

	case CmdSpanJump, CmdSpanCompile, CmdSpanCopy, CmdSpanDefine,
		CmdSpanExecute, CmdSpanExecuteNoRecompile, CmdSpanTransfer:
		if TparGet1(tparam, command, &request) {
			newName = request.Str.Slice(1, request.Len)
			switch command {
			case CmdSpanDefine:
				if rept == LeadParamMinus {
					if SpanFind(newName, &newSpan, &oldSpan) {
						cmdSuccess = SpanDestroy(&newSpan)
					} else {
						ScreenMessage(MsgNoSuchSpan)
					}
				} else {
					cmdSuccess = SpanCreate(newName, theMark, CurrentFrame.Dot)
				}

			case CmdSpanJump:
				if SpanFind(newName, &newSpan, &oldSpan) {
					if rept == LeadParamMinus {
						newCol = newSpan.MarkOne.Col
						newLine = newSpan.MarkOne.Line
					} else {
						newCol = newSpan.MarkTwo.Col
						newLine = newSpan.MarkTwo.Line
					}
					if newLine.Group.Frame == CurrentFrame {
						cmdSuccess = MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col,
							&CurrentFrame.Marks[MarkEquals-MinMarkNumber])
						if cmdSuccess {
							cmdSuccess = MarkCreate(newLine, newCol, &CurrentFrame.Dot)
						}
					} else {
						fr := newLine.Group.Frame
						if FrameEdit(fr.Span.Name) {
							if fr.Marks[MarkEquals-MinMarkNumber] != nil {
								MarkDestroy(&fr.Marks[MarkEquals-MinMarkNumber])
							}
							cmdSuccess = MarkCreate(newLine, newCol, &fr.Dot)
						}
					}
				} else {
					ScreenMessage(MsgNoSuchSpan)
				}

			case CmdSpanCopy, CmdSpanTransfer:
				if SpanFind(newName, &newSpan, &oldSpan) {
					cmdSuccess = TextMove(
						command == CmdSpanCopy,
						count,
						newSpan.MarkOne,
						newSpan.MarkTwo,
						CurrentFrame.Dot, // Dest
						&CurrentFrame.Marks[MarkEquals-MinMarkNumber], // New_Start
						&CurrentFrame.Dot, // New_End
					)
					if command == CmdSpanTransfer && newSpan.Frame == nil && cmdSuccess {
						MarkCreate(CurrentFrame.Marks[MarkEquals-MinMarkNumber].Line,
							CurrentFrame.Marks[MarkEquals-MinMarkNumber].Col, &newSpan.MarkOne)
						MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &newSpan.MarkTwo)
					}
				} else {
					ScreenMessage(MsgNoSuchSpan)
				}

			case CmdSpanCompile, CmdSpanExecute, CmdSpanExecuteNoRecompile:
				if SpanFind(newName, &newSpan, &oldSpan) {
					if newSpan.Code == nil || command != CmdSpanExecuteNoRecompile {
						if !CodeCompile(newSpan, true) {
							goto l99
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
		if !TparGet2(tparam, command, &request, &request2) {
			goto l99
		}
		if request.Len == 0 {
			goto l99
		}
		newName = request.Str.Slice(1, request.Len)
		if SpanFind(newName, &newSpan, &oldSpan) {
			// Grunge the old one
			if newSpan == FrameOops.Span {
				if !TextRemove(FrameOops.Span.MarkOne, FrameOops.Span.MarkTwo) {
					goto l99
				}
			} else {
				// Make sure oops_span is okay
				if !MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Span.MarkTwo) {
					goto l99
				}
				if !TextMove(
					false, // Don't copy, transfer
					1,     // One instance of
					newSpan.MarkOne,
					newSpan.MarkTwo,
					FrameOops.Span.MarkTwo, // destination
					&FrameOops.Marks[MarkEquals-MinMarkNumber], // leave at start
					&FrameOops.Dot, // leave at end
				) {
					goto l99
				}
			}
		} else {
			// Create a span in frame "HEAP"
			if !MarkCreate(FrameHeap.LastGroup.LastLine, 1, &FrameHeap.Span.MarkTwo) {
				goto l99
			}
			if !SpanCreate(newName, FrameHeap.Span.MarkTwo, FrameHeap.Span.MarkTwo) {
				goto l99
			}
			if !SpanFind(newName, &newSpan, &oldSpan) {
				goto l99
			}
		}
		// Now copy the tpar into the span
		if !TextInsertTpar(&request2, newSpan.MarkTwo, &newSpan.MarkOne) {
			goto l99
		}
		fr := newSpan.MarkTwo.Line.Group.Frame
		fr.TextModified = true
		cmdSuccess = MarkCreate(newSpan.MarkTwo.Line, newSpan.MarkTwo.Col, &fr.Marks[MarkModified-MinMarkNumber])

	case CmdSplitLine:
		if CurrentFrame.Dot.Line.FLink == nil {
			if !TextRealizeNull(CurrentFrame.Dot.Line) {
				goto l99
			}
		}
		cmdSuccess = TextSplitLine(CurrentFrame.Dot, 0, &CurrentFrame.Marks[MarkEquals-MinMarkNumber])

	case CmdSwapLine:
		cmdSuccess = SwapLine(rept, count)

	case CmdUserCommandIntroducer:
		if LudwigMode != LudwigScreen {
			ScreenMessage(MsgScreenModeOnly)
			goto l99
		}
		cmdSuccess = UserCommandIntroducer()

	case CmdUserKey:
		if LudwigMode != LudwigScreen {
			ScreenMessage(MsgScreenModeOnly)
			goto l99
		}
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
			goto l99
		}
		cmdSuccess = UserParent()

	case CmdUserSubprocess:
		if LudwigMode == LudwigBatch {
			ScreenMessage(MsgInteractiveModeOnly)
			goto l99
		}
		cmdSuccess = UserSubprocess()

	case CmdUserUndo:
		cmdSuccess = UserUndo()

	case CmdWindowBackward, CmdWindowEnd, CmdWindowForward, CmdWindowLeft,
		CmdWindowMiddle, CmdWindowNew, CmdWindowRight, CmdWindowScroll,
		CmdWindowSetHeight, CmdWindowTop, CmdWindowUpdate:
		cmdSuccess = WindowCommand(command, rept, count, fromSpan)

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
		switch CmdAttrib[command].EqAction {
		case EqOld:
			eqSet = MarkCreate(oldDot.Line, oldDot.Col, &oldFrame.Marks[MarkEquals-MinMarkNumber])
		case EqDel:
			eqSet = (oldFrame.Marks[MarkEquals-MinMarkNumber] == nil) ||
				MarkDestroy(&oldFrame.Marks[MarkEquals-MinMarkNumber])
		case EqNil:
			eqSet = true
		}
		if !eqSet {
			ScreenMessage(MsgEqualsNotSet)
		}
	}

l99:
	TparCleanObject(&request)
	TparCleanObject(&request2)
	ExecLevel--
	return cmdSuccess
}
