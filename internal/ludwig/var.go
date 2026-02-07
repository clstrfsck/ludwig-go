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
var BlankString StrObject = *NewFilled(' ')
var InitialVerify VerifyArray
var DefaultTabStops TabArray

// Sets of characters
var PrintableSet big.Int
var SpaceSet big.Int
var AlphaSet big.Int
var LowerSet big.Int
var UpperSet big.Int
var NumericSet big.Int
var PunctuationSet big.Int

// Output file actions
var FileData FileDataType

// Info about the terminal
var TerminalInfo TerminalInfoType

// Word definition sets
var WordElements [MaxWordSets](big.Int)

func init() {
	// Initialize DefaultTabStops with the pattern: false, true, false, false...
	for i := range DefaultTabStops {
		DefaultTabStops[i] = (i % 8) == 1
	}

	// Initialize character sets for pattern matching
	// SpaceSet: ' '
	SpaceSet.SetBit(&SpaceSet, ' ', 1)

	// LowerSet: 'a'..'z'
	for i := byte('a'); i <= byte('z'); i++ {
		LowerSet.SetBit(&LowerSet, int(i), 1)
	}

	// UpperSet: 'A'..'Z'
	for i := byte('A'); i <= byte('Z'); i++ {
		UpperSet.SetBit(&UpperSet, int(i), 1)
	}

	// AlphaSet: union of LowerSet and UpperSet
	AlphaSet.Or(&AlphaSet, &LowerSet)
	AlphaSet.Or(&AlphaSet, &UpperSet)

	// NumericSet: '0'..'9'
	for i := byte('0'); i <= byte('9'); i++ {
		NumericSet.SetBit(&NumericSet, int(i), 1)
	}

	// PrintableSet: 32..126 (space to tilde)
	for i := 32; i <= 126; i++ {
		PrintableSet.SetBit(&PrintableSet, i, 1)
	}

	// PunctuationSet: '!','"','''','(',')',',','.',':',';','?','`'
	punctChars := []byte{33, 34, 39, 40, 41, 44, 46, 58, 59, 63, 96}
	for _, ch := range punctChars {
		PunctuationSet.SetBit(&PunctuationSet, int(ch), 1)
	}
}
