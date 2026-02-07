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

const blankName = "                               "

// FileName returns a file's name, in the specified width.
func FileName(fp *FileObject, maxLen int, actFnm *string) {
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
		*actFnm = fp.Filename[:headLen]
	} else {
		*actFnm = fp.Filename[:headLen] + "---" + fp.Filename[len(fp.Filename)-tailLen:]
	}
}

// FileTable lists the current files.
func FileTable() {
	ScreenUnload()
	ScreenHome(false)
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

			ScreenWriteNameStr(1, frameName, len(frameName))
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
			var compressedFnm string
			FileName(Files[fileSlot], room, &compressedFnm)
			ScreenWriteFileNameStr(1, compressedFnm, len(compressedFnm))
			ScreenWriteln()
		}
	}
	ScreenPause()
}

// FileFixEOP updates the end-of-page marker
func FileFixEOP(eof bool, eopLine *LineHdrObject) {
	if eof {
		eopLine.Str.FillCopyBytes([]byte("<End of File>  "), 1, MaxStrLen, ' ')
	} else {
		eopLine.Str.FillCopyBytes([]byte("<Page Boundary>"), 1, MaxStrLen, ' ')
	}
	if eopLine.ScrRowNr != 0 {
		ScreenDrawLine(eopLine)
	}
}

// FileCreateOpen parses fn and creates I/O streams to files.
func FileCreateOpen(fn *string, parse ParseType, inputfp **FileObject, outputfp **FileObject) bool {
	switch parse {
	case ParseCommand, ParseInput, ParseEdit, ParseStdin, ParseExecute:
		if *inputfp != nil {
			ScreenMessage(MsgFileAlreadyInUse)
			return false
		}
		*inputfp = &FileObject{}
		(*inputfp).Valid = false
		(*inputfp).FirstLine = nil
		(*inputfp).LastLine = nil
		(*inputfp).LineCount = 0
		(*inputfp).OutputFlag = false
		(*inputfp).Eof = false
		(*inputfp).Idx = MaxStrLen
		(*inputfp).Len = 0
	}

	switch parse {
	case ParseCommand, ParseOutput, ParseEdit:
		if *outputfp != nil {
			ScreenMessage(MsgFileAlreadyInUse)
			return false
		}
		*outputfp = &FileObject{}
		(*outputfp).Valid = false
		(*outputfp).FirstLine = nil
		(*outputfp).LastLine = nil
		(*outputfp).LineCount = 0
		(*outputfp).OutputFlag = true
	}

	result := FilesysParse(*fn, parse, &FileData, *inputfp, *outputfp)
	if *inputfp != nil && !(*inputfp).Valid {
		*inputfp = nil
	}
	if *outputfp != nil && !(*outputfp).Valid {
		*outputfp = nil
	}
	return result
}

// FileCloseDelete closes a file, if it is an output file it can optionally be deleted.
func FileCloseDelete(fp *FileObject, delet bool, msgs bool) bool {
	if fp != nil {
		deletFlag := 0
		if delet {
			deletFlag = 1
		}
		if FilesysClose(fp, deletFlag, msgs) {
			if fp.FirstLine != nil {
				var tmp1, tmp2 *LineHdrObject = fp.FirstLine, fp.LastLine
				LinesDestroy(&tmp1, &tmp2)
			}
		}
		return true
	}
	return false
}

