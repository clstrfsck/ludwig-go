/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         QUIT
//
// Description:  Quit Ludwig

package ludwig

const (
	noOutputFileMsg = "This frame has no output file--are you sure you want to QUIT? "
)

// QuitCommand handles the quit command
func QuitCommand() bool {
	if LudwigMode != LudwigBatch {
		newSpan := FirstSpan
		for newSpan != nil {
			if newSpan.Frame != nil {
				if newSpan.Frame.TextModified && newSpan.Frame.OutputFile == 0 &&
					newSpan.Frame.InputFile != 0 {
					CurrentFrame = newSpan.Frame
					mm := newSpan.Frame.Marks[MarkModified]
					MarkCreate(mm.Line, mm.Col, &newSpan.Frame.Dot)
					if LudwigMode == LudwigScreen {
						ScreenFixup()
					}
					ScreenBeep()
					switch ScreenVerify(noOutputFileMsg) {
					case VerifyReplyYes:
						// Nothing to do here
					case VerifyReplyAlways:
						goto l2
					case VerifyReplyNo, VerifyReplyQuit:
						ExitAbort = true
						return false
					}
				}
			}
			newSpan = newSpan.FLink
		}
	}
l2:
	ScreenUnload()
	if LudwigMode != LudwigBatch {
		ScreenMessage(MsgQuitting)
	}
	if LudwigMode == LudwigScreen {
		VduFlush()
	}
	LudwigAborted = false
	QuitCloseFiles()
	SysExitSuccess()
	return true // Given the exit above, this shouldn't happen
}

// doFrame handles closing files for a single frame
func doFrame(f *FrameObject) bool {
	if f.OutputFile == 0 {
		return true
	}
	if Files[f.OutputFile] == nil {
		return true
	}
	// Wind out and close the associated input file
	if !FileWindthru(f, true) {
		return false
	}
	if f.InputFile != 0 {
		if Files[f.InputFile] != nil {
			if !FileCloseDelete(Files[f.InputFile], false, true) {
				return false
			}
			f.InputFile = 0
		}
	}
	// Close the output file
	result := true
	if !LudwigAborted {
		result = FileCloseDelete(Files[f.OutputFile], !f.TextModified, f.TextModified)
	}
	f.OutputFile = 0
	return result
}

// QuitCloseFiles closes all files during quit
// THIS ROUTINE DOES BOTH THE NORMAL "Q" COMMAND, AND ALSO IS CALLED AS PART
// OF THE LUDWIG "PROG_WINDUP" SEQUENCE. THUS BY TYPING "^Y EXIT" USERS MAY
// SAFELY ABORT LUDWIG AND NOT LOSE ANY FILES.
func QuitCloseFiles() {
	nextSpan := FirstSpan
	for nextSpan != nil {
		nextFrame := nextSpan.Frame
		if nextFrame != nil {
			if !doFrame(nextFrame) {
				goto l99
			}
		}
		nextSpan = nextSpan.FLink
	}

	// Close all remaining files
	if !LudwigAborted {
		for fileIndex := 1; fileIndex <= MaxFiles; fileIndex++ {
			if Files[fileIndex] != nil {
				if !FileCloseDelete(Files[fileIndex], false, true) {
					goto l99
				}
			}
		}
	}
l99:
	// Now free up the VDU, thus re-setting anything we have changed
	if !VduFreeFlag { // Has it been called already?
		VduFree()
		VduFreeFlag = true // Well it has now
		LudwigMode = LudwigBatch
	}
	if LudwigAborted {
		ScreenMessage(MsgNotRenamed)
		ScreenMessage(MsgAbort)
	}
}
