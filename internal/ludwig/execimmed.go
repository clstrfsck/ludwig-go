/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         EXECIMMED
//
// Description:  Outermost level of command execution for LUDWIG.

package ludwig

import "fmt"

const (
	// Editors Extraordinaire
	defaultSpanName = "L. Wittgenstein und Sohn."
)

func executeImmedScreen(frame *FrameObject) {
	cmdSpan := SpanObject{
		Name: defaultSpanName,
	}
	jammed := false
	var cmdSuccess bool
outerLoop:
	for {
		cmdSuccess = true
		skipExec := false

		// MAKE SURE THE USER CAN SEE THE CURRENT DOT POSITION.
		ScreenFixup(frame)

		var key int
		if EditMode == ModeCommand {
			key = CommandIntroducer
		} else {
			// OVERTYPE/INSERTMODE IS DONE HERE AS A SPECIAL CASE.
			// THIS IS NECESSARY BECAUSE THE SCREEN IS UPDATED BY
			// VDU_GET_TEXT.
			for {
				// Check for boundaries where text cannot be accepted.
				if jammed || frame.Dot.Col == MaxStrLenP {
					key = VduGetKey()
					if TtControlC {
						skipExec = true
						break
					}
					if ChIsPrintable(rune(key)) && key != CommandIntroducer {
						cmdSuccess = false
						skipExec = true
						break
					}
					VduTakeBackKey(key)
					break
				}

				// DECIDE MAX CHARS THAT CAN BE READ.
				tmpScrCol := frame.Dot.Col - frame.ScrOffset
				inputLen := MaxStrLenP - frame.Dot.Col
				if frame.Dot.Col <= frame.MarginRight {
					inputLen = frame.MarginRight - frame.Dot.Col + 1
				}
				if inputLen > frame.ScrWidth+1-tmpScrCol {
					inputLen = frame.ScrWidth + 1 - tmpScrCol
				}

				// WATCH OUT FOR NULL LINE.
				if frame.Dot.Line.FLink == nil {
					key = VduGetKey()
					if TtControlC {
						skipExec = true
						break
					}
					VduTakeBackKey(key)
					if ChIsPrintable(rune(key)) && key != CommandIntroducer {
						// If printing char, realize NULL, re-fix cursor.
						TextRealizeNull(frame.Dot.Line)
					}

					// MAKE SURE THE USER CAN SEE THE CURRENT DOT POSITION.
					ScreenFixup(frame)
				}

				// GET THE ECHOING TEXT
				if EditMode == ModeInsert {
					VduInsertMode(true)
				}
				inputBuf := NewBlankStrObject(MaxStrLen)
				inputLen = VduGetText(inputLen, inputBuf)
				if EditMode == ModeInsert {
					VduInsertMode(false)
					VduFlush() // Make sure in mode IS off!
				}
				if TtControlC {
					skipExec = true
					break
				}
				if inputLen == 0 {
					break // Simulate a continue
				}

				if EditMode == ModeOvertype {
					cmdSuccess = TextOvertype(false, 1, inputBuf, inputLen, frame.Dot)
				} else {
					cmdSuccess = TextInsert(false, 1, inputBuf, inputLen, frame.Dot)
				}
				if cmdSuccess {
					frame.TextModified = true
					MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
					MarkCreate(frame.Dot.Line, frame.Dot.Col-inputLen, &frame.Marks[MarkEquals])
				} else {
					// IF, FOR SOME REASON, THAT FAILED, CORRECT THE VDU IMAGE OF
					// THE LINE. THIS IS BECAUSE VDU_GET_TEXT HAS CORRUPTED IT.
					ScreenDrawLine(frame.Dot.Line)
					skipExec = true
					break
				}
				if frame.Dot.Col != frame.MarginRight+1 {
					// FOLLOW THE DOT.
					ScreenPosition(frame.Dot.Line, frame.Dot.Col)
					VduMoveCurs(
						frame.Dot.Col-frame.ScrOffset,
						frame.Dot.Line.ScrRowNr,
					)
				} else {
					// AT THE RIGHT MARGIN.
					if frame.Options.Has(OptAutoWrap) {
						// Take care of Wrap Option.
						key = VduGetKey()
						if TtControlC {
							skipExec = true
							break
						}
						if ChIsPrintable(rune(key)) && key != CommandIntroducer {
							col1 := frame.MarginRight
							if key != ' ' {
								for frame.Dot.Line.Str.Get(col1) != ' ' &&
									col1 > frame.MarginLeft {
									col1--
								}
								col2 := col1
								for frame.Dot.Line.Str.Get(col2) == ' ' &&
									col2 > frame.MarginLeft {
									col2--
								}
								if col2 == frame.MarginLeft { // Line has only one word
									col1 = frame.MarginRight // Split at right margin
								}
								VduTakeBackKey(key)
							}
							frame.Dot.Col = col1 + 1
							TextSplitLine(frame.Dot, 0, &frame.Marks[MarkEquals])
							frame.Dot.Col += frame.MarginRight - col1
							continue outerLoop
						}
						VduTakeBackKey(key)
					} else {
						VduBeep()
						frame.Dot.Col--
						VduMoveCurs(
							frame.Dot.Col-frame.ScrOffset,
							frame.Dot.Line.ScrRowNr,
						)
						jammed = true
					}
				}
			} // of overtyping loop

			if !skipExec {
				key = VduGetKey() // key is a terminator
				if TtControlC {
					skipExec = true
				}
			}

			// DEBUG code removed - would check for printable characters
		}

		if !skipExec {
			if key == CommandIntroducer {
				var ok bool
				frame, ok = CodeCompile(frame, &cmdSpan, false)
				if ok {
					frame, cmdSuccess = CodeInterpret(frame, LeadParamNone, 1, cmdSpan.Code, false)
				} else {
					cmdSuccess = false
				}
			} else {
				if Lookup[key].Command == CmdExtended {
					frame, cmdSuccess = CodeInterpret(frame, LeadParamNone, 1, Lookup[key].Code, true)
				} else {
					frame, cmdSuccess = Execute(
						frame, Lookup[key].Command, LeadParamNone, 1, Lookup[key].Tpar, false,
					)
				}
			}
		}

		if TtControlC {
			TtControlC = false
			if frame.Dot.Line.ScrRowNr != 0 {
				ScreenRedraw()
			} else {
				ScreenUnload()
			}
		} else if !cmdSuccess {
			VduBeep()  // Complain.
			VduFlush() // Make sure they hear the complaint.
		} else {
			jammed = false
		}
		ExitAbort = false
	}
}