// FileRead reads a series of lines from input file.
func FileRead(fp *FileObject, count int, bestTry bool, first **LineHdrObject, last **LineHdrObject, actualCnt *int) bool {
	if fp.OutputFlag {
		ScreenMessage(MsgNotInputFile)
		return false
	}

	var line, line2 *LineHdrObject
	for count > fp.LineCount && !fp.Eof {
		var buffer StrObject
		var outlen int
		if FilesysRead(fp, &buffer, &outlen) {
			if outlen > 0 {
				outlen = buffer.Length(' ', outlen)
			}
			if !LinesCreate(1, &line, &line2) {
				return false
			}
			if !LineChangeLength(line, outlen) {
				var tmp1, tmp2 *LineHdrObject = line, line2
				LinesDestroy(&tmp1, &tmp2)
				return false
			}
			ChFillCopy(&buffer, 1, outlen, line.Str, 1, line.Len, ' ')
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
			ScreenMessage(MsgNotEnoughInputLeft)
			return false
		}
		count = fp.LineCount
	}

	// Break off the required lines.
	*actualCnt = count
	if count == 0 {
		*first = nil
		*last = nil
	} else if fp.LineCount == count {
		// Give caller the whole list.
		*first = fp.FirstLine
		*last = fp.LastLine
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
		*first = fp.FirstLine
		*last = line
		fp.FirstLine = line.FLink
		line.FLink = nil
		fp.FirstLine.BLink = nil
		fp.LineCount -= count
	}

	return true
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
		ScreenMessage(MsgWritingFile)
		if LudwigMode == LudwigScreen {
			VduFlush()
		}
	}

	firstLine := current.FirstGroup.FirstLine
	lastLine := current.LastGroup.LastLine.BLink
	result := false

	if firstLine != nil && lastLine != nil {
		if current.TextModified {
			if !FileWrite(firstLine, lastLine, Files[current.OutputFile]) {
				goto l98
			}
		}
		if !MarksSqueeze(firstLine, 1, lastLine.FLink, 1) {
			goto l98
		}
		if !LinesExtract(firstLine, lastLine) {
			goto l98
		}
		var tmp1, tmp2 *LineHdrObject = firstLine, lastLine
		if !LinesDestroy(&tmp1, &tmp2) {
			goto l98
		}
		if current.InputFile != 0 {
			if Files[current.InputFile] != nil {
				Files[current.InputFile].LineCount = 0
			}
		}
	}

	result = true
	if current.TextModified {
		if current.InputFile != 0 {
			if Files[current.InputFile] != nil {
				if !Files[current.InputFile].Eof {
					var buffer StrObject
					var outlen int
					for FilesysRead(Files[current.InputFile], &buffer, &outlen) {
						buflen := 0
						if outlen > 0 {
							buflen = buffer.Length(' ', outlen)
						}
						if !FilesysWrite(Files[current.OutputFile], &buffer, buflen) {
							result = false
							goto l98
						}
					}
				}
				result = Files[current.InputFile].Eof
			}
		}
	}
l98:
	if current.TextModified && !fromSpan {
		ScreenClearMsgs(false)
	}
	return result
}

