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
func QuitCommand(frame *FrameObject) (*FrameObject, bool) {
	if LudwigMode != LudwigBatch {
		newSpan := FirstSpan
	spanLoop:
		for newSpan != nil {
			if newSpan.Frame != nil {
				if newSpan.Frame.TextModified && newSpan.Frame.OutputFile == 0 &&
					newSpan.Frame.InputFile != 0 {
					frame = newSpan.Frame
					MarkCreate(
						newSpan.Frame.Marks[MarkModified].Line,
						newSpan.Frame.Marks[MarkModified].Col,
						&newSpan.Frame.Dot,
					)
					if LudwigMode == LudwigScreen {
						ScreenFixup(&Screen, frame)
					}
					ScreenBeep()
					switch ScreenVerify(&Screen, frame, noOutputFileMsg) {
					case VerifyReplyYes:
						// Nothing to do here
					case VerifyReplyAlways:
						break spanLoop
					case VerifyReplyNo, VerifyReplyQuit:
						ExitAbort = true
						return frame, false
					}
				}
			}
			newSpan = newSpan.FLink
		}
	}
	ScreenUnload(&Screen)
	if LudwigMode != LudwigBatch {
		ScreenMessage(&Screen, MsgQuitting)
	}
	if LudwigMode == LudwigScreen {
		VduFlush()
	}
	LudwigAborted = false
	QuitCloseFiles()
	SysExitSuccess()
	return frame, true // Given the exit above, this shouldn't happen
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
			if !FileCloseDelete(&Files[f.InputFile], false, true) {
				return false
			}
			f.InputFile = 0
		}
	}
	// Close the output file
	result := true
	if !LudwigAborted {
		result = FileCloseDelete(&Files[f.OutputFile], !f.TextModified, f.TextModified)
	}
	f.OutputFile = 0
	return result
}

func closeAllFiles() {
	nextSpan := FirstSpan
	for nextSpan != nil {
		nextFrame := nextSpan.Frame
		if nextFrame != nil {
			if !doFrame(nextFrame) {
				return
			}
		}
		nextSpan = nextSpan.FLink
	}

	// Close all remaining files
	if !LudwigAborted {
		for fileIndex := 1; fileIndex <= MaxFiles; fileIndex++ {
			if Files[fileIndex] != nil {
				if !FileCloseDelete(&Files[fileIndex], false, true) {
					return
				}
			}
		}
	}
}

// QuitCloseFiles closes all files during quit
// THIS ROUTINE DOES BOTH THE NORMAL "Q" COMMAND, AND ALSO IS CALLED AS PART
// OF THE LUDWIG "PROG_WINDUP" SEQUENCE. THUS BY TYPING "^Y EXIT" USERS MAY
// SAFELY ABORT LUDWIG AND NOT LOSE ANY FILES.
func QuitCloseFiles() {
	closeAllFiles()

	// Now free up the VDU, thus re-setting anything we have changed
	if !VduFreeFlag { // Has it been called already?
		VduFree()
		VduFreeFlag = true // Well it has now
		LudwigMode = LudwigBatch
	}
	if LudwigAborted {
		ScreenMessage(&Screen, MsgNotRenamed)
		ScreenMessage(&Screen, MsgAbort)
	}
}