func executeImmedBatchHardcopy(frame *FrameObject) {
	cmdSpan := SpanObject{
		Name: defaultSpanName,
	}
	// Allocate marks for the command span
	cmdSpan.MarkOne = &MarkObject{}
	cmdSpan.MarkOne.Line = nil
	cmdSpan.MarkOne.Col = 1
	cmdSpan.MarkTwo = &MarkObject{}
	cmdSpan.MarkTwo.Line = nil
	var cmdCount int
	if LudwigMode == LudwigHardcopy {
		cmdCount = 1
	} else {
		cmdCount = MaxInt
	}

	// Open standard input as Ludwig command input file.
	var cmdFile *FileObject
	var dummyFptr *FileObject
	if FileCreateOpen(nil, ParseStdin, &cmdFile, &dummyFptr) {
		for {
			// Destroy all of cmd_span's contents.
			if cmdSpan.MarkOne.Line != nil {
				cmdSpan.MarkOne.Line = nil
				cmdSpan.MarkTwo.Line = nil
			}

			// If necessary, prompt.
			if LudwigMode == LudwigHardcopy {
				ScreenLoad(frame.Dot.Line)
				fmt.Println("COMMAND: ")
			}

			// Read, compile, and execute the next lot of commands.
			var ok bool
			if cmdSpan.MarkOne.Line, cmdSpan.MarkTwo.Line, _, ok = FileRead(
				cmdFile,
				cmdCount,
				true,
			); ok {
				if cmdSpan.MarkOne.Line != nil {
					cmdSpan.MarkTwo.Col = cmdSpan.MarkTwo.Line.Used + 1
					frame, ok = CodeCompile(frame, &cmdSpan, true)
					if ok {
						frame, ok = CodeInterpret(frame, LeadParamNone, 1, cmdSpan.Code, true)
						if !ok {
							fmt.Println("\aCOMMAND FAILED")
						}
					}
					ExitAbort = false
					TtControlC = false
				}
			}
			if cmdFile.Eof {
				break
			}
		}
		LudwigAborted = false
		QuitCloseFiles()
	}
}

// ExecuteImmed is the main execution loop for Ludwig
func ExecuteImmed(frame *FrameObject) {
	// Vector off to the appropriate main execution mode. Each mode behaves
	// slightly differently at this level.
	switch LudwigMode {
	case LudwigScreen:
		executeImmedScreen(frame)

	case LudwigHardcopy, LudwigBatch:
		executeImmedBatchHardcopy(frame)
	}
}
