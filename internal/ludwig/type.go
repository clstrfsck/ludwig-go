/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Type declarations used throughout the Ludwig code base.

package ludwig

import (
	"math/big"
)

// VerifyResponse represents user's response to a verify prompt
type VerifyResponse int

const (
	VerifyReplyYes VerifyResponse = iota
	VerifyReplyNo
	VerifyReplyAlways
	VerifyReplyQuit
)

// ParseType represents different parsing contexts
type ParseType int

const (
	ParseCommand ParseType = iota
	ParseInput
	ParseOutput
	ParseEdit
	ParseStdin
	ParseExecute
)

// FrameOptionsElts represents frame option flags
type FrameOptionsElts int

const (
	OptAutoIndent FrameOptionsElts = iota
	OptAutoWrap
	OptNewLine
	OptSpecialFrame // OOPS,COMMAND,HEAP
)

// FrameOptions is a set of frame options (bitset)
type FrameOptions uint32

// Commands represents all available Ludwig commands
type Commands int

const (
	CmdNoop Commands = iota

	// Cursor movement
	CmdUp
	CmdDown
	CmdLeft
	CmdRight
	CmdHome
	CmdReturn
	CmdTab
	CmdBacktab

	CmdRubout
	CmdJump
	CmdAdvance
	CmdPositionColumn
	CmdPositionLine
	CmdOpSysCommand

	// Window control
	CmdWindowForward
	CmdWindowBackward
	CmdWindowLeft
	CmdWindowRight
	CmdWindowScroll
	CmdWindowTop
	CmdWindowEnd
	CmdWindowNew
	CmdWindowMiddle
	CmdWindowSetHeight
	CmdWindowUpdate

	// Search and comparison
	CmdGet
	CmdNext
	CmdBridge
	CmdReplace
	CmdEqualString
	CmdEqualColumn
	CmdEqualMark
	CmdEqualEol
	CmdEqualEop
	CmdEqualEof

	CmdOvertypeMode
	CmdInsertMode

	// Text insertion/deletion
	CmdOvertypeText
	CmdInsertText
	CmdTypeText
	CmdInsertLine
	CmdInsertChar
	CmdInsertInvisible
	CmdDeleteLine
	CmdDeleteChar

	// Text manipulation
	CmdSwapLine
	CmdSplitLine
	CmdDittoUp
	CmdDittoDown
	CmdCaseUp
	CmdCaseLow
	CmdCaseEdit
	CmdSetMarginLeft
	CmdSetMarginRight

	// Word processing
	CmdLineFill
	CmdLineJustify
	CmdLineSquash
	CmdLineCentre
	CmdLineLeft
	CmdLineRight
	CmdWordAdvance
	CmdWordDelete
	CmdAdvanceParagraph
	CmdDeleteParagraph

	// Span commands
	CmdSpanDefine
	CmdSpanTransfer
	CmdSpanCopy
	CmdSpanCompile
	CmdSpanJump
	CmdSpanIndex
	CmdSpanAssign

	// Block commands
	CmdBlockDefine
	CmdBlockTransfer
	CmdBlockCopy

	// Frame commands
	CmdFrameKill
	CmdFrameEdit
	CmdFrameReturn
	CmdSpanExecute
	CmdSpanExecuteNoRecompile
	CmdFrameParameters

	// File commands
	CmdFileInput
	CmdFileOutput
	CmdFileEdit
	CmdFileRead
	CmdFileWrite
	CmdFileClose
	CmdFileRewind
	CmdFileKill
	CmdFileExecute
	CmdFileSave
	CmdFileTable
	CmdFileGlobalInput
	CmdFileGlobalOutput
	CmdFileGlobalRewind
	CmdFileGlobalKill

	CmdUserCommandIntroducer
	CmdUserKey
	CmdUserParent
	CmdUserSubprocess
	CmdUserUndo
	CmdUserLearn
	CmdUserRecall

	CmdResizeWindow

	// Miscellaneous
	CmdHelp
	CmdVerify
	CmdCommand
	CmdMark
	CmdPage
	CmdQuit
	CmdDump
	CmdValidate
	CmdExecuteString
	CmdDoLastCommand

	CmdExtended

	CmdExitAbort
	CmdExitFail
	CmdExitSuccess

	// End of user commands

	// Dummy commands for pattern matcher
	CmdPatternDummyPattern
	CmdPatternDummyText

	// Compiler commands
	CmdPcJump
	CmdExitTo
	CmdFailTo
	CmdIterate

	// Prefix commands
	CmdPrefixAst
	CmdPrefixA
	CmdPrefixB
	CmdPrefixC
	CmdPrefixD
	CmdPrefixE
	CmdPrefixEo
	CmdPrefixEq
	CmdPrefixF
	CmdPrefixFg
	CmdPrefixI
	CmdPrefixK
	CmdPrefixL
	CmdPrefixO
	CmdPrefixP
	CmdPrefixS
	CmdPrefixT
	CmdPrefixTc
	CmdPrefixTf
	CmdPrefixU
	CmdPrefixW
	CmdPrefixX
	CmdPrefixY
	CmdPrefixZ
	CmdPrefixTilde

	CmdNoSuch
)

