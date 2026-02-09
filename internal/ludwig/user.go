/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         USER
//
// Description:  The user commands (UC, UK, UP, US, and UU).

package ludwig

// specialCommand checks if a command is a special command
func specialCommand(cmd Commands) bool {
	return cmd == CmdVerify || cmd == CmdExitAbort ||
		cmd == CmdExitFail || cmd == CmdExitSuccess
}

// UserKeyCodeToName converts a key code to its name
func UserKeyCodeToName(keyCode int, keyName *string) bool {
	for i := 1; i < len(KeyNameList); i++ {
		if KeyNameList[i].KeyCode == keyCode {
			*keyName = KeyNameList[i].KeyName
			return true
		}
	}
	return false
}

// UserKeyNameToCode converts a key name to its code
func UserKeyNameToCode(keyName string, keyCode *int) bool {
	for i := 1; i < len(KeyNameList); i++ {
		if KeyNameList[i].KeyName == keyName {
			*keyCode = KeyNameList[i].KeyCode
			return true
		}
	}
	return false
}

// UserKeyInitialize initializes terminal-defined key map table
func UserKeyInitialize() {
	var keyCode int
	VduKeyboardInit(&NrKeyNames, &KeyNameList, &KeyIntroducers, &TerminalInfo)

	if UserKeyNameToCode("UP-ARROW", &keyCode) {
		Lookup[keyCode].Command = CmdUp
	}
	if UserKeyNameToCode("DOWN-ARROW", &keyCode) {
		Lookup[keyCode].Command = CmdDown
	}
	if UserKeyNameToCode("RIGHT-ARROW", &keyCode) {
		Lookup[keyCode].Command = CmdRight
	}
	if UserKeyNameToCode("LEFT-ARROW", &keyCode) {
		Lookup[keyCode].Command = CmdLeft
	}
	if UserKeyNameToCode("HOME", &keyCode) {
		Lookup[keyCode].Command = CmdHome
	}
	if UserKeyNameToCode("BACK-TAB", &keyCode) {
		Lookup[keyCode].Command = CmdBacktab
	}
	if UserKeyNameToCode("INSERT-CHAR", &keyCode) {
		Lookup[keyCode].Command = CmdInsertChar
	}
	if UserKeyNameToCode("DELETE-CHAR", &keyCode) {
		Lookup[keyCode].Command = CmdDeleteChar
	}
	if UserKeyNameToCode("INSERT-LINE", &keyCode) {
		Lookup[keyCode].Command = CmdInsertLine
	}
	if UserKeyNameToCode("DELETE-LINE", &keyCode) {
		Lookup[keyCode].Command = CmdDeleteLine
	}
	if UserKeyNameToCode("HELP", &keyCode) {
		Lookup[keyCode].Command = CmdHelp
		tpar := &TParObject{Str: *NewFilled(' ', MaxStrLen)}
		tpar.Dlm = TpdPrompt
		tpar.Len = 0
		tpar.Con = nil
		tpar.Nxt = nil
		Lookup[keyCode].Tpar = tpar
	}
	if UserKeyNameToCode("FIND", &keyCode) {
		Lookup[keyCode].Command = CmdGet
		tpar := &TParObject{Str: *NewFilled(' ', MaxStrLen)}
		tpar.Dlm = TpdPrompt
		tpar.Len = 0
		tpar.Con = nil
		tpar.Nxt = nil
		Lookup[keyCode].Tpar = tpar
	}
	if UserKeyNameToCode("PREV-SCREEN", &keyCode) {
		Lookup[keyCode].Command = CmdWindowBackward
	}
	if UserKeyNameToCode("NEXT-SCREEN", &keyCode) {
		Lookup[keyCode].Command = CmdWindowForward
	}
	if UserKeyNameToCode("PAGE-UP", &keyCode) {
		Lookup[keyCode].Command = CmdWindowBackward
	}
	if UserKeyNameToCode("PAGE-DOWN", &keyCode) {
		Lookup[keyCode].Command = CmdWindowForward
	}
	if UserKeyNameToCode("WINDOW-RESIZE-EVENT", &keyCode) {
		Lookup[keyCode].Command = CmdResizeWindow
	}
}

