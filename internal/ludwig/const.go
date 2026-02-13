/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Constants used throughout the Ludwig codebase, including limits,
// messages, and other fixed values.

package ludwig

import (
	"math"
)

const (
	// MaxInt is the maximum integer value
	MaxInt = math.MaxInt32

	// OrdMaxChar is the max value of character
	OrdMaxChar = 255

	// MaxFiles is the maximum number of files
	MaxFiles = 100

	// MaxGroupLines is the maximum lines per group
	MaxGroupLines      = 64
	MaxGroupLineOffset = MaxGroupLines - 1

	// MaxLines is the maximum lines per frame
	MaxLines = MaxInt

	MaxMarkNumber     = 10
	MinUserMarkNumber = 1
	MaxUserMarkNumber = 9

	MarkEquals   = 0
	MarkModified = 10

	// MaxSpace is the max chars allowed per frame
	MaxSpace = 1000000

	// MaxRecSize is the max length rec. in inp file
	MaxRecSize = 512

	// MaxStrLen is the max length of a string
	MaxStrLen  = 400
	MaxStrLenP = MaxStrLen + 1

	// MaxScrRows is the max nr of rows on screen
	MaxScrRows = 100

	// MaxScrCols is the max nr of cols on screen
	MaxScrCols = 255

	// MaxCode is the length of code array
	MaxCode = 4000

	// MaxVerify is the max nr of V commands in span
	MaxVerify = 256

	MaxTparRecursion = 100
	MaxTpCount       = 2
	MaxExecRecursion = 100

	// MaxWordSets is the max nr of word element sets
	MaxWordSets   = 2
	MaxWordSetsM1 = MaxWordSets - 1

	// TPar interpretations
	TpdLit         = '\'' // don't do fancy processing on this one
	TpdSmart       = '`'  // search target is a pattern
	TpdExact       = '"'  // use exact case during search
	TpdSpan        = '$'  // span substitution
	TpdPrompt      = '&'  // get parameter from user terminal
	TpdEnvironment = '?'  // environment enquiry
	ExpandLim      = 130  // at least # of multi-letter commands + 1

	// String lengths
	NameLen     = 31   // Max length of a spn/frm name
	FileNameLen = 1024 // Default length of file name (FILENAME_MAX varies by OS)
	KeyLen      = 4    // Key length for HELP FILE

	// Keyboard interface
	MaxSpecialKeys = 1000 // Taken from original XWin def
	MaxNrKeyNames  = 1000
	MaxParseTable  = 300

	// Application exit codes
	NormalExit   = 0
	AbnormalExit = 1

	// Regular expression state machine
	MaxNFAStateRange = 200        // no of states in NFA
	MaxDFAStateRange = 255        // no of states in DFA
	MaxSetRange      = OrdMaxChar // no of elts in accept sets
	PatternNull      = 0          // acts as nil in array ptrs
	PatternNFAStart  = 1          // the first NFA state allocated
	PatternDFAKill   = 0          // the dfa killer state
	PatternDFAFail   = 0          // To keep old versions happy
	PatternDFAStart  = 2          // The DFA starting state
	PatternMaxDepth  = 20         // maximum recursion depth in parser

	// Symbols used in pattern specification
	PatternKStar       = '*' // Kleene star
	PatternComma       = ',' // The context Delimiter
	PatternRParen      = ')'
	PatternLParen      = '('
	PatternDefineSetU  = 'D' // These two used as a pair
	PatternDefineSetL  = 'd' // The define a char set symbol
	PatternMark        = '@' // match the numbered mark
	PatternEquals      = '=' // match the Equals mark
	PatternModified    = '%' // match the Modified mark
	PatternPlus        = '+' // Kleene Plus
	PatternNegate      = '-' // To negate a char set
	PatternBar         = '|' // Alternation ( OR )
	PatternLRangeDelim = '[' // For specification of a range of
	PatternRRangeDelim = ']' // repetitions
	PatternSpace       = ' '

	// set locations for line end specifiers
	PatternBegLine     = 0 // the <  (beginning of line) specifier
	PatternEndLine     = 1 // the >  (end of line) specifier
	PatternLeftMargin  = 3 // l.brace( left margin ) specifier
	PatternRightMargin = 4 // r.brace( right margin ) specifier
	PatternDotColumn   = 5 // the ^  ( dots column)   specifier

	PatternMarksStart    = 19 // mark 1 = 20, 2 = 21, etc
	PatternMarksModified = 29 // marks_start + mark_modified
	PatternMarksEquals   = 19 // marks_start + mark_equals

	PatternAlphaStart = 32 // IE. ASCII
)

// Frame names
const (
	BlankFrameName   = ""
	DefaultFrameName = "LUDWIG"
)