// LeadParam represents leading parameter types
type LeadParam int

const (
	LeadParamNone LeadParam = iota
	LeadParamPlus
	LeadParamMinus
	LeadParamPInt
	LeadParamNInt
	LeadParamPIndef
	LeadParamNIndef
	LeadParamMarker
)

// EqualAction is a command attribute type used to control the behaviour of mark Equals
type EqualAction int

const (
	EqNil EqualAction = iota // leave mark alone
	EqDel                    // delete mark e.g. delete and kill
	EqOld                    // set mark to cursor posn. before cmd
)

// PromptType represents different prompt types
type PromptType int

const (
	NoPrompt PromptType = iota
	CharPrompt
	GetPrompt
	EqualPrompt
	KeyPrompt
	CmdPrompt
	SpanPrompt
	TextPrompt
	FramePrompt
	FilePrompt
	ColumnPrompt
	MarkPrompt
	ParamPrompt
	TopicPrompt
	ReplacePrompt
	ByPrompt
	VerifyPrompt
	PatternPrompt
	PatternSetPrompt
)

// ParameterType represents pattern parameter types
type ParameterType int

const (
	PatternFail ParameterType = iota
	PatternRange
	NullParam
)

// Data structures

// TParObject represents trailing parameter for command
type TParObject struct {
	Len int // strlen_range
	Dlm byte
	Str *StrObject
	Nxt *TParObject
	Con *TParObject
}

// CodeHeader represents a code header structure
type CodeHeader struct {
	FLink *CodeHeader
	BLink *CodeHeader
	Ref   int
	Code  int
	Len   int
}

// FileObject represents a file object
type FileObject struct {
	// Fields for "FILE.PAS" only
	Valid     bool
	FirstLine *LineHdrObject
	LastLine  *LineHdrObject
	LineCount int

	// Fields set by "FILESYS", read by "FILE"
	OutputFlag bool
	Eof        bool
	Filename   string
	LCounter   int

	// Fields for "FILESYS" only
	Memory         string
	Tnm            string
	Entab          bool
	Create         bool
	Fd             int
	Mode           int
	Idx            int
	Len            int
	Buf            []byte
	PreviousFileId int64

	// Fields for controlling version backup
	Purge    bool
	Versions int
}

// MarkObject represents a mark in the editor
type MarkObject struct {
	Line *LineHdrObject
	Col  int
}

// MarkArray represents an array of mark pointers
type MarkArray [MaxMarkNumber - MinMarkNumber + 1]*MarkObject

// TabArray represents an array of tab stops
type TabArray [MaxStrLenP + 1]bool

// VerifyArray represents an array of verify flags
type VerifyArray [MaxVerify + 1]bool

// FrameObject represents a frame in the editor
type FrameObject struct {
	FirstGroup    *GroupObject
	LastGroup     *GroupObject
	Dot           *MarkObject
	Marks         MarkArray
	ScrHeight     int
	ScrWidth      int
	ScrOffset     int
	ScrDotLine    int
	Span          *SpanObject
	ReturnFrame   *FrameObject
	InputCount    uint32
	SpaceLimit    int
	SpaceLeft     int
	TextModified  bool
	MarginLeft    int
	MarginRight   int
	MarginTop     int
	MarginBottom  int
	TabStops      TabArray
	Options       FrameOptions
	InputFile     int
	OutputFile    int
	GetTpar       TParObject
	GetPatternPtr *DFATableObject
	EqsTpar       TParObject
	EqsPatternPtr *DFATableObject
	Rep1Tpar      TParObject
	RepPatternPtr *DFATableObject
	Rep2Tpar      TParObject
	VerifyTpar    TParObject
}

// GroupObject represents a group of lines
type GroupObject struct {
	FLink       *GroupObject
	BLink       *GroupObject
	Frame       *FrameObject
	FirstLine   *LineHdrObject
	LastLine    *LineHdrObject
	FirstLineNr int
	NrLines     int
}

