/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         VALUE
//
// Description:  Initialization of global variables.

package ludwig

// setupInitialValues initializes all global variables to their default values
func setupInitialValues() {
	CurrentFrame = nil
	LudwigAborted = false
	ExitAbort = false
	Hangup = false
	EditMode = ModeInsert
	PreviousMode = ModeInsert

	for i := 1; i <= MaxFiles; i++ {
		Files[i] = nil
		FilesFrames[i] = nil
	}

	FgiFile = 0
	FgoFile = 0
	FirstSpan = nil
	LudwigMode = LudwigBatch
	CommandIntroducer = '\\'
	ScrFrame = nil
	ScrMsgRow = MaxInt
	VduFreeFlag = false
	ExecLevel = 0

	// Set up all the Default Default characteristics for a frame
	for i := 0; i <= MaxMarkNumber; i++ {
		InitialMarks[i] = nil
	}

	InitialScrHeight = 1  // Set to tt_height for terminals
	InitialScrWidth = 132 // Set to tt_width for terminals
	InitialScrOffset = 0
	InitialMarginLeft = 1
	InitialMarginRight = 132 // Set to tt_width for terminals
	InitialMarginTop = 0
	InitialMarginBottom = 0
	InitialOptions = 0

	// Set up sets for prefixes
	// NOTE - this matches prefix commands
	for cmd := CmdPrefixAst; cmd <= CmdPrefixTilde; cmd++ {
		Prefixes.SetBit(&Prefixes, int(cmd), 1)
	}

	// Initialize default prompts
	DfltPrompts[NoPrompt] = "        "
	DfltPrompts[CharPrompt] = "Charset:"
	DfltPrompts[GetPrompt] = "Get    :"
	DfltPrompts[EqualPrompt] = "Equal  :"
	DfltPrompts[KeyPrompt] = "Key    :"
	DfltPrompts[CmdPrompt] = "Command:"
	DfltPrompts[SpanPrompt] = "Span   :"
	DfltPrompts[TextPrompt] = "Text   :"
	DfltPrompts[FramePrompt] = "Frame  :"
	DfltPrompts[FilePrompt] = "File   :"
	DfltPrompts[ColumnPrompt] = "Column :"
	DfltPrompts[MarkPrompt] = "Mark   :"
	DfltPrompts[ParamPrompt] = "Param  :"
	DfltPrompts[TopicPrompt] = "Topic  :"
	DfltPrompts[ReplacePrompt] = "Replace:"
	DfltPrompts[ByPrompt] = "By     :"
	DfltPrompts[VerifyPrompt] = "Verify ?"
	DfltPrompts[PatternPrompt] = "Pattern:"
	DfltPrompts[PatternSetPrompt] = "Pat Set:"

	FileData.OldCmds = true
	FileData.Entab = false
	FileData.Space = 500000
	FileData.Purge = false
	FileData.Versions = 1
	FileData.Initial = ""
}

// initCmd is a helper function to initialize command attributes
func initCmd(
	cmd Commands,
	lps []LeadParam,
	eqa EqualAction,
	tpc int,
	pnm1 PromptType,
	tr1 bool,
	mla1 bool,
	pnm2 PromptType,
	tr2 bool,
	mla2 bool,
) {
	p := &CmdAttrib[cmd]
	p.LpAllowed = 0
	for _, lp := range lps {
		p.LpAllowed |= (1 << uint(lp))
	}
	p.EqAction = eqa
	p.TpCount = tpc

	if tpc >= 1 {
		p.TparInfo[1].PromptName = pnm1
		p.TparInfo[1].TrimReply = tr1
		p.TparInfo[1].MlAllowed = mla1
	}
	if tpc >= 2 {
		p.TparInfo[2].PromptName = pnm2
		p.TparInfo[2].TrimReply = tr2
		p.TparInfo[2].MlAllowed = mla2
	}
}

// allLeadParams returns a slice with all lead parameter types
func allLeadParams() []LeadParam {
	return []LeadParam{
		LeadParamNone,
		LeadParamPlus,
		LeadParamMinus,
		LeadParamPInt,
		LeadParamNInt,
		LeadParamPIndef,
		LeadParamNIndef,
		LeadParamMarker,
	}
}

