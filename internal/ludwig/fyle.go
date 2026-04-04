/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         FYLE
//
// Description:  Open/Create, Read/Write, Close/Delete Input/Output files.

package ludwig

import "errors"

var (
	errNoFileOpen         = errors.New(MsgNoFileOpen)
	errFileAlreadyOpen    = errors.New(MsgFileAlreadyOpen)
	errNotOutputFile      = errors.New(MsgNotOutputFile)
	errNotInputFile       = errors.New(MsgNotInputFile)
	errNoMoreFilesAllowed = errors.New(MsgNoMoreFilesAllowed)
	errNoOutput           = errors.New(MsgNoOutput)
	errEmpty              = errors.New("")
)

const blankName = "                               "

// FileName returns a file's name, in the specified width.
func FileName(fp *FileObject, maxLen int) string {
	if maxLen < 5 {
		maxLen = 5
	}
	var headLen, tailLen int
	if len(fp.Filename) <= maxLen {
		headLen = len(fp.Filename)
		tailLen = 0
	} else {
		// Cut chars out the middle of the file name, insert '---'
		tailLen = (maxLen - 3) / 2
		headLen = maxLen - 3 - tailLen
	}
	if tailLen == 0 {
		return fp.Filename[:headLen]
	} else {
		return fp.Filename[:headLen] + "---" + fp.Filename[len(fp.Filename)-tailLen:]
	}
}

// FileTable lists the current files.
func FileTable() {
	ScreenUnload(&Screen)
	ScreenHome(true)
	ScreenWriteStrWidth(0, "Usage   Mod Frame  Filename", 27)
	ScreenWriteln()
	ScreenWriteStrWidth(0, "------- --- ------ --------", 27)
	ScreenWriteln()
	ScreenWriteln()
	for fileSlot := 1; fileSlot <= MaxFiles; fileSlot++ {
		if Files[fileSlot] != nil {
			var frameName string
			if FilesFrames[fileSlot] != nil {
				frameName = FilesFrames[fileSlot].Span.Name
			} else {
				frameName = blankName[:6]
			}

			if FilesFrames[fileSlot] != nil {
				if Files[fileSlot].OutputFlag {
					ScreenWriteStrWidth(0, "FO ", 3)
				} else {
					ScreenWriteStrWidth(0, "FI ", 3)
				}
			} else if fileSlot == FgiFile {
				ScreenWriteStrWidth(0, "FGI", 3)
			} else if fileSlot == FgoFile {
				ScreenWriteStrWidth(0, "FGO", 3)
			} else if Files[fileSlot].OutputFlag {
				ScreenWriteStrWidth(0, "FFO", 3)
			} else {
				ScreenWriteStrWidth(0, "FFI", 3)
			}

			if Files[fileSlot].Eof {
				ScreenWriteStrWidth(1, "EOF", 3)
			} else {
				ScreenWriteStrWidth(1, "   ", 3)
			}

			if FilesFrames[fileSlot] != nil {
				if FilesFrames[fileSlot].TextModified {
					ScreenWriteStrWidth(1, " * ", 3)
				} else {
					ScreenWriteStrWidth(1, "   ", 3)
				}
			} else {
				ScreenWriteStrWidth(1, "   ", 3)
			}

			ScreenWriteNameStr(1, frameName, max(6, len(frameName)))
			if len(frameName) > 6 {
				ScreenWriteln()
				ScreenWriteStrWidth(0, "                  ", 18)
			}

			var room int
			if LudwigMode == LudwigScreen {
				room = TerminalInfo.Width - 18 - 1
			} else {
				room = FileNameLen
			}
			ScreenWriteFileNameStr(1, FileName(Files[fileSlot], room))
			ScreenWriteln()
		}
	}
	ScreenPause(&Screen)
}

// FileFixEOP updates the end-of-page marker
func FileFixEOP(eof bool, eopLine *LineHdrObject) {
	if eof {
		eopLine.Str.FillCopyBytes([]byte("<End of File>  "), 1, MaxStrLen, ' ')
	} else {
		eopLine.Str.FillCopyBytes([]byte("<Page Boundary>"), 1, MaxStrLen, ' ')
	}
	if eopLine.ScrRowNr != 0 {
		VduDim()
		ScreenDrawLine(&Screen, eopLine)
		VduNormal()
	}
}