// LineHdrObject represents a line header
type LineHdrObject struct {
	FLink    *LineHdrObject
	BLink    *LineHdrObject
	Group    *GroupObject
	OffsetNr int
	Marks    []*MarkObject
	Str      *StrObject
	Used     int
	ScrRowNr int
}

func (l *LineHdrObject) Len() int {
	if l.Str == nil {
		return 0
	}
	return l.Str.Len()
}

// SpanObject represents a span
type SpanObject struct {
	FLink   *SpanObject
	BLink   *SpanObject
	Frame   *FrameObject
	MarkOne *MarkObject
	MarkTwo *MarkObject
	Name    string
	Code    *CodeHeader
}

// PromptRegionAttrib represents prompt region attributes
type PromptRegionAttrib struct {
	LineNr int
	Redraw *LineHdrObject
}

// TransitionObject represents a DFA transition
type TransitionObject struct {
	TransitionAcceptSet big.Int // bitset for accept set
	AcceptNextState     int
	NextTransition      *TransitionObject
	StartFlag           bool
}

// NFAAttributeType represents NFA attributes
type NFAAttributeType struct {
	GeneratorSet [MaxNFAStateRange + 1]bool // bitset
	EquivList    *StateEltObject
	EquivSet     [MaxNFAStateRange + 1]bool // bitset
}

// StateEltObject represents a state element
type StateEltObject struct {
	StateElt int
	NextElt  *StateEltObject
}

// DFAStateType represents a DFA state
type DFAStateType struct {
	Transitions      *TransitionObject
	Marked           bool
	NFAAttributes    NFAAttributeType
	PatternStart     bool
	FinalAccept      bool
	LeftTransition   bool
	RightTransition  bool
	LeftContextCheck bool
}

// PatternDefType represents a pattern definition
type PatternDefType struct {
	Strng  StrObject
	Length int
}

// DFATableObject represents a DFA table
type DFATableObject struct {
	DFATable      [MaxDFAStateRange + 1]DFAStateType
	DFAStatesUsed int
	Definition    PatternDefType
}

// NFATransitionType represents NFA transition
type NFATransitionType struct {
	Indefinite bool
	Fail       bool
	EpsilonOut bool
	// Union-like structure - only one set is valid at a time
	FirstOut  int     // for epsilon transitions
	SecondOut int     // for epsilon transitions
	NextState int     // for non-epsilon transitions
	AcceptSet big.Int // bitset for non-epsilon transitions
}

// NFATableType represents an NFA table
type NFATableType [MaxNFAStateRange + 1]NFATransitionType

// FileDataType represents global file defaults
type FileDataType struct {
	OldCmds  bool
	Entab    bool
	Space    int
	Initial  string
	Purge    bool
	Versions int
}

// CodeObject represents a code instruction
type CodeObject struct {
	Rep  LeadParam
	Cnt  int
	Op   Commands
	Tpar *TParObject
	Code *CodeHeader
	Lbl  int
}

// CommandObject represents a command
type CommandObject struct {
	Command Commands
	Code    *CodeHeader
	Tpar    *TParObject
}

// TerminalInfoType represents terminal information
type TerminalInfoType struct {
	Name   string
	Width  int
	Height int
}

// KeyNameRecord represents a key name mapping
type KeyNameRecord struct {
	KeyName string
	KeyCode int
}

// TParAttribute represents trailing parameter attributes
type TParAttribute struct {
	PromptName PromptType
	TrimReply  bool
	MlAllowed  bool
}

// CmdAttribRec represents command attributes
type CmdAttribRec struct {
	LpAllowed uint32 // bitset of allowed LeadParam values
	EqAction  EqualAction
	TpCount   int
	TparInfo  [MaxTpCount + 1]TParAttribute
}

// HelpRecord represents a help record
type HelpRecord struct {
	Key string
	Txt string
}

// Helper functions for FrameOptions bitset operations
func (f FrameOptions) Has(opt FrameOptionsElts) bool {
	return f&(1<<uint(opt)) != 0
}

func (f *FrameOptions) Set(opt FrameOptionsElts) {
	*f |= (1 << uint(opt))
}

func (f *FrameOptions) Clear(opt FrameOptionsElts) {
	*f &^= (1 << uint(opt))
}

func (f *FrameOptions) Toggle(opt FrameOptionsElts) {
	*f ^= (1 << uint(opt))
}