// UserCommandIntroducer enters command introducer into text in correct keyboard mode
func UserCommandIntroducer() bool {
	if !ChIsPrintable(rune(CommandIntroducer)) {
		ScreenMessage(MsgNonprintableIntroducer)
		return false
	}

	temp := NewFilled(' ', MaxStrLen)
	temp.Set(1, byte(CommandIntroducer))
	cmdSuccess := true

	switch EditMode {
	case ModeInsert:
		cmdSuccess = TextInsert(true, 1, temp, 1, CurrentFrame.Dot)
	case ModeCommand:
		if PreviousMode == ModeInsert {
			cmdSuccess = TextInsert(true, 1, temp, 1, CurrentFrame.Dot)
		} else {
			cmdSuccess = TextOvertype(true, 1, temp, 1, CurrentFrame.Dot)
		}
	case ModeOvertype:
		cmdSuccess = TextOvertype(true, 1, temp, 1, CurrentFrame.Dot)
	}

	if cmdSuccess {
		CurrentFrame.TextModified = true
		if !MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkModified-MinMarkNumber],
		) {
			cmdSuccess = false
		}
	}
	return cmdSuccess
}

// UserKey assigns a key to a command string
func UserKey(key *TParObject, strng *TParObject) bool {
	result := false
	var keyCode int

	if key.Len == 1 {
		keyCode = int(key.Str.Get(1))
	} else {
		keyName := key.Str.Slice(1, key.Len)
		if !UserKeyNameToCode(keyName, &keyCode) {
			ScreenMessage(MsgUnrecognizedKeyName)
			return false
		}
	}

	// Create a span in frame "HEAP"
	if !MarkCreate(FrameHeap.LastGroup.LastLine, 1, &FrameHeap.Span.MarkTwo) {
		return false
	}
	if !SpanCreate(BlankFrameName, FrameHeap.Span.MarkTwo, FrameHeap.Span.MarkTwo) {
		return false
	}

	var keySpan *SpanObject
	var oldSpan *SpanObject
	if SpanFind(BlankFrameName, &keySpan, &oldSpan) {
		success := false
		var keyMarkOne *MarkObject = keySpan.MarkOne
		if TextInsertTpar(strng, keySpan.MarkTwo, &keyMarkOne) {
			if CodeCompile(keySpan, true) {
				// discard code_ptr, if it exists, NOW!
				if Lookup[keyCode].Code != nil {
					CodeDiscard(&Lookup[keyCode].Code)
				}
				if Lookup[keyCode].Tpar != nil {
					TparCleanObject(Lookup[keyCode].Tpar)
					Lookup[keyCode].Tpar = nil
				}

				code := keySpan.Code
				if (code.Len == 2) && (CompilerCode[code.Code].Rep == LeadParamNone) &&
					!specialCommand(CompilerCode[code.Code].Op) {
					// simple command, put directly into lookup table
					Lookup[keyCode].Command = CompilerCode[code.Code].Op
					Lookup[keyCode].Tpar = CompilerCode[code.Code].Tpar
					CompilerCode[code.Code].Tpar = nil
				} else {
					Lookup[keyCode].Command = CmdExtended
					Lookup[keyCode].Code = code
					keySpan.Code = nil
				}
				success = true
			}
		}
		SpanDestroy(&keySpan)
		result = success
	}
	return result
}

// UserParent suspends Ludwig and returns to parent shell
func UserParent() bool {
	return SysSuspend()
}

// UserSubprocess spawns a subshell
func UserSubprocess() bool {
	return SysShell()
}

// UserUndo performs undo operation (not implemented)
func UserUndo() bool {
	ScreenMessage(MsgNotImplemented)
	return false
}