// FileCreateOpen parses argv and creates I/O streams to files.
func FileCreateOpen(argv []string, parse ParseType, inputfp **FileObject, outputfp **FileObject) bool {
	switch parse {
	case ParseCommand, ParseInput, ParseEdit, ParseStdin, ParseExecute:
		if *inputfp != nil {
			ScreenMessage(&Screen, MsgFileAlreadyInUse)
			return false
		}
		*inputfp = &FileObject{}
		(*inputfp).Valid = false
		(*inputfp).FirstLine = nil
		(*inputfp).LastLine = nil
		(*inputfp).LineCount = 0
		(*inputfp).OutputFlag = false
		(*inputfp).Eof = false
	}

	switch parse {
	case ParseCommand, ParseOutput, ParseEdit:
		if *outputfp != nil {
			ScreenMessage(&Screen, MsgFileAlreadyInUse)
			return false
		}
		*outputfp = &FileObject{}
		(*outputfp).Valid = false
		(*outputfp).FirstLine = nil
		(*outputfp).LastLine = nil
		(*outputfp).LineCount = 0
		(*outputfp).OutputFlag = true
	}

	result := FilesysParse(argv, parse, &FileData, *inputfp, *outputfp)
	if *inputfp != nil && !(*inputfp).Valid {
		*inputfp = nil
	}
	if *outputfp != nil && !(*outputfp).Valid {
		*outputfp = nil
	}
	return result
}

// FileCloseDelete closes a file, if it is an output file it can optionally be deleted.
func FileCloseDelete(fp **FileObject, delet bool, msgs bool) bool {
	if fp != nil && *fp != nil {
		deletFlag := 0
		if delet {
			deletFlag = 1
		}
		if FilesysClose(*fp, deletFlag, msgs) {
			*fp = nil
			return true
		}
	}
	return false
}

// FileRead reads a series of lines from input file.
func FileRead(fp *FileObject, count int, bestTry bool) (*LineHdrObject, *LineHdrObject, int, bool) {
	if fp.OutputFlag {
		ScreenMessage(&Screen, MsgNotInputFile)
		return nil, nil, 0, false
	}

	var line *LineHdrObject
	for count > fp.LineCount && !fp.Eof {
		buffer := NewBlankStrObject(MaxStrLen)
		var outlen int
		if FilesysRead(fp, buffer, &outlen) {
			if outlen > 0 {
				outlen = buffer.TrimmedLen(' ', outlen)
			}
			line, _ = LinesCreate(1)
			LineChangeLength(line, outlen)
			line.Str.FillCopy(buffer, 1, outlen, 1, line.Len(), ' ')
			line.Used = outlen
			line.BLink = fp.LastLine
			if fp.LastLine != nil {
				fp.LastLine.FLink = line
			} else {
				fp.FirstLine = line
			}
			fp.LastLine = line
			fp.LineCount++
		} else if !fp.Eof {
			// Something drastically wrong with the input!
			// As a TEMPORARY measure, ignore.
		}
	}

	// Check there are enough lines.
	if fp.LineCount < count {
		if !bestTry {
			ScreenMessage(&Screen, MsgNotEnoughInputLeft)
			return nil, nil, 0, false
		}
		count = fp.LineCount
	}

	// Break off the required lines.
	var first, last *LineHdrObject
	if count == 0 {
		first = nil
		last = nil
	} else if fp.LineCount == count {
		// Give caller the whole list.
		first = fp.FirstLine
		last = fp.LastLine
		fp.FirstLine = nil
		fp.LastLine = nil
		fp.LineCount = 0
	} else {
		// Give caller the first 'count' lines in the list.
		// Find last line to be removed.
		if count < fp.LineCount/2 {
			line = fp.FirstLine
			for i := 2; i <= count; i++ {
				line = line.FLink
			}
		} else {
			line = fp.LastLine
			for i := fp.LineCount; i > count; i-- {
				line = line.BLink
			}
		}

		// Remove lines from list.
		first = fp.FirstLine
		last = line
		fp.FirstLine = line.FLink
		line.FLink = nil
		fp.FirstLine.BLink = nil
		fp.LineCount -= count
	}

	return first, last, count, true
}