// initializeCommandTablePart1 initializes the first part of the command table
func initializeCommandTablePart1() {
	initCmd(CmdNoop, allLeadParams(), EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdUp, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdDown, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdRight, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdLeft, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdHome, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdReturn, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdTab, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdBacktab, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdRubout, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdJump, allLeadParams(), EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdAdvance, allLeadParams(), EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdPositionColumn, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdPositionLine, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdOpSysCommand, []LeadParam{LeadParamNone}, EqNil, 1, CmdPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowForward, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowBackward, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowRight, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowLeft, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowScroll, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt, LeadParamPIndef, LeadParamNIndef}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowTop, []LeadParam{LeadParamNone}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowEnd, []LeadParam{LeadParamNone}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowNew, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowMiddle, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowSetHeight, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWindowUpdate, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdGet, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt}, EqNil, 1, GetPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdNext, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt}, EqNil, 1, CharPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdBridge, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 1, CharPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdReplace, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt, LeadParamPIndef, LeadParamNIndef}, EqNil, 2, ReplacePrompt, false, false, ByPrompt, false, true)
	initCmd(CmdEqualString, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPIndef, LeadParamNIndef}, EqNil, 1, EqualPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdEqualColumn, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPIndef, LeadParamNIndef}, EqNil, 1, ColumnPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdEqualMark, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPIndef, LeadParamNIndef}, EqNil, 1, MarkPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdEqualEol, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPIndef, LeadParamNIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdEqualEop, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdEqualEof, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdOvertypeMode, []LeadParam{LeadParamNone}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdInsertMode, []LeadParam{LeadParamNone}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdOvertypeText, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqOld, 1, TextPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdInsertText, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqOld, 1, TextPrompt, false, true, NoPrompt, false, false)
	initCmd(CmdTypeText, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqOld, 1, TextPrompt, false, true, NoPrompt, false, false)
	initCmd(CmdInsertLine, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdInsertChar, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdInsertInvisible, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdDeleteLine, allLeadParams(), EqDel, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdDeleteChar, allLeadParams(), EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
}

// initializeCommandTablePart2 initializes the second part of the command table
func initializeCommandTablePart2() {
	initCmd(CmdSwapLine, allLeadParams(), EqDel, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdSplitLine, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdDittoUp, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt, LeadParamPIndef, LeadParamNIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdDittoDown, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt, LeadParamPIndef, LeadParamNIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdCaseUp, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt, LeadParamPIndef, LeadParamNIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdCaseLow, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt, LeadParamPIndef, LeadParamNIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdCaseEdit, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt, LeadParamPIndef, LeadParamNIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdSetMarginLeft, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdSetMarginRight, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdLineFill, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdLineJustify, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdLineSquash, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdLineCentre, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdLineLeft, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdLineRight, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWordAdvance, allLeadParams(), EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdWordDelete, allLeadParams(), EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdAdvanceParagraph, allLeadParams(), EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdDeleteParagraph, allLeadParams(), EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdSpanDefine, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamMarker}, EqNil, 1, SpanPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdSpanTransfer, []LeadParam{LeadParamNone}, EqNil, 1, SpanPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdSpanCopy, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqNil, 1, SpanPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdSpanCompile, []LeadParam{LeadParamNone}, EqNil, 1, SpanPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdSpanJump, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 1, SpanPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdSpanIndex, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdSpanAssign, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt, LeadParamPIndef}, EqNil, 2, SpanPrompt, true, false, TextPrompt, false, true)
	initCmd(CmdBlockDefine, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamMarker}, EqNil, 1, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdBlockTransfer, []LeadParam{LeadParamNone}, EqNil, 1, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdBlockCopy, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqNil, 1, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFrameKill, []LeadParam{LeadParamNone}, EqNil, 1, FramePrompt, true, false, NoPrompt, false, false)
	initCmd(CmdFrameEdit, []LeadParam{LeadParamNone}, EqNil, 1, FramePrompt, true, false, NoPrompt, false, false)
	initCmd(CmdFrameReturn, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdSpanExecute, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 1, SpanPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdSpanExecuteNoRecompile, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 1, SpanPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdFrameParameters, []LeadParam{LeadParamNone}, EqNil, 1, ParamPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdFileInput, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 1, FilePrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileOutput, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 1, FilePrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileEdit, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 1, FilePrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileRead, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileWrite, allLeadParams(), EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileClose, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileRewind, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileKill, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileExecute, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 1, FilePrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileSave, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileTable, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileGlobalInput, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 1, FilePrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileGlobalOutput, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 1, FilePrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileGlobalRewind, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdFileGlobalKill, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdUserCommandIntroducer, []LeadParam{LeadParamNone}, EqOld, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdUserKey, []LeadParam{LeadParamNone}, EqNil, 2, KeyPrompt, true, false, CmdPrompt, false, true)
	initCmd(CmdUserParent, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdUserSubprocess, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdUserUndo, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdHelp, []LeadParam{LeadParamNone}, EqNil, 1, TopicPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdVerify, []LeadParam{LeadParamNone}, EqNil, 1, VerifyPrompt, true, false, NoPrompt, false, false)
	initCmd(CmdCommand, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdMark, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamMinus, LeadParamPInt, LeadParamNInt}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdPage, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdQuit, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdDump, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdValidate, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdExecuteString, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 1, CmdPrompt, false, true, NoPrompt, false, false)
	initCmd(CmdDoLastCommand, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdExtended, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdExitAbort, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdExitFail, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdExitSuccess, []LeadParam{LeadParamNone, LeadParamPlus, LeadParamPInt, LeadParamPIndef}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdPatternDummyPattern, []LeadParam{}, EqNil, 1, PatternPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdPatternDummyText, []LeadParam{}, EqNil, 1, TextPrompt, false, false, NoPrompt, false, false)
	initCmd(CmdResizeWindow, []LeadParam{LeadParamNone}, EqNil, 0, NoPrompt, false, false, NoPrompt, false, false)
}

// ValueInitializations performs all value initializations
func ValueInitializations() {
	setupInitialValues()
	initializeCommandTablePart1()
	initializeCommandTablePart2()
}
