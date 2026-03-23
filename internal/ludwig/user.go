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
func UserKeyCodeToName(keyCode int) (string, bool) {
	for i := 1; i < len(KeyNameList); i++ {
		if KeyNameList[i].KeyCode == keyCode {
			return KeyNameList[i].KeyName, true
		}
	}
	return "", false
}

// UserKeyNameToCode converts a key name to its code
func UserKeyNameToCode(keyName string) (int, bool) {
	for i := 1; i < len(KeyNameList); i++ {
		if KeyNameList[i].KeyName == keyName {
			return KeyNameList[i].KeyCode, true
		}
	}
	return -1, false
}

// UserKeyInitialize initializes terminal-defined key map table
func UserKeyInitialize() {
	VduKeyboardInit(&NrKeyNames, &KeyNameList, KeyIntroducers, &TerminalInfo)

	if keyCode, found := UserKeyNameToCode("UP-ARROW"); found {
		Lookup[keyCode].Command = CmdUp
	}
	if keyCode, found := UserKeyNameToCode("DOWN-ARROW"); found {
		Lookup[keyCode].Command = CmdDown
	}
	if keyCode, found := UserKeyNameToCode("RIGHT-ARROW"); found {
		Lookup[keyCode].Command = CmdRight
	}
	if keyCode, found := UserKeyNameToCode("LEFT-ARROW"); found {
		Lookup[keyCode].Command = CmdLeft
	}
	if keyCode, found := UserKeyNameToCode("HOME"); found {
		Lookup[keyCode].Command = CmdHome
	}
	if keyCode, found := UserKeyNameToCode("BACK-TAB"); found {
		Lookup[keyCode].Command = CmdBacktab
	}
	if keyCode, found := UserKeyNameToCode("INSERT-CHAR"); found {
		Lookup[keyCode].Command = CmdInsertChar
	}
	if keyCode, found := UserKeyNameToCode("DELETE-CHAR"); found {
		Lookup[keyCode].Command = CmdDeleteChar
	}
	if keyCode, found := UserKeyNameToCode("INSERT-LINE"); found {
		Lookup[keyCode].Command = CmdInsertLine
	}
	if keyCode, found := UserKeyNameToCode("DELETE-LINE"); found {
		Lookup[keyCode].Command = CmdDeleteLine
	}
	if keyCode, found := UserKeyNameToCode("HELP"); found {
		Lookup[keyCode].Command = CmdHelp
		tpar := &TParObject{
			Str: EmptyStrObject(),
			Dlm: TpdPrompt,
		}
		Lookup[keyCode].Tpar = tpar
	}
	if keyCode, found := UserKeyNameToCode("FIND"); found {
		Lookup[keyCode].Command = CmdGet
		tpar := &TParObject{
			Str: EmptyStrObject(),
			Dlm: TpdPrompt,
		}
		Lookup[keyCode].Tpar = tpar
	}
	if keyCode, found := UserKeyNameToCode("PREV-SCREEN"); found {
		Lookup[keyCode].Command = CmdWindowBackward
	}
	if keyCode, found := UserKeyNameToCode("NEXT-SCREEN"); found {
		Lookup[keyCode].Command = CmdWindowForward
	}
	if keyCode, found := UserKeyNameToCode("PAGE-UP"); found {
		Lookup[keyCode].Command = CmdWindowBackward
	}
	if keyCode, found := UserKeyNameToCode("PAGE-DOWN"); found {
		Lookup[keyCode].Command = CmdWindowForward
	}
	if keyCode, found := UserKeyNameToCode("WINDOW-RESIZE-EVENT"); found {
		Lookup[keyCode].Command = CmdResizeWindow
	}
}

// UserCommandIntroducer enters command introducer into text in correct keyboard mode
func UserCommandIntroducer() bool {
	if !ChIsPrintable(rune(CommandIntroducer)) {
		ScreenMessage(MsgNonprintableIntroducer)
		return false
	}

	temp := NewBlankStrObject(MaxStrLen)
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
		MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified])
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
		var found bool
		if keyCode, found = UserKeyNameToCode(keyName); !found {
			ScreenMessage(MsgUnrecognizedKeyName)
			return false
		}
	}

	// Create a span in frame "HEAP"
	MarkCreate(FrameHeap.LastGroup.LastLine, 1, &FrameHeap.Span.MarkTwo)
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
				Lookup[keyCode].Tpar = nil

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