// Messages
const (
	MsgBlank                   = ""
	MsgAbort                   = "Aborted. Output Files may be CORRUPTED"
	MsgBadFormatInTabTable     = "Bad Format for list of Tab stops."
	MsgCantKillFrame           = "Can't Kill Frame."
	MsgCantSplitNullLine       = "Can't split the Null line."
	MsgCommandNotValid         = "No Command starts with this character."
	MsgCommandRecursionLimit   = "Command recursion limit exceeded."
	MsgCommentsIllegal         = "Immediate mode comments are not allowed."
	MsgCompilerCodeOverflow    = "Compiler code overflow, too many compiled spans."
	MsgCopyrightAndLoadingFile = "Copyright (C) 1981, 1987,  University of Adelaide."
	MsgCountTooLarge           = "Count too large."
	MsgDecommitted             = "Warning - Decommitted feature."
	MsgEmptySpan               = "Span is empty."
	MsgEqualsNotSet            = "The Equals mark is not defined."
	MsgErrorOpeningKeysFile    = "Error opening keys definitions file."
	MsgExecutingInitFile       = "Executing initialization file."
	MsgFileAlreadyInUse        = "File already in use."
	MsgFileAlreadyOpen         = "File already open."
	MsgFrameHasFilesAttached   = "Frame Has Files Attached."
	MsgFrameOfThatNameExists   = "Frame of that Name Already Exists."
	MsgIllegalLeadingParam     = "Illegal leading parameter."
	MsgIllegalMarkNumber       = "Illegal mark number."
	MsgIllegalParamDelimiter   = "Illegal parameter delimiter."
	MsgIncompat                = "Incompatible switches specified."
	MsgInteractiveModeOnly     = "Allowed in interactive mode only."
	MsgInvalidCmdIntroducer    = "Invalid command introducer."
	MsgInvalidInteger          = "Trailing parameter integer is invalid."
	MsgInvalidKeysFile         = "Invalid keys definition file."
	MsgInvalidScreenHeight     = "Invalid height for screen."
	MsgInvalidSlotNumber       = "Invalid file slot number."
	MsgInvalidTOption          = "Invalid Tab Option."
	MsgInvalidRuler            = "Invalid Ruler."
	MsgInvalidParameterCode    = "Invalid Parameter Code."
	MsgLeftMarginGeRight       = "Specified Left Margin is not less than Right Margin."
	MsgLongInputLine           = "Long input line has been split."
	MsgMarginOutOfRange        = "Margin out of Range."
	MsgMarginSyntaxError       = "Margin Syntax Error."
	MsgMarkNotDefined          = "Mark Not Defined."
	MsgMissingTrailingDelim    = "Missing trailing delimiter."
	MsgNoDefaultStr            = "No default for trailing parameter string."
	MsgNoFileOpen              = "No file open."
	MsgNoMoreFilesAllowed      = "No more files are allowed."
	MsgNoRoomOnLine            = "Operation would cause a line to become too long."
	MsgNoSuchFrame             = "No such frame."
	MsgNoSuchSpan              = "No such span."
	MsgNonprintableIntroducer  = "Command Introducer is not printable"
	MsgNotEnoughInputLeft      = "Not enough input left to satisfy request."
	MsgNotImplemented          = "Not implemented."
	MsgNotInputFile            = "File is not an input file."
	MsgNotOutputFile           = "File is not an output file."
	MsgNotWhileEditingCmd      = "Operation not allowed while editing frame COMMAND."
	MsgNotAllowedInInsertMode  = "Command not allowed in insert mode."
	MsgOptionsSyntaxError      = "Syntax error in options."
	MsgOutOfRangeTabValue      = "Invalid value for tab stop."
	MsgParameterTooLong        = "Parameter is too long."
	MsgPromptsAreOneLine       = "A prompt string must be on one line."
	MsgScreenModeOnly          = "Command allowed in screen mode only."
	MsgScreenWidthInvalid      = "Invalid screen width specified."
	MsgSpanMustBeOneLine       = "A span used as a trailing parameter for this command must be one line."
	MsgSpanNamesAreOneLine     = "A span name must be on one line."
	MsgSpanOfThatNameExists    = "Span of that name already exists."
	MsgEnquiryMustBeOneLine    = "An enquiry item must be on one line."
	MsgUnknownItem             = "Unknown enquiry item."
	MsgSyntaxError             = "Syntax error."
	MsgSyntaxErrorInOptions    = "Syntax error in options."
	MsgSyntaxErrorInParamCmd   = "Syntax error in parameter command."
	MsgTopMarginLssBottom      = "Top margin must be less than or equal to bottom margin."
	MsgTparTooDeep             = "Trailing parameter translation has gone too deep."
	MsgUnknownOption           = "Not a valid option."
	MsgPatNoMatchingDelim      = "Pattern - No matching delimiter in pattern."
	MsgPatIllegalParameter     = "Pattern - Illegal parameter in pattern."
	MsgPatIllegalMarkNumber    = "Pattern - Illegal mark number in pattern."
	MsgPatPrematurePatternEnd  = "Pattern - Premature pattern end."
	MsgPatSetNotDefined        = "Pattern - Set not defined."
	MsgPatIllegalSymbol        = "Pattern - Illegal symbol in pattern."
	MsgPatSyntaxError          = "Pattern - Syntax error in pattern."
	MsgPatNullPattern          = "Pattern - Null pattern."
	MsgPatPatternTooComplex    = "Pattern - Pattern too complex."
	MsgPatErrorInSpan          = "Pattern - Error in dereferenced span."
	MsgPatErrorInRange         = "Pattern - Error in range specification."
	MsgReservedTpd             = "Delimiter reserved for future use."
	MsgIntegerNotInRange       = "Integer not in range"
	MsgModeError               = "Illegal Mode specification -- must be O,C or I"
	MsgWritingFile             = "Writing File."
	MsgLoadingFile             = "Loading File."
	MsgSavingFile              = "Saving File."
	MsgPaging                  = "Paging."
	MsgSearching               = "Searching."
	MsgQuitting                = "Quitting."
	MsgNoOutput                = "This Frame has no Output File attached."
	MsgNotModified             = "This Frame has not been modified."
	MsgNotRenamed              = "Output Files have '-lw*' appended to filename"
	MsgCantInvoke              = "Character cannot be invoked by a key"
	MsgExceededDynamicMemory   = "Exceeded dynamic memory limit."
	MsgInconsistentQualifier   = "Use of this qualifier is inconsistent with file operation"
	MsgUnrecognizedKeyName     = "Unrecognized key name"
	MsgKeyNameTruncated        = "Key name too long, name truncated"
	DbgInternalLogicError      = "Internal logic error."
	DbgBadFile                 = "FILE and FILESYS definition of file_object disagree."
	DbgCantMarkScrBotLine      = "Can't mark scr bot line."
	DbgCodePtrIsNil            = "Code ptr is nil."
	DbgFailedToUnloadAllScr    = "Failed to unload all scr."
	DbgFatalErrorSet           = "Fatal error set."
	DbgFirstFollowsLast        = "First follows last."
	DbgFirstNotAtTop           = "First Not at Top."
	DbgFlinkOrBlinkNotNil      = "Flink or blink not nil."
	DbgFrameCreationFailed     = "Frame creation failed."
	DbgFramePtrIsNil           = "Frame ptr is nil."
	DbgGroupHasLines           = "Group has lines."
	DbgGroupPtrIsNil           = "Group ptr is nil."
	DbgIllegalInstruction      = "Illegal Instruction."
	DbgInvalidBlink            = "Incorrect blink."
	DbgInvalidColumnNumber     = "Invalid column number."
	DbgInvalidFlags            = "Invalid flags."
	DbgInvalidFramePtr         = "Invalid frame ptr."
	DbgInvalidGroupPtr         = "Invalid group ptr."
	DbgInvalidLineLength       = "Invalid line length."
	DbgInvalidLineNr           = "Invalid line nr."
	DbgInvalidLinePtr          = "Invalid line ptr."
	DbgInvalidLineUsedLength   = "Invalid line used length."
	DbgInvalidNrLines          = "Invalid nr lines."
	DbgInvalidOffsetNr         = "Invalid offset nr."
	DbgInvalidScrParam         = "Invalid SCR Parameter."
	DbgInvalidScrRowNr         = "Invalid scr row nr."
	DbgInvalidSpanPtr          = "Invalid span ptr."
	DbgLastNotAtEnd            = "Last not at end."
	DbgLibraryRoutineFailure   = "Library routine call failed."
	DbgLineFromNumberFailed    = "Line from number failed."
	DbgLineHasMarks            = "Line has marks."
	DbgLineIsEop               = "Line is eop."
	DbgLineNotInScrFrame       = "Line not in scr frame."
	DbgLineOnScreen            = "Line on screen."
	DbgLinePtrIsNil            = "Line ptr is nil."
	DbgLineToNumberFailed      = "Line to number failed."
	DbgLinesFromDiffFrames     = "Lines from diff frames."
	DbgMarkInWrongFrame        = "Mark in wrong frame."
	DbgMarkMoveFailed          = "Mark move failed."
	DbgMarkPtrIsNil            = "Mark ptr is nil."
	DbgMarksFromDiffFrames     = "Marks from diff frames."
	DbgNeededFrameNotFound     = "Frame C or OOPS is not in the span list"
	DbgNotImmedCmd             = "Not immed cmd."
	DbgNxtNotNil               = "Nxt should be nil here."
	DbgPcOutOfRange            = "PC is out of Range."
	DbgRefCountIsZero          = "Reference count is zero."
	DbgRepeatNegative          = "Repeat Negative."
	DbgSpanNotDestroyed        = "Span not destroyed."
	DbgTopLineNotDrawn         = "Top line not drawn."
	DbgTparNil                 = "Tpar should not be nil."
	DbgWrongRowNr              = "Wrong row nr."
)

// GetFileNameLen returns the appropriate file name length for the current OS
func GetFileNameLen() int {
	// In Go, we could use a more dynamic approach
	// For Unix-like systems, PATH_MAX is typically 4096
	// For simplicity, we'll use a reasonable default
	return FileNameLen
}

// ExitSuccess returns the appropriate exit code for success
func ExitSuccess() int {
	return 0
}

// ExitFailure returns the appropriate exit code for failure
func ExitFailure() int {
	return 1
}