// FileWrite writes a series of lines to an output file.
func FileWrite(firstLine *LineHdrObject, lastLine *LineHdrObject, fp *FileObject) bool {
	for firstLine != nil {
		if !FilesysWrite(fp, firstLine.Str, firstLine.Used) {
			return false
		}
		if firstLine == lastLine {
			return true
		}
		firstLine = firstLine.FLink
	}
	return true
}

// FileWindthru writes all the remaining input file to the output file.
func FileWindthru(current *FrameObject, fromSpan bool) bool {
	if current.OutputFile == 0 {
		return false
	}
	if Files[current.OutputFile] == nil {
		return false
	}
	if current.TextModified && !fromSpan {
		ScreenMessage(&Screen, MsgWritingFile)
		if LudwigMode == LudwigScreen {
			VduFlush()
		}
	}
	defer func() {
		if current.TextModified && !fromSpan {
			ScreenClearMsgs(&Screen, false)
		}
	}()

	firstLine := current.FirstGroup.FirstLine
	lastLine := current.LastGroup.LastLine.BLink

	if firstLine != nil && lastLine != nil {
		if current.TextModified {
			if !FileWrite(firstLine, lastLine, Files[current.OutputFile]) {
				return false
			}
		}
		MarksSqueeze(firstLine, 1, lastLine.FLink, 1)
		LinesExtract(firstLine, lastLine)
		if current.InputFile != 0 {
			if Files[current.InputFile] != nil {
				Files[current.InputFile].LineCount = 0
			}
		}
	}

	result := true
	if current.TextModified {
		if current.InputFile != 0 {
			if Files[current.InputFile] != nil {
				if !Files[current.InputFile].Eof {
					buffer := NewBlankStrObject(MaxStrLen)
					var outlen int
					for FilesysRead(Files[current.InputFile], buffer, &outlen) {
						buflen := 0
						if outlen > 0 {
							buflen = buffer.TrimmedLen(' ', outlen)
						}
						if !FilesysWrite(Files[current.OutputFile], buffer, buflen) {
							return false
						}
					}
				}
				result = Files[current.InputFile].Eof
			}
		}
	}
	return result
}

// FileRewind rewinds a file.
func FileRewind(fp **FileObject) bool {
	if *fp != nil {
		if (*fp).FirstLine != nil {
			(*fp).FirstLine = nil
			(*fp).LastLine = nil
			(*fp).LineCount = 0
		}
	}
	FilesysRewind(*fp)
	return true
}

// FilePage handles paging operations.
func FilePage(currentFrame *FrameObject, exitAbort *bool) bool {
	var firstLine, lastLine *LineHdrObject
	if !ExecComputeLineRange(currentFrame, LeadParamNIndef, 0, &firstLine, &lastLine) {
		ScreenMessage(&Screen, DbgInternalLogicError)
		return false
	}

	// Page out the stuff above the dot line.
	if firstLine != nil {
		if currentFrame.OutputFile != 0 &&
			!FileWrite(firstLine, lastLine, Files[currentFrame.OutputFile]) {
			*exitAbort = true
			return false
		}
		if lastLine.FLink == nil {
			return false
		}
		MarksSqueeze(firstLine, 1, lastLine.FLink, 1)
		LinesExtract(firstLine, lastLine)
	}

	defer func() {
		SetupSyntaxHighlighting(currentFrame)
		if currentFrame.InputFile != 0 {
			FileFixEOP(Files[currentFrame.InputFile].Eof, currentFrame.LastGroup.LastLine)
		}
	}()

	// Page in the new lines
	if currentFrame.InputFile == 0 {
		return true
	}
	for (currentFrame.SpaceLeft*10 > currentFrame.SpaceLimit) && !TtControlC {
		var i int
		var ok bool
		if firstLine, lastLine, i, ok = FileRead(Files[currentFrame.InputFile], 50, true); !ok {
			return false
		}
		currentFrame.InputCount += i

		if firstLine == nil {
			return true
		}
		LinesInject(firstLine, lastLine, currentFrame.LastGroup.LastLine)

		// If dot was on the null line, shift it onto the first line
		if currentFrame.Dot.Line.FLink == nil {
			MarkCreate(firstLine, currentFrame.Dot.Col, &currentFrame.Dot)
		}
	}
	return true
}