// FileRewind rewinds a file.
func FileRewind(fp **FileObject) bool {
	if *fp != nil {
		if (*fp).FirstLine != nil {
			var tmp1, tmp2 *LineHdrObject = (*fp).FirstLine, (*fp).LastLine
			LinesDestroy(&tmp1, &tmp2)
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
		ScreenMessage(DbgInternalLogicError)
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
		if !MarksSqueeze(firstLine, 1, lastLine.FLink, 1) {
			return false
		}
		if !LinesExtract(firstLine, lastLine) {
			return false
		}
		var tmp1, tmp2 *LineHdrObject = firstLine, lastLine
		if !LinesDestroy(&tmp1, &tmp2) {
			return false
		}
	}

	// Page in the new lines
	if currentFrame.InputFile == 0 {
		goto l98
	}
	for (currentFrame.SpaceLeft*10 > currentFrame.SpaceLimit) && !TtControlC {
		var i int
		if !FileRead(Files[currentFrame.InputFile], 50, true, &firstLine, &lastLine, &i) {
			return false
		}
		currentFrame.InputCount += uint32(i)

		if firstLine == nil {
			goto l98
		}
		if !LinesInject(firstLine, lastLine, currentFrame.LastGroup.LastLine) {
			return false
		}

		// If dot was on the null line, shift it onto the first line
		if currentFrame.Dot.Line.FLink == nil {
			if !MarkCreate(firstLine, currentFrame.Dot.Col, &currentFrame.Dot) {
				return false
			}
		}
	}
l98:
	if currentFrame.InputFile != 0 {
		FileFixEOP(Files[currentFrame.InputFile].Eof, currentFrame.LastGroup.LastLine)
	}
	return true
}

func checkSlotAllocation(slot int, mustBeAllocated bool, status *string) bool {
	if (slot == 0) == mustBeAllocated {
		if mustBeAllocated {
			*status = MsgNoFileOpen
		} else {
			*status = MsgFileAlreadyOpen
		}
		return false
	}
	return true
}

func checkSlotUsage(slot int, mustBeInUse bool, status *string) bool {
	if checkSlotAllocation(slot, true, status) {
		if (Files[slot] == nil) == mustBeInUse {
			if mustBeInUse {
				*status = MsgNoFileOpen
			} else {
				*status = MsgFileAlreadyOpen
			}
		} else {
			return true
		}
	}
	return false
}

func checkSlotDirection(slot int, mustBeOutput bool, status *string) bool {
	if checkSlotUsage(slot, true, status) {
		if Files[slot].OutputFlag != mustBeOutput {
			if mustBeOutput {
				*status = MsgNotOutputFile
			} else {
				*status = MsgNotInputFile
			}
		} else {
			return true
		}
	}
	return false
}

func freeFile(slot int, status *string) bool {
	if !checkSlotAllocation(slot, true, status) {
		return false
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
	return true
}

func getFreeSlot(newSlot *int, fileSlot int, status *string) bool {
	slot := 1
	for slot < MaxFiles && (Files[slot] != nil || slot == fileSlot) {
		slot++
	}
	if Files[slot] != nil {
		*status = MsgNoMoreFilesAllowed
		return false
	}
	*newSlot = slot
	return true
}

func getFileName(tparam *TParObject, fnm *string, command Commands) bool {
	tpFileName := TParObject{}
	tpFileName.Con = nil
	tpFileName.Nxt = nil
	if !TparGet1(tparam, command, &tpFileName) {
		return false
	}
	*fnm = tpFileName.Str.Slice(1, tpFileName.Len)
	TparCleanObject(&tpFileName)
	return true
}

// FileCommand executes file commands.
func FileCommand(command Commands, rept LeadParam, count int, tparam *TParObject, fromSpan bool) bool {
	savedCmd := command
	if rept == LeadParamMinus && command != CmdFileWrite {
		savedCmd = command
		command = CmdFileClose
	}

	var status string
	fileSlot := 0
	var fnm string
	result := false

	switch command {
	case CmdFileInput:
		if !checkSlotAllocation(CurrentFrame.InputFile, false, &status) {
			goto l99
		}
		if !getFreeSlot(&fileSlot, fileSlot, &status) {
			goto l99
		}
		if !getFileName(tparam, &fnm, command) {
			goto l99
		}
		var dummyFptr *FileObject
		if !FileCreateOpen(&fnm, ParseInput, &Files[fileSlot], &dummyFptr) {
			goto l99
		}
		CurrentFrame.InputFile = fileSlot
		FilesFrames[fileSlot] = CurrentFrame
		if !fromSpan {
			ScreenMessage(MsgLoadingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		FilePage(CurrentFrame, &ExitAbort)
		if !fromSpan {
			ScreenClearMsgs(false)
		}

	case CmdFileGlobalInput:
		if !checkSlotAllocation(FgiFile, false, &status) {
			goto l99
		}
		if !getFreeSlot(&fileSlot, fileSlot, &status) {
			goto l99
		}
		if Files[fileSlot] == nil {
			if !getFileName(tparam, &fnm, command) {
				goto l99
			}
			var dummyFptr *FileObject
			if !FileCreateOpen(&fnm, ParseInput, &Files[fileSlot], &dummyFptr) {
				goto l99
			}
		}
		FgiFile = fileSlot

	case CmdFileEdit:
		if !checkSlotAllocation(CurrentFrame.InputFile, false, &status) {
			goto l99
		}
		if !checkSlotAllocation(CurrentFrame.OutputFile, false, &status) {
			goto l99
		}
		if !getFreeSlot(&fileSlot, fileSlot, &status) {
			goto l99
		}
		var fileSlot2 int
		if !getFreeSlot(&fileSlot2, fileSlot, &status) {
			goto l99
		}
		if !getFileName(tparam, &fnm, command) {
			goto l99
		}
		if !FileCreateOpen(&fnm, ParseEdit, &Files[fileSlot], &Files[fileSlot2]) {
			goto l99
		}
		CurrentFrame.InputFile = fileSlot
		FilesFrames[fileSlot] = CurrentFrame
		CurrentFrame.OutputFile = fileSlot2
		FilesFrames[fileSlot2] = CurrentFrame
		if !fromSpan {
			ScreenMessage(MsgLoadingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		FilePage(CurrentFrame, &ExitAbort)
		if !fromSpan {
			ScreenClearMsgs(false)
		}

	case CmdFileExecute:
		if !checkSlotAllocation(CurrentFrame.InputFile, false, &status) {
			goto l99
		}
		if !getFreeSlot(&fileSlot, fileSlot, &status) {
			goto l99
		}
		if !getFileName(tparam, &fnm, command) {
			goto l99
		}
		var dummyFptr *FileObject
		if !FileCreateOpen(&fnm, ParseExecute, &Files[fileSlot], &dummyFptr) {
			goto l99
		}
		CurrentFrame.InputFile = fileSlot
		FilesFrames[fileSlot] = CurrentFrame
		FilePage(CurrentFrame, &ExitAbort)
		if !freeFile(fileSlot, &status) {
			goto l99
		}
		if !FileCloseDelete(Files[fileSlot], true, false) {
			goto l99
		}

	case CmdFileClose:
		switch savedCmd {
		case CmdFileInput:
			fileSlot = CurrentFrame.InputFile
		case CmdFileOutput:
			fileSlot = CurrentFrame.OutputFile
		case CmdFileGlobalInput:
			fileSlot = FgiFile
		case CmdFileGlobalOutput:
			fileSlot = FgoFile
		case CmdFileEdit:
			fileSlot = CurrentFrame.InputFile
		}
		if savedCmd == CmdFileOutput || savedCmd == CmdFileEdit {
			if !FileWindthru(CurrentFrame, fromSpan) {
				goto l99
			}
			ScreenFixup()
		}
		if !freeFile(fileSlot, &status) {
			goto l99
		}
		if savedCmd == CmdFileGlobalInput || savedCmd == CmdFileGlobalOutput {
			if !FileCloseDelete(Files[fileSlot], false, true) {
				goto l99
			}
		} else {
			if !FileCloseDelete(Files[fileSlot], !CurrentFrame.TextModified,
				CurrentFrame.TextModified || !Files[fileSlot].OutputFlag) {
				goto l99
			}
		}
		if savedCmd == CmdFileEdit {
			fileSlot = CurrentFrame.OutputFile
			if !freeFile(fileSlot, &status) {
				goto l99
			}
			if !FileCloseDelete(Files[fileSlot], !CurrentFrame.TextModified, CurrentFrame.TextModified) {
				goto l99
			}
		}
		if savedCmd == CmdFileOutput || savedCmd == CmdFileEdit {
			CurrentFrame.TextModified = false
		}

	case CmdFileKill:
		fileSlot = CurrentFrame.OutputFile
		if !freeFile(fileSlot, &status) {
			goto l99
		}
		if !FileCloseDelete(Files[fileSlot], true, true) {
			goto l99
		}

	case CmdFileGlobalKill:
		fileSlot = FgoFile
		if !freeFile(fileSlot, &status) {
			goto l99
		}
		if !FileCloseDelete(Files[fileSlot], true, true) {
			goto l99
		}

	case CmdFileOutput:
		if !checkSlotAllocation(CurrentFrame.OutputFile, false, &status) {
			goto l99
		}
		if !getFreeSlot(&fileSlot, fileSlot, &status) {
			goto l99
		}
		if !getFileName(tparam, &fnm, command) {
			goto l99
		}
		if CurrentFrame.InputFile != 0 {
			if !FileCreateOpen(&fnm, ParseOutput, &Files[CurrentFrame.InputFile], &Files[fileSlot]) {
				goto l99
			}
		} else {
			var dummyFptr *FileObject
			if !FileCreateOpen(&fnm, ParseOutput, &dummyFptr, &Files[fileSlot]) {
				goto l99
			}
		}
		CurrentFrame.OutputFile = fileSlot
		FilesFrames[fileSlot] = CurrentFrame

	case CmdFileGlobalOutput:
		if !checkSlotAllocation(FgoFile, false, &status) {
			goto l99
		}
		if !getFreeSlot(&fileSlot, fileSlot, &status) {
			goto l99
		}
		if Files[fileSlot] == nil {
			if !getFileName(tparam, &fnm, command) {
				goto l99
			}
			var dummyFptr *FileObject
			if !FileCreateOpen(&fnm, ParseOutput, &dummyFptr, &Files[fileSlot]) {
				goto l99
			}
		}
		FgoFile = fileSlot

	case CmdFileRead:
		if !checkSlotAllocation(FgiFile, true, &status) {
			goto l99
		}
		linesToRead := count
		if rept == LeadParamPIndef {
			linesToRead = MaxInt
		}
		var first, last *LineHdrObject
		var i int
		if !FileRead(Files[FgiFile], linesToRead, rept == LeadParamPIndef, &first, &last, &i) {
			goto l99
		}
		if first != nil {
			if !LinesInject(first, last, CurrentFrame.Dot.Line) {
				goto l99
			}
			if !MarkCreate(first, 1, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
				goto l99
			}
			CurrentFrame.TextModified = true
			if !MarkCreate(last.FLink, 1, &CurrentFrame.Marks[MarkModified-MinMarkNumber]) {
				goto l99
			}
			if !MarkCreate(last.FLink, 1, &CurrentFrame.Dot) {
				goto l99
			}
		}

	case CmdFileWrite:
		if !checkSlotAllocation(FgoFile, true, &status) {
			goto l99
		}
		if !fromSpan {
			ScreenMessage(MsgWritingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		var first, last *LineHdrObject
		if !ExecComputeLineRange(CurrentFrame, rept, count, &first, &last) {
			goto l99
		}
		if first != nil {
			if !FileWrite(first, last, Files[FgoFile]) {
				goto l99
			}
		}
		if !fromSpan {
			ScreenClearMsgs(false)
		}

	case CmdFileRewind:
		if !checkSlotDirection(CurrentFrame.InputFile, false, &status) {
			goto l99
		}
		if !FileRewind(&Files[CurrentFrame.InputFile]) {
			goto l99
		}
		if !fromSpan {
			ScreenMessage(MsgLoadingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		FilePage(CurrentFrame, &ExitAbort)
		if !fromSpan {
			ScreenClearMsgs(false)
		}

	case CmdFileGlobalRewind:
		if !checkSlotDirection(FgiFile, false, &status) {
			goto l99
		}
		if !FileRewind(&Files[FgiFile]) {
			goto l99
		}

	case CmdFileSave:
		if CurrentFrame.OutputFile == 0 {
			status = MsgNoOutput
			goto l99
		}
		if !CurrentFrame.TextModified {
			if !fromSpan {
				ScreenMessage(MsgNotModified)
				if LudwigMode == LudwigScreen {
					VduFlush()
				}
			}
			result = true
			goto l99
		}
		if !fromSpan {
			ScreenMessage(MsgSavingFile)
			if LudwigMode == LudwigScreen {
				VduFlush()
			}
		}
		linesWritten := Files[CurrentFrame.OutputFile].LCounter
		first := CurrentFrame.FirstGroup.FirstLine
		last := CurrentFrame.LastGroup.LastLine.BLink
		if last != nil {
			if !FileWrite(first, last, Files[CurrentFrame.OutputFile]) {
				goto l99
			}
		}
		var dummyFptr *FileObject
		if CurrentFrame.InputFile != 0 {
			dummyFptr = Files[CurrentFrame.InputFile]
		}
		if !FilesysSave(dummyFptr, Files[CurrentFrame.OutputFile], linesWritten) {
			goto l99
		}
		var nrLines int
		if last == nil {
			nrLines = 0
		} else if !LineToNumber(last, &nrLines) {
			nrLines = 0
		}
		CurrentFrame.InputCount = uint32(Files[CurrentFrame.OutputFile].LCounter + nrLines)
		if CurrentFrame.InputFile != 0 {
			Files[CurrentFrame.InputFile].LCounter = int(CurrentFrame.InputCount)
		}
		CurrentFrame.TextModified = false
	}

	result = true

l99:
	if status != "" {
		ScreenMessage(status)
	}
	return result
}
