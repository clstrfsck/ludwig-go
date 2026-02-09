/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Global variables used throughout the Ludwig code base.

package ludwig

import "math/big"

// ModeType represents the editor mode
type ModeType int

const (
	ModeOvertype ModeType = iota
	ModeInsert
	ModeCommand
)

// LudwigModeType represents the Ludwig execution mode
type LudwigModeType int

const (
	LudwigBatch LudwigModeType = iota
	LudwigHardcopy
	LudwigScreen
)

// LookupExpType represents a lookup expansion entry
type LookupExpType struct {
	Extn    byte
	Command Commands
}

// Global variables

// Version and configuration
var LudwigVersion string = "X5.0-006"
var ProgramDirectory string
var TtControlC bool
var TtWinChanged bool

// Keyboard interface
var NrKeyNames int
var KeyNameList []KeyNameRecord
var KeyIntroducers [MaxSetRange + 1]bool

// Special frames
var CurrentFrame *FrameObject
var FrameOops *FrameObject
var FrameCmd *FrameObject
var FrameHeap *FrameObject

// Global variables
var LudwigAborted bool
var ExitAbort bool
var VduFreeFlag bool
var Hangup bool

var EditMode ModeType
var PreviousMode ModeType

var Files [MaxFiles + 1]*FileObject
var FilesFrames [MaxFiles + 1]*FrameObject

var FgiFile int
var FgoFile int

var FirstSpan *SpanObject

var LudwigMode LudwigModeType
var CommandIntroducer int

var PromptRegion [MaxTpCount + 1]PromptRegionAttrib

var ScrFrame *FrameObject
var ScrTopLine *LineHdrObject
var ScrBotLine *LineHdrObject
var ScrMsgRow int
var ScrNeedsFix bool

// Compiler variables
var CompilerCode [MaxCode + 1]CodeObject
var CodeList *CodeHeader
var CodeTop int

// Variables used in interpreting a command
var Prefixes big.Int
var Lookup [OrdMaxChar + MaxSpecialKeys + 1]CommandObject
var LookupExp [ExpandLim + 1]LookupExpType
var LookupExpPtr [CmdNoSuch + 1]int
var CmdAttrib [CmdNoSuch + 1]CmdAttribRec
var DfltPrompts [PatternSetPrompt + 1]string
var ExecLevel int

// Initial frame settings
var InitialMarks MarkArray
var InitialScrHeight int
var InitialScrWidth int
var InitialScrOffset int
var InitialMarginLeft int
var InitialMarginRight int
var InitialMarginTop int
var InitialMarginBottom int
var InitialTabStops TabArray
var InitialOptions FrameOptions

// Useful constants
var BlankString *StrObject = NewFilled(' ', MaxStrLen)
var InitialVerify VerifyArray
var DefaultTabStops TabArray

// Output file actions
var FileData FileDataType

// Info about the terminal
var TerminalInfo TerminalInfoType

func init() {
	// Initialize DefaultTabStops with the pattern: false, true, false, false...
	for i := range DefaultTabStops {
		DefaultTabStops[i] = (i % 8) == 1
	}
}