func checkSlotAllocation(slot int, mustBeAllocated bool) error {
	if (slot == 0) == mustBeAllocated {
		if mustBeAllocated {
			return errNoFileOpen
		} else {
			return errFileAlreadyOpen
		}
	}
	return nil
}

func checkSlotUsage(slot int, mustBeInUse bool) error {
	if err := checkSlotAllocation(slot, true); err != nil {
		return err
	}
	if (Files[slot] == nil) == mustBeInUse {
		if mustBeInUse {
			return errNoFileOpen
		} else {
			return errFileAlreadyOpen
		}
	}
	return nil
}

func checkSlotDirection(slot int, mustBeOutput bool) error {
	if err := checkSlotUsage(slot, true); err != nil {
		return err
	}
	if Files[slot].OutputFlag != mustBeOutput {
		if mustBeOutput {
			return errNotOutputFile
		} else {
			return errNotInputFile
		}
	}
	return nil
}

func freeFile(slot int) error {
	if err := checkSlotAllocation(slot, true); err != nil {
		return err
	}
	if FilesFrames[slot] != nil {
		if slot == FilesFrames[slot].OutputFile {
			FilesFrames[slot].OutputFile = 0
		} else {
			FileFixEOP(true, FilesFrames[slot].LastGroup.LastLine)
			FilesFrames[slot].InputFile = 0
		}
		FilesFrames[slot] = nil
	} else if slot == FgiFile {
		FgiFile = 0
	} else if slot == FgoFile {
		FgoFile = 0
	}
	return nil
}

func getFreeSlot(fileSlot int) (int, error) {
	slot := 1
	for slot <= MaxFiles && (Files[slot] != nil || slot == fileSlot) {
		slot++
	}
	if slot > MaxFiles {
		return -1, errNoMoreFilesAllowed
	}
	return slot, nil
}

func getFileName(frame *FrameObject, tparam *TParObject, command Commands) (string, error) {
	tpFileName := TParObject{}
	if !TparGet1(frame, tparam, command, &tpFileName) {
		// TparAnalyze has already output the message, so we return an empty one
		return "", errEmpty
	}
	return tpFileName.Str.Slice(1, tpFileName.Len), nil
}

