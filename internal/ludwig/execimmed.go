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

import (
	"fmt"
)

const (
	// Editors Extraordinaire
	defaultSpanName = "L. Wittgenstein und Sohn."
)

// ExecuteImmed is the main execution loop for Ludwig
func ExecuteImmed() {
	var cmdSpan SpanObject
	cmdSpan.FLink = nil
	cmdSpan.BLink = nil
	cmdSpan.Name = defaultSpanName
	cmdSpan.Frame = nil
	cmdSpan.MarkOne = nil
	cmdSpan.MarkTwo = nil
	cmdSpan.Code = nil

	// Vector off to the appropriate main execution mode. Each mode behaves
	// slightly differently at this level.
	switch LudwigMode {
	case LudwigScreen:
		jammed := false
		var cmdSuccess bool
		for {
		l2:
			cmdSuccess = true

			// MAKE SURE THE USER CAN SEE THE CURRENT DOT POSITION.
			ScreenFixup()

			var key int
			if EditMode == ModeCommand {
				key = CommandIntroducer
			} else {
				// OVERTYPE/INSERTMODE IS DONE HERE AS A SPECIAL CASE.
				// THIS IS NECESSARY BECAUSE THE SCREEN IS UPDATED BY
				// VDU_GET_TEXT.
				for {
					// Check for boundaries where text cannot be accepted.
					if jammed || CurrentFrame.Dot.Col == MaxStrLenP {
						key = VduGetKey()
						if TtControlC {
							goto l9
						}
						if ChIsPrintable(rune(key)) && key != CommandIntroducer {
							cmdSuccess = false
							goto l9
						}
						VduTakeBackKey(key)
						break
					}

					// DECIDE MAX CHARS THAT CAN BE READ.
					tmpScrCol := CurrentFrame.Dot.Col - CurrentFrame.ScrOffset
					inputLen := MaxStrLenP - CurrentFrame.Dot.Col
					if CurrentFrame.Dot.Col <= CurrentFrame.MarginRight {
						inputLen = CurrentFrame.MarginRight - CurrentFrame.Dot.Col + 1
					}
					if inputLen > CurrentFrame.ScrWidth+1-tmpScrCol {
						inputLen = CurrentFrame.ScrWidth + 1 - tmpScrCol
					}

					// WATCH OUT FOR NULL LINE.
					if CurrentFrame.Dot.Line.FLink == nil {
						key = VduGetKey()
						if TtControlC {
							goto l9
						}
						VduTakeBackKey(key)
						if ChIsPrintable(rune(key)) && key != CommandIntroducer {
							// If printing char, realize NULL, re-fix cursor.
							if !TextRealizeNull(CurrentFrame.Dot.Line) {
								cmdSuccess = false
								goto l9
							}
						}

						// MAKE SURE THE USER CAN SEE THE CURRENT DOT POSITION.
						ScreenFixup()
					}

					// GET THE ECHOING TEXT
					if EditMode == ModeInsert {
						VduInsertMode(true)
					}
					inputBuf := NewBlankStrObject(MaxStrLen)
					VduGetText(inputLen, inputBuf, &inputLen)
					if EditMode == ModeInsert {
						VduInsertMode(false)
						VduFlush() // Make sure in mode IS off!
					}
					if TtControlC {
						goto l9
					}
					if inputLen == 0 {
						break // Simulate a continue
					}

					if EditMode == ModeOvertype {
						cmdSuccess = TextOvertype(false, 1, inputBuf, inputLen, CurrentFrame.Dot)
					} else {
						cmdSuccess = TextInsert(false, 1, inputBuf, inputLen, CurrentFrame.Dot)
					}
					if cmdSuccess {
						CurrentFrame.TextModified = true
						if !MarkCreate(
							CurrentFrame.Dot.Line,
							CurrentFrame.Dot.Col,
							&CurrentFrame.Marks[MarkModified],
						) {
							cmdSuccess = false
						}
						if !MarkCreate(
							CurrentFrame.Dot.Line,
							CurrentFrame.Dot.Col-inputLen,
							&CurrentFrame.Marks[MarkEquals],
						) {
							cmdSuccess = false
						}
					} else {
						// IF, FOR SOME REASON, THAT FAILED, CORRECT THE VDU IMAGE OF
						// THE LINE. THIS IS BECAUSE VDU_GET_TEXT HAS CORRUPTED IT.
						ScreenDrawLine(CurrentFrame.Dot.Line)
						goto l9
					}
					if CurrentFrame.Dot.Col != CurrentFrame.MarginRight+1 {
						// FOLLOW THE DOT.
						ScreenPosition(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col)
						VduMoveCurs(
							CurrentFrame.Dot.Col-CurrentFrame.ScrOffset,
							CurrentFrame.Dot.Line.ScrRowNr,
						)
					} else {
						// AT THE RIGHT MARGIN.
						if CurrentFrame.Options.Has(OptAutoWrap) {
							// Take care of Wrap Option.
							key = VduGetKey()
							if TtControlC {
								goto l9
							}
							if ChIsPrintable(rune(key)) && key != CommandIntroducer {
								col1 := CurrentFrame.MarginRight
								if key != ' ' {
									for CurrentFrame.Dot.Line.Str.Get(col1) != ' ' &&
										col1 > CurrentFrame.MarginLeft {
										col1--
									}
									col2 := col1
									for CurrentFrame.Dot.Line.Str.Get(col2) == ' ' &&
										col2 > CurrentFrame.MarginLeft {
										col2--
									}
									if col2 == CurrentFrame.MarginLeft { // Line has only one word
										col1 = CurrentFrame.MarginRight // Split at right margin
									}
									VduTakeBackKey(key)
								}
								CurrentFrame.Dot.Col++
								cmdSuccess = TextSplitLine(
									CurrentFrame.Dot, 0, &CurrentFrame.Marks[MarkEquals],
								)
								CurrentFrame.Dot.Col += CurrentFrame.MarginRight - col1
								goto l2 // Simulate break of inner loop
							}
							VduTakeBackKey(key)
						} else {
							VduBeep()
							CurrentFrame.Dot.Col--
							VduMoveCurs(
								CurrentFrame.Dot.Col-CurrentFrame.ScrOffset,
								CurrentFrame.Dot.Line.ScrRowNr,
							)
							jammed = true
						}
					}
				} // of overtyping loop

				key = VduGetKey() // key is a terminator
				if TtControlC {
					goto l9
				}

				// DEBUG code removed - would check for printable characters
			}

			if key == CommandIntroducer {
				if CodeCompile(&cmdSpan, false) {
					cmdSuccess = CodeInterpret(LeadParamNone, 1, cmdSpan.Code, false)
				} else {
					cmdSuccess = false
				}
			} else {
				if Lookup[key].Command == CmdExtended {
					cmdSuccess = CodeInterpret(LeadParamNone, 1, Lookup[key].Code, true)
				} else {
					cmdSuccess = Execute(
						Lookup[key].Command, LeadParamNone, 1, Lookup[key].Tpar, false,
					)
				}
			}

		l9:
			if TtControlC {
				TtControlC = false
				if CurrentFrame.Dot.Line.ScrRowNr != 0 {
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

	case LudwigHardcopy, LudwigBatch:
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
		var cmdFnm string
		var cmdFile *FileObject
		var dummyFptr *FileObject
		if FileCreateOpen(&cmdFnm, ParseStdin, &cmdFile, &dummyFptr) {
			for {
				// Destroy all of cmd_span's contents.
				if cmdSpan.MarkOne.Line != nil {
					if !LinesDestroy(&cmdSpan.MarkOne.Line, &cmdSpan.MarkTwo.Line) {
						return
					}
					cmdSpan.MarkOne.Line = nil
					cmdSpan.MarkTwo.Line = nil
				}

				// If necessary, prompt.
				if LudwigMode == LudwigHardcopy {
					ScreenLoad(CurrentFrame.Dot.Line)
					fmt.Println("COMMAND: ")
				}

				// Read, compile, and execute the next lot of commands.
				var i int
				if FileRead(
					cmdFile,
					cmdCount,
					true,
					&cmdSpan.MarkOne.Line,
					&cmdSpan.MarkTwo.Line,
					&i,
				) {
					if cmdSpan.MarkOne.Line != nil {
						cmdSpan.MarkTwo.Col = cmdSpan.MarkTwo.Line.Used + 1
						if CodeCompile(&cmdSpan, true) {
							if !CodeInterpret(LeadParamNone, 1, cmdSpan.Code, true) {
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
}