// FileCommand executes file commands.
func FileCommand(frame *FrameObject, command Commands, rept LeadParam, count int, tparam *TParObject, fromSpan bool) bool {
	savedCmd := command
	if rept == LeadParamMinus && command != CmdFileWrite {
		savedCmd = command
		command = CmdFileClose
	}

	var err error
	fileSlot := 0

	defer func() {
		if err != nil && err != errEmpty {
			ScreenMessage(&Screen, err.Error())
		}
	}()

	switch command {
	case CmdFileInput:
		if err = checkSlotAllocation(frame.InputFile, false); err != nil {
			return false
		}
		if fileSlot, err = getFreeSlot(fileSlot); err != nil {
			return false
		}
		var fnm string
		if fnm, err = getFileName(frame, tparam, command); err != nil {
			return false
		}
		var dummyFptr *FileObject
		if !FileCreateOpen([]string{fnm}, ParseInput, &Files[fileSlot], &dummyFptr) {
			return false
		}
		frame.InputFile = fileSlot
		FilesFrames[fileSlot] = frame
		if !fromSpan {
			ScreenMessage(&Screen, MsgLoadingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		FilePage(frame, &ExitAbort)
		if !fromSpan {
			ScreenClearMsgs(&Screen, false)
		}

	case CmdFileGlobalInput:
		if err = checkSlotAllocation(FgiFile, false); err != nil {
			return false
		}
		if fileSlot, err = getFreeSlot(fileSlot); err != nil {
			return false
		}
		if Files[fileSlot] == nil {
			var fnm string
			if fnm, err = getFileName(frame, tparam, command); err != nil {
				return false
			}
			var dummyFptr *FileObject
			if !FileCreateOpen([]string{fnm}, ParseInput, &Files[fileSlot], &dummyFptr) {
				return false
			}
		}
		FgiFile = fileSlot

	case CmdFileEdit:
		if err = checkSlotAllocation(frame.InputFile, false); err != nil {
			return false
		}
		if err = checkSlotAllocation(frame.OutputFile, false); err != nil {
			return false
		}
		if fileSlot, err = getFreeSlot(fileSlot); err != nil {
			return false
		}
		var fileSlot2 int
		if fileSlot2, err = getFreeSlot(fileSlot); err != nil {
			return false
		}
		var fnm string
		if fnm, err = getFileName(frame, tparam, command); err != nil {
			return false
		}
		if !FileCreateOpen([]string{fnm}, ParseEdit, &Files[fileSlot], &Files[fileSlot2]) {
			return false
		}
		frame.InputFile = fileSlot
		FilesFrames[fileSlot] = frame
		frame.OutputFile = fileSlot2
		FilesFrames[fileSlot2] = frame
		if !fromSpan {
			ScreenMessage(&Screen, MsgLoadingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		FilePage(frame, &ExitAbort)
		if !fromSpan {
			ScreenClearMsgs(&Screen, false)
		}

	case CmdFileExecute:
		if err = checkSlotAllocation(frame.InputFile, false); err != nil {
			return false
		}
		if fileSlot, err = getFreeSlot(fileSlot); err != nil {
			return false
		}
		var fnm string
		if fnm, err = getFileName(frame, tparam, command); err != nil {
			return false
		}
		var dummyFptr *FileObject
		if !FileCreateOpen([]string{fnm}, ParseExecute, &Files[fileSlot], &dummyFptr) {
			return false
		}
		frame.InputFile = fileSlot
		FilesFrames[fileSlot] = frame
		FilePage(frame, &ExitAbort)
		if err = freeFile(fileSlot); err != nil {
			return false
		}
		if !FileCloseDelete(&Files[fileSlot], true, false) {
			return false
		}

	case CmdFileClose:
		switch savedCmd {
		case CmdFileInput:
			fileSlot = frame.InputFile
		case CmdFileOutput:
			fileSlot = frame.OutputFile
		case CmdFileGlobalInput:
			fileSlot = FgiFile
		case CmdFileGlobalOutput:
			fileSlot = FgoFile
		case CmdFileEdit:
			fileSlot = frame.InputFile
		}
		if savedCmd == CmdFileOutput || savedCmd == CmdFileEdit {
			if !FileWindthru(frame, fromSpan) {
				return false
			}
			ScreenFixup(&Screen, frame)
		}
		if err = freeFile(fileSlot); err != nil {
			return false
		}
		if savedCmd == CmdFileGlobalInput || savedCmd == CmdFileGlobalOutput {
			if !FileCloseDelete(&Files[fileSlot], false, true) {
				return false
			}
		} else {
			if !FileCloseDelete(&Files[fileSlot], !frame.TextModified,
				frame.TextModified || !Files[fileSlot].OutputFlag) {
				return false
			}
		}
		if savedCmd == CmdFileEdit {
			fileSlot = frame.OutputFile
			if err = freeFile(fileSlot); err != nil {
				return false
			}
			if !FileCloseDelete(&Files[fileSlot], !frame.TextModified, frame.TextModified) {
				return false
			}
		}
		if savedCmd == CmdFileOutput || savedCmd == CmdFileEdit {
			frame.TextModified = false
		}

	case CmdFileKill:
		fileSlot = frame.OutputFile
		if err = freeFile(fileSlot); err != nil {
			return false
		}
		if !FileCloseDelete(&Files[fileSlot], true, true) {
			return false
		}

	case CmdFileGlobalKill:
		fileSlot = FgoFile
		if err = freeFile(fileSlot); err != nil {
			return false
		}
		if !FileCloseDelete(&Files[fileSlot], true, true) {
			return false
		}

	case CmdFileOutput:
		if err = checkSlotAllocation(frame.OutputFile, false); err != nil {
			return false
		}
		if fileSlot, err = getFreeSlot(fileSlot); err != nil {
			return false
		}
		var fnm string
		if fnm, err = getFileName(frame, tparam, command); err != nil {
			return false
		}
		if frame.InputFile != 0 {
			if !FileCreateOpen([]string{fnm}, ParseOutput, &Files[frame.InputFile], &Files[fileSlot]) {
				return false
			}
		} else {
			var dummyFptr *FileObject
			if !FileCreateOpen([]string{fnm}, ParseOutput, &dummyFptr, &Files[fileSlot]) {
				return false
			}
		}
		frame.OutputFile = fileSlot
		FilesFrames[fileSlot] = frame

	case CmdFileGlobalOutput:
		if err = checkSlotAllocation(FgoFile, false); err != nil {
			return false
		}
		if fileSlot, err = getFreeSlot(fileSlot); err != nil {
			return false
		}
		if Files[fileSlot] == nil {
			var fnm string
			if fnm, err = getFileName(frame, tparam, command); err != nil {
				return false
			}
			var dummyFptr *FileObject
			if !FileCreateOpen([]string{fnm}, ParseOutput, &dummyFptr, &Files[fileSlot]) {
				return false
			}
		}
		FgoFile = fileSlot

	case CmdFileRead:
		if err = checkSlotAllocation(FgiFile, true); err != nil {
			return false
		}
		linesToRead := count
		if rept == LeadParamPIndef {
			linesToRead = MaxInt
		}
		first, last, _, ok := FileRead(Files[FgiFile], linesToRead, rept == LeadParamPIndef)
		if !ok {
			return false
		}
		if first != nil {
			LinesInject(first, last, frame.Dot.Line)
			MarkCreate(first, 1, &frame.Marks[MarkEquals])
			frame.TextModified = true
			MarkCreate(last.FLink, 1, &frame.Marks[MarkModified])
			MarkCreate(last.FLink, 1, &frame.Dot)
		}

	case CmdFileWrite:
		if err = checkSlotAllocation(FgoFile, true); err != nil {
			return false
		}
		if !fromSpan {
			ScreenMessage(&Screen, MsgWritingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		var first, last *LineHdrObject
		if !ExecComputeLineRange(frame, rept, int(count), &first, &last) {
			return false
		}
		if first != nil {
			if !FileWrite(first, last, Files[FgoFile]) {
				return false
			}
		}
		if !fromSpan {
			ScreenClearMsgs(&Screen, false)
		}

	case CmdFileRewind:
		if err = checkSlotDirection(frame.InputFile, false); err != nil {
			return false
		}
		if !FileRewind(&Files[frame.InputFile]) {
			return false
		}
		if !fromSpan {
			ScreenMessage(&Screen, MsgLoadingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		FilePage(frame, &ExitAbort)
		if !fromSpan {
			ScreenClearMsgs(&Screen, false)
		}

	case CmdFileGlobalRewind:
		if err = checkSlotDirection(FgiFile, false); err != nil {
			return false
		}
		if !FileRewind(&Files[FgiFile]) {
			return false
		}

	case CmdFileSave:
		if frame.OutputFile == 0 {
			err = errNoOutput
			return false
		}
		if !frame.TextModified {
			if !fromSpan {
				ScreenMessage(&Screen, MsgNotModified)
				if LudwigMode == LudwigScreen {
					VduFlush()
				}
			}
			return true
		}
		if !fromSpan {
			ScreenMessage(&Screen, MsgSavingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		linesWritten := Files[frame.OutputFile].LCounter
		first := frame.FirstGroup.FirstLine
		last := frame.LastGroup.LastLine.BLink
		if last != nil {
			if !FileWrite(first, last, Files[frame.OutputFile]) {
				return false
			}
		}
		var dummyFptr *FileObject
		if frame.InputFile != 0 {
			dummyFptr = Files[frame.InputFile]
		}
		if !FilesysSave(dummyFptr, Files[frame.OutputFile], linesWritten) {
			return false
		}
		var nrLines int
		if last == nil {
			nrLines = 0
		} else {
			nrLines = LineToNumber(last)
		}
		frame.InputCount = Files[frame.OutputFile].LCounter + nrLines
		if frame.InputFile != 0 {
			Files[frame.InputFile].LCounter = frame.InputCount
		}
		frame.TextModified = false
	}

	return true
}
