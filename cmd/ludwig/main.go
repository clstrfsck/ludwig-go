/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         LUDWIG
//
// Description:  LUDWIG startup and shutdown.   Organize the timing of
//               the session, and the other general details.

package main

import (
	"os"
	"strings"

	. "ludwig-go/internal/ludwig"
)

func progWindup(setHangup bool) {
	Hangup = setHangup

	// DISABLE EVERYTHING TO DO WITH VDU'S AND INTERACTIVE USERS.
	// THIS IS BECAUSE VDU_FREE MUST HAVE BEEN INVOKED BEFORE
	// THIS EXIT HANDLER WAS, HENCE THE VDU IS NO LONGER AVAIL.

	LudwigMode = LudwigBatch
	ScrFrame = nil
	ScrTopLine = nil
	ScrBotLine = nil
	TtControlC = false
	ExitAbort = false

	// WIND OUT EVERYTHING FOR THE USER -- Gee that's nice of us!

	QuitCloseFiles()
}

func initialize() {
	InitialTabStops = DefaultTabStops

	// Now create the Code Header for the compiler to use
	CodeTop = 0
	CodeList = &CodeHeader{}
	// with code_list^ do
	CodeList.FLink = CodeList
	CodeList.BLink = CodeList
	CodeList.Ref = 1
	CodeList.Code = 1
	CodeList.Len = 0
}

func addLookupExp(index int, ch byte, cmd Commands) {
	LookupExp[index].Extn = ch
	LookupExp[index].Command = cmd
}

// lookupIndex converts a key code (possibly negative) to a Lookup array index
func lookupIndex(keyCode int) int {
	return keyCode // + MaxSpecialKeys
}

func loadCommandTable(oldVersion bool) {
	var keyCode int

	// for keyCode = -MaxSpecialKeys; keyCode <= -1; keyCode++ {
	// 	Lookup[lookupIndex(keyCode)].Command = CmdNoop
	// }
	if oldVersion {
		Lookup[lookupIndex(0)].Command = CmdNoop
		Lookup[lookupIndex(1)].Command = CmdNoop
		Lookup[lookupIndex(2)].Command = CmdWindowBackward
		Lookup[lookupIndex(3)].Command = CmdNoop
		Lookup[lookupIndex(4)].Command = CmdDeleteChar
		Lookup[lookupIndex(5)].Command = CmdWindowEnd
		Lookup[lookupIndex(6)].Command = CmdWindowForward
		Lookup[lookupIndex(7)].Command = CmdDoLastCommand
		Lookup[lookupIndex(8)].Command = CmdRubout
		Lookup[lookupIndex(9)].Command = CmdTab
		Lookup[lookupIndex(10)].Command = CmdDown
		Lookup[lookupIndex(11)].Command = CmdDeleteLine
		Lookup[lookupIndex(12)].Command = CmdInsertLine
		Lookup[lookupIndex(13)].Command = CmdReturn
		Lookup[lookupIndex(14)].Command = CmdWindowNew
		Lookup[lookupIndex(15)].Command = CmdNoop
		Lookup[lookupIndex(16)].Command = CmdUserCommandIntroducer
		Lookup[lookupIndex(17)].Command = CmdNoop
		Lookup[lookupIndex(18)].Command = CmdRight
		Lookup[lookupIndex(19)].Command = CmdNoop
		Lookup[lookupIndex(20)].Command = CmdWindowTop
		Lookup[lookupIndex(21)].Command = CmdUp
		Lookup[lookupIndex(22)].Command = CmdNoop
		Lookup[lookupIndex(23)].Command = CmdWordAdvance
		Lookup[lookupIndex(24)].Command = CmdNoop
		Lookup[lookupIndex(25)].Command = CmdNoop
		Lookup[lookupIndex(26)].Command = CmdUserParent
		Lookup[lookupIndex(27)].Command = CmdNoop
		Lookup[lookupIndex(28)].Command = CmdNoop
		Lookup[lookupIndex(29)].Command = CmdNoop
		Lookup[lookupIndex(30)].Command = CmdInsertChar
		Lookup[lookupIndex(31)].Command = CmdNoop
		Lookup[lookupIndex(' ')].Command = CmdNoop
		Lookup[lookupIndex('!')].Command = CmdNoop
		Lookup[lookupIndex('"')].Command = CmdDittoUp
		Lookup[lookupIndex('#')].Command = CmdNoop
		Lookup[lookupIndex('$')].Command = CmdNoop
		Lookup[lookupIndex('%')].Command = CmdNoop
		Lookup[lookupIndex('&')].Command = CmdNoop
		Lookup[lookupIndex('\'')].Command = CmdDittoDown
		Lookup[lookupIndex('(')].Command = CmdNoop
		Lookup[lookupIndex(')')].Command = CmdNoop
		Lookup[lookupIndex('*')].Command = CmdPrefixAst
		Lookup[lookupIndex('+')].Command = CmdNoop
		Lookup[lookupIndex(';')].Command = CmdNoop
		Lookup[lookupIndex('-')].Command = CmdNoop
		Lookup[lookupIndex('.')].Command = CmdNoop
		Lookup[lookupIndex('/')].Command = CmdNoop
		for keyCode = '0'; keyCode <= '9'; keyCode++ {
			Lookup[lookupIndex(keyCode)].Command = CmdNoop
		}
		Lookup[lookupIndex(':')].Command = CmdNoop
		Lookup[lookupIndex(';')].Command = CmdNoop
		Lookup[lookupIndex('<')].Command = CmdNoop
		Lookup[lookupIndex('=')].Command = CmdNoop
		Lookup[lookupIndex('>')].Command = CmdNoop
		Lookup[lookupIndex('?')].Command = CmdInsertInvisible
		Lookup[lookupIndex('@')].Command = CmdNoop
		Lookup[lookupIndex('A')].Command = CmdAdvance
		Lookup[lookupIndex('B')].Command = CmdPrefixB
		Lookup[lookupIndex('C')].Command = CmdInsertChar
		Lookup[lookupIndex('D')].Command = CmdDeleteChar
		Lookup[lookupIndex('E')].Command = CmdPrefixE
		Lookup[lookupIndex('F')].Command = CmdPrefixF
		Lookup[lookupIndex('G')].Command = CmdGet
		Lookup[lookupIndex('H')].Command = CmdHelp
		Lookup[lookupIndex('I')].Command = CmdInsertText
		Lookup[lookupIndex('J')].Command = CmdJump
		Lookup[lookupIndex('K')].Command = CmdDeleteLine
		Lookup[lookupIndex('L')].Command = CmdInsertLine
		Lookup[lookupIndex('M')].Command = CmdMark
		Lookup[lookupIndex('N')].Command = CmdNext
		Lookup[lookupIndex('O')].Command = CmdOvertypeText
		Lookup[lookupIndex('P')].Command = CmdNoop
		Lookup[lookupIndex('Q')].Command = CmdQuit
		Lookup[lookupIndex('R')].Command = CmdReplace
		Lookup[lookupIndex('S')].Command = CmdPrefixS
		Lookup[lookupIndex('T')].Command = CmdNoop
		Lookup[lookupIndex('U')].Command = CmdPrefixU
		Lookup[lookupIndex('V')].Command = CmdVerify
		Lookup[lookupIndex('W')].Command = CmdPrefixW
		Lookup[lookupIndex('X')].Command = CmdPrefixX
		Lookup[lookupIndex('Y')].Command = CmdPrefixY
		Lookup[lookupIndex('Z')].Command = CmdPrefixZ
		Lookup[lookupIndex('[')].Command = CmdNoop
		Lookup[lookupIndex('\\')].Command = CmdCommand
		Lookup[lookupIndex(']')].Command = CmdNoop
		Lookup[lookupIndex('^')].Command = CmdExecuteString
		Lookup[lookupIndex('_')].Command = CmdNoop
		Lookup[lookupIndex('`')].Command = CmdNoop
		for keyCode = 'a'; keyCode <= 'z'; keyCode++ {
			Lookup[lookupIndex(keyCode)].Command = CmdNoop
		}
		Lookup[lookupIndex('{')].Command = CmdSetMarginLeft
		Lookup[lookupIndex('|')].Command = CmdNoop
		Lookup[lookupIndex('}')].Command = CmdSetMarginRight
		Lookup[lookupIndex('~')].Command = CmdPrefixTilde
		Lookup[lookupIndex(127)].Command = CmdRubout
		for keyCode = 128; keyCode <= OrdMaxChar; keyCode++ {
			Lookup[lookupIndex(keyCode)].Command = CmdNoop
		}
		// for keyCode = -MaxSpecialKeys; keyCode <= OrdMaxChar; keyCode++ {
		// 	// with lookup[key_code] do
		// 	Lookup[lookupIndex(keyCode)].Code = nil
		// 	Lookup[lookupIndex(keyCode)].Tpar = nil
		// }

		// initialize lookupexp
		// case change command ; command =  * prefix }      {start at 1}
		addLookupExp(1, 'U', CmdCaseUp)
		addLookupExp(2, 'L', CmdCaseLow)
		addLookupExp(3, 'E', CmdCaseEdit)

		// A prefix }    {4}
		// There aren't any in this table! }

		// B prefix }    {4}
		addLookupExp(4, 'R', CmdBridge)

		// C prefix }    {5}
		// There aren't any in this table! }

		// D prefix }    {5}
		// There aren't any in this table! }

		// E prefix }    {5}
		addLookupExp(5, 'X', CmdSpanExecute)
		addLookupExp(6, 'D', CmdFrameEdit)
		addLookupExp(7, 'R', CmdFrameReturn)
		addLookupExp(8, 'N', CmdSpanExecuteNoRecompile)
		addLookupExp(9, 'Q', CmdPrefixEq)
		addLookupExp(10, 'O', CmdPrefixEo)
		addLookupExp(11, 'K', CmdFrameKill)
		addLookupExp(12, 'P', CmdFrameParameters)

		// EO prefix }   {13}
		addLookupExp(13, 'L', CmdEqualEol)
		addLookupExp(14, 'F', CmdEqualEof)
		addLookupExp(15, 'P', CmdEqualEop)

		// EQ prefix }   {16}
		addLookupExp(16, 'S', CmdEqualString)
		addLookupExp(17, 'C', CmdEqualColumn)
		addLookupExp(18, 'M', CmdEqualMark)

		// F prefix - files }    {19}
		addLookupExp(19, 'S', CmdFileSave)
		addLookupExp(20, 'B', CmdFileRewind)
		addLookupExp(21, 'I', CmdFileInput)
		addLookupExp(22, 'E', CmdFileEdit)
		addLookupExp(23, 'O', CmdFileOutput)
		addLookupExp(24, 'G', CmdPrefixFg)
		addLookupExp(25, 'K', CmdFileKill)
		addLookupExp(26, 'X', CmdFileExecute)
		addLookupExp(27, 'T', CmdFileTable)
		addLookupExp(28, 'P', CmdPage)

		// FG prefix - global files }    {29}
		addLookupExp(29, 'I', CmdFileGlobalInput)
		addLookupExp(30, 'O', CmdFileGlobalOutput)
		addLookupExp(31, 'B', CmdFileGlobalRewind)
		addLookupExp(32, 'K', CmdFileGlobalKill)
		addLookupExp(33, 'R', CmdFileRead)
		addLookupExp(34, 'W', CmdFileWrite)

		// I prefix }    {35}
		// There aren't any in this table! }

		// K prefix }    {35}
		// There aren't any in this table! }

		// L prefix }    {35}
		// There aren't any in this table! }

		// O prefix }    {35}
		// There aren't any in this table! }

		// P prefix }    {35}
		// There aren't any in this table! }

		// S prefix - mainly spans }     {35}
		addLookupExp(35, 'A', CmdSpanAssign)
		addLookupExp(36, 'C', CmdSpanCopy)
		addLookupExp(37, 'D', CmdSpanDefine)
		addLookupExp(38, 'T', CmdSpanTransfer)
		addLookupExp(39, 'W', CmdSwapLine)
		addLookupExp(40, 'L', CmdSplitLine)
		addLookupExp(41, 'J', CmdSpanJump)
		addLookupExp(42, 'I', CmdSpanIndex)
		addLookupExp(43, 'R', CmdSpanCompile)

		// T prefix }    {44}
		// There aren't any in this table! }

		// TC prefix }    {44}
		// There aren't any in this table! }

		// TF prefix }    {44}
		// There aren't any in this table! }

		// U prefix - user keyboard mappings }   {44}
		addLookupExp(44, 'C', CmdUserCommandIntroducer)
		addLookupExp(45, 'K', CmdUserKey)
		addLookupExp(46, 'P', CmdUserParent)
		addLookupExp(47, 'S', CmdUserSubprocess)

		// W prefix - window commands }  {48}
		addLookupExp(48, 'F', CmdWindowForward)
		addLookupExp(49, 'B', CmdWindowBackward)
		addLookupExp(50, 'M', CmdWindowMiddle)
		addLookupExp(51, 'T', CmdWindowTop)
		addLookupExp(52, 'E', CmdWindowEnd)
		addLookupExp(53, 'N', CmdWindowNew)
		addLookupExp(54, 'R', CmdWindowRight)
		addLookupExp(55, 'L', CmdWindowLeft)
		addLookupExp(56, 'H', CmdWindowSetHeight)
		addLookupExp(57, 'S', CmdWindowScroll)
		addLookupExp(58, 'U', CmdWindowUpdate)

		// X prefix - exit }             {59}
		addLookupExp(59, 'S', CmdExitSuccess)
		addLookupExp(60, 'F', CmdExitFail)
		addLookupExp(61, 'A', CmdExitAbort)

		// Y prefix - word processing }  {62}
		addLookupExp(62, 'F', CmdLineFill)
		addLookupExp(63, 'J', CmdLineJustify)
		addLookupExp(64, 'S', CmdLineSquash)
		addLookupExp(65, 'C', CmdLineCentre)
		addLookupExp(66, 'L', CmdLineLeft)
		addLookupExp(67, 'R', CmdLineRight)
		addLookupExp(68, 'A', CmdWordAdvance)
		addLookupExp(69, 'D', CmdWordDelete)

		// Z prefix - cursor commands }  {70}
		addLookupExp(70, 'U', CmdUp)
		addLookupExp(71, 'D', CmdDown)
		addLookupExp(72, 'R', CmdRight)
		addLookupExp(73, 'L', CmdLeft)
		addLookupExp(74, 'H', CmdHome)
		addLookupExp(75, 'C', CmdReturn)
		addLookupExp(76, 'T', CmdTab)
		addLookupExp(77, 'B', CmdBacktab)
		addLookupExp(78, 'Z', CmdRubout)

		// ~ prefix - miscellaneous debugging commands}  {79}
		addLookupExp(79, 'V', CmdValidate)
		addLookupExp(80, 'D', CmdDump)

		// sentinel }                    {81}
		addLookupExp(81, '?', CmdNoSuch)

		// initialize lookupexp_ptr }
		// These magic numbers point to the start of each section in lookupexp table }
		LookupExpPtr[CmdPrefixAst] = 1
		LookupExpPtr[CmdPrefixA] = 4
		LookupExpPtr[CmdPrefixB] = 4
		LookupExpPtr[CmdPrefixC] = 5
		LookupExpPtr[CmdPrefixD] = 5
		LookupExpPtr[CmdPrefixE] = 5
		LookupExpPtr[CmdPrefixEo] = 13
		LookupExpPtr[CmdPrefixEq] = 16
		LookupExpPtr[CmdPrefixF] = 19
		LookupExpPtr[CmdPrefixFg] = 29
		LookupExpPtr[CmdPrefixI] = 35
		LookupExpPtr[CmdPrefixK] = 35
		LookupExpPtr[CmdPrefixL] = 35
		LookupExpPtr[CmdPrefixO] = 35
		LookupExpPtr[CmdPrefixP] = 35
		LookupExpPtr[CmdPrefixS] = 35
		LookupExpPtr[CmdPrefixT] = 44
		LookupExpPtr[CmdPrefixTc] = 44
		LookupExpPtr[CmdPrefixTf] = 44
		LookupExpPtr[CmdPrefixU] = 44
		LookupExpPtr[CmdPrefixW] = 48
		LookupExpPtr[CmdPrefixX] = 59
		LookupExpPtr[CmdPrefixY] = 62
		LookupExpPtr[CmdPrefixZ] = 70
		LookupExpPtr[CmdPrefixTilde] = 79
		LookupExpPtr[CmdNoSuch] = 81
	} else {
		Lookup[lookupIndex(0)].Command = CmdNoop
		Lookup[lookupIndex(1)].Command = CmdNoop
		Lookup[lookupIndex(2)].Command = CmdWindowBackward
		Lookup[lookupIndex(3)].Command = CmdNoop
		Lookup[lookupIndex(4)].Command = CmdDeleteChar
		Lookup[lookupIndex(5)].Command = CmdWindowEnd
		Lookup[lookupIndex(6)].Command = CmdWindowForward
		Lookup[lookupIndex(7)].Command = CmdDoLastCommand
		Lookup[lookupIndex(8)].Command = CmdLeft
		Lookup[lookupIndex(9)].Command = CmdTab
		Lookup[lookupIndex(10)].Command = CmdDown
		Lookup[lookupIndex(11)].Command = CmdDeleteLine
		Lookup[lookupIndex(12)].Command = CmdInsertLine
		Lookup[lookupIndex(13)].Command = CmdReturn
		Lookup[lookupIndex(14)].Command = CmdWindowNew
		Lookup[lookupIndex(15)].Command = CmdNoop
		Lookup[lookupIndex(16)].Command = CmdUserCommandIntroducer
		Lookup[lookupIndex(17)].Command = CmdNoop
		Lookup[lookupIndex(18)].Command = CmdRight
		Lookup[lookupIndex(19)].Command = CmdNoop
		Lookup[lookupIndex(20)].Command = CmdWindowTop
		Lookup[lookupIndex(21)].Command = CmdUp
		Lookup[lookupIndex(22)].Command = CmdNoop
		Lookup[lookupIndex(23)].Command = CmdWordAdvance
		Lookup[lookupIndex(24)].Command = CmdNoop
		Lookup[lookupIndex(25)].Command = CmdNoop
		Lookup[lookupIndex(26)].Command = CmdUserParent
		Lookup[lookupIndex(27)].Command = CmdNoop
		Lookup[lookupIndex(28)].Command = CmdNoop
		Lookup[lookupIndex(29)].Command = CmdNoop
		Lookup[lookupIndex(30)].Command = CmdInsertChar
		Lookup[lookupIndex(31)].Command = CmdNoop
		Lookup[lookupIndex(' ')].Command = CmdNoop
		Lookup[lookupIndex('!')].Command = CmdNoop
		Lookup[lookupIndex('"')].Command = CmdDittoUp
		Lookup[lookupIndex('#')].Command = CmdNoop
		Lookup[lookupIndex('$')].Command = CmdNoop
		Lookup[lookupIndex('%')].Command = CmdNoop
		Lookup[lookupIndex('&')].Command = CmdNoop
		Lookup['\''].Command = CmdDittoDown
		Lookup[lookupIndex('(')].Command = CmdNoop
		Lookup[lookupIndex(')')].Command = CmdNoop
		Lookup[lookupIndex('*')].Command = CmdNoop
		Lookup[lookupIndex('+')].Command = CmdNoop
		Lookup[lookupIndex(';')].Command = CmdNoop
		Lookup[lookupIndex('-')].Command = CmdNoop
		Lookup[lookupIndex('.')].Command = CmdNoop
		Lookup[lookupIndex('/')].Command = CmdNoop
		for keyCode = '0'; keyCode <= '9'; keyCode++ {
			Lookup[lookupIndex(keyCode)].Command = CmdNoop
		}
		Lookup[lookupIndex(':')].Command = CmdNoop
		Lookup[lookupIndex(';')].Command = CmdNoop
		Lookup[lookupIndex('<')].Command = CmdNoop
		Lookup[lookupIndex('=')].Command = CmdNoop
		Lookup[lookupIndex('>')].Command = CmdNoop
		Lookup[lookupIndex('?')].Command = CmdNoop
		Lookup[lookupIndex('@')].Command = CmdNoop
		Lookup[lookupIndex('A')].Command = CmdPrefixA
		Lookup[lookupIndex('B')].Command = CmdPrefixB
		Lookup[lookupIndex('C')].Command = CmdPrefixC
		Lookup[lookupIndex('D')].Command = CmdPrefixD
		Lookup[lookupIndex('E')].Command = CmdPrefixE
		Lookup[lookupIndex('F')].Command = CmdPrefixF
		Lookup[lookupIndex('G')].Command = CmdGet
		Lookup[lookupIndex('H')].Command = CmdHelp
		Lookup[lookupIndex('I')].Command = CmdNoop
		Lookup[lookupIndex('J')].Command = CmdNoop
		Lookup[lookupIndex('K')].Command = CmdPrefixK
		Lookup[lookupIndex('L')].Command = CmdPrefixL
		Lookup[lookupIndex('M')].Command = CmdMark
		Lookup[lookupIndex('N')].Command = CmdNoop
		Lookup[lookupIndex('O')].Command = CmdPrefixO
		Lookup[lookupIndex('P')].Command = CmdPrefixP
		Lookup[lookupIndex('Q')].Command = CmdQuit
		Lookup[lookupIndex('R')].Command = CmdReplace
		Lookup[lookupIndex('S')].Command = CmdPrefixS
		Lookup[lookupIndex('T')].Command = CmdPrefixT
		Lookup[lookupIndex('U')].Command = CmdPrefixU
		Lookup[lookupIndex('V')].Command = CmdVerify
		Lookup[lookupIndex('W')].Command = CmdPrefixW
		Lookup[lookupIndex('X')].Command = CmdPrefixX
		Lookup[lookupIndex('Y')].Command = CmdNoop
		Lookup[lookupIndex('Z')].Command = CmdNoop
		Lookup[lookupIndex('[')].Command = CmdNoop
		Lookup['\\'].Command = CmdCommand
		Lookup[lookupIndex(']')].Command = CmdNoop
		Lookup[lookupIndex('^')].Command = CmdNoop
		Lookup[lookupIndex('_')].Command = CmdNoop
		Lookup[lookupIndex('`')].Command = CmdNoop
		for keyCode = 'a'; keyCode <= 'z'; keyCode++ {
			Lookup[lookupIndex(keyCode)].Command = CmdNoop
		}
		Lookup[lookupIndex('{')].Command = CmdSetMarginLeft
		Lookup[lookupIndex('|')].Command = CmdNoop
		Lookup[lookupIndex('}')].Command = CmdSetMarginRight
		Lookup[lookupIndex('~')].Command = CmdPrefixTilde
		Lookup[lookupIndex(127)].Command = CmdRubout
		for keyCode = 128; keyCode <= OrdMaxChar; keyCode++ {
			Lookup[lookupIndex(keyCode)].Command = CmdNoop
		}
		// for keyCode = -MaxSpecialKeys; keyCode <= OrdMaxChar; keyCode++ {
		// 	// with lookup[key_code] do
		// 	Lookup[lookupIndex(keyCode)].Code = nil
		// 	Lookup[lookupIndex(keyCode)].Tpar = nil
		// }

		// initialize lookupexp
		// Ast ( * ) prefix } {start at 1}
		// There aren't any in this table!

		// A prefix }    {start at 1}
		addLookupExp(1, 'C', CmdJump)
		addLookupExp(2, 'L', CmdAdvance)
		addLookupExp(3, 'O', CmdBridge)
		addLookupExp(4, 'P', CmdAdvanceParagraph)
		addLookupExp(5, 'S', CmdNoop)
		addLookupExp(6, 'T', CmdNext)
		addLookupExp(7, 'W', CmdWordAdvance)

		// B prefix }    {8}
		addLookupExp(8, 'B', CmdNoop)
		addLookupExp(9, 'C', CmdNoop)  // cmd_block_copy
		addLookupExp(10, 'D', CmdNoop) // cmd_block_define
		addLookupExp(11, 'I', CmdNoop)
		addLookupExp(12, 'K', CmdNoop)
		addLookupExp(13, 'M', CmdNoop) // cmd_block_transfer
		addLookupExp(14, 'O', CmdNoop)

		// C prefix }    {15}
		addLookupExp(15, 'C', CmdInsertChar)
		addLookupExp(16, 'L', CmdInsertLine)

		// D prefix }    {17}
		addLookupExp(17, 'C', CmdDeleteChar)
		addLookupExp(18, 'L', CmdDeleteLine)
		addLookupExp(19, 'P', CmdDeleteParagraph)
		addLookupExp(20, 'S', CmdNoop)
		addLookupExp(21, 'W', CmdWordDelete)

		// E prefix }    {22}
		addLookupExp(22, 'D', CmdFrameEdit)
		addLookupExp(23, 'K', CmdFrameKill)
		addLookupExp(24, 'O', CmdPrefixEo)
		addLookupExp(25, 'P', CmdFrameParameters)
		addLookupExp(26, 'Q', CmdPrefixEq)
		addLookupExp(27, 'R', CmdFrameReturn)

		// EO prefix }   {28}
		addLookupExp(28, 'L', CmdEqualEol)
		addLookupExp(29, 'F', CmdEqualEof)
		addLookupExp(30, 'P', CmdEqualEop)

		// EQ prefix }   {31}
		addLookupExp(31, 'C', CmdEqualColumn)
		addLookupExp(32, 'L', CmdNoop)
		addLookupExp(33, 'M', CmdEqualMark)
		addLookupExp(34, 'S', CmdEqualString)

		// F prefix - files }    {35}
		addLookupExp(35, 'S', CmdFileSave)
		addLookupExp(36, 'B', CmdFileRewind)
		addLookupExp(37, 'E', CmdFileEdit)
		addLookupExp(38, 'G', CmdPrefixFg)
		addLookupExp(39, 'I', CmdFileInput)
		addLookupExp(40, 'K', CmdFileKill)
		addLookupExp(41, 'O', CmdFileOutput)
		addLookupExp(42, 'P', CmdPage)
		addLookupExp(43, 'S', CmdNoop)
		addLookupExp(44, 'T', CmdFileTable)
		addLookupExp(45, 'X', CmdFileExecute)

		// FG prefix - global files }    {46}
		addLookupExp(46, 'B', CmdFileGlobalRewind)
		addLookupExp(47, 'I', CmdFileGlobalInput)
		addLookupExp(48, 'K', CmdFileGlobalKill)
		addLookupExp(49, 'O', CmdFileGlobalOutput)
		addLookupExp(50, 'R', CmdFileRead)
		addLookupExp(51, 'W', CmdFileWrite)

		// I prefix }    {52}
		// There aren't any yet! }

		// K prefix }    {52}
		addLookupExp(52, 'B', CmdBacktab)
		addLookupExp(53, 'C', CmdReturn)
		addLookupExp(54, 'D', CmdDown)
		addLookupExp(55, 'H', CmdHome)
		addLookupExp(56, 'I', CmdInsertMode)
		addLookupExp(57, 'L', CmdLeft)
		addLookupExp(58, 'M', CmdUserKey)
		addLookupExp(59, 'O', CmdOvertypeMode)
		addLookupExp(60, 'R', CmdRight)
		addLookupExp(61, 'T', CmdTab)
		addLookupExp(62, 'U', CmdUp)
		addLookupExp(63, 'X', CmdRubout)

		// L prefix }    {64}
		addLookupExp(64, 'R', CmdNoop)
		addLookupExp(65, 'S', CmdNoop)

		// O prefix }    {66}
		addLookupExp(66, 'P', CmdUserParent)
		addLookupExp(67, 'S', CmdUserSubprocess)
		addLookupExp(68, 'X', CmdOpSysCommand)

		// P prefix }    {69}
		addLookupExp(69, 'C', CmdPositionColumn)
		addLookupExp(70, 'L', CmdPositionLine)

		// S prefix }    {71}
		addLookupExp(71, 'A', CmdSpanAssign)
		addLookupExp(72, 'C', CmdSpanCopy)
		addLookupExp(73, 'D', CmdSpanDefine)
		addLookupExp(74, 'E', CmdSpanExecuteNoRecompile)
		addLookupExp(75, 'J', CmdSpanJump)
		addLookupExp(76, 'M', CmdSpanTransfer)
		addLookupExp(77, 'R', CmdSpanCompile)
		addLookupExp(78, 'T', CmdSpanIndex)
		addLookupExp(79, 'X', CmdSpanExecute)

		// T prefix }    {80}
		addLookupExp(80, 'B', CmdSplitLine)
		addLookupExp(81, 'C', CmdPrefixTc)
		addLookupExp(82, 'F', CmdPrefixTf)
		addLookupExp(83, 'I', CmdInsertText)
		addLookupExp(84, 'N', CmdInsertInvisible)
		addLookupExp(85, 'O', CmdOvertypeText)
		addLookupExp(86, 'R', CmdNoop)
		addLookupExp(87, 'S', CmdSwapLine)
		addLookupExp(88, 'X', CmdExecuteString)

		// TC prefix }   {89}
		addLookupExp(89, 'E', CmdCaseEdit)
		addLookupExp(90, 'L', CmdCaseLow)
		addLookupExp(91, 'U', CmdCaseUp)

		// TF prefix }   {92}
		addLookupExp(92, 'C', CmdLineCentre)
		addLookupExp(93, 'F', CmdLineFill)
		addLookupExp(94, 'J', CmdLineJustify)
		addLookupExp(95, 'L', CmdLineLeft)
		addLookupExp(96, 'R', CmdLineRight)
		addLookupExp(97, 'S', CmdLineSquash)

		// U prefix - user keyboard mappings }   {98}
		addLookupExp(98, 'C', CmdUserCommandIntroducer)

		// W prefix - window commands }  {99}
		addLookupExp(99, 'B', CmdWindowBackward)
		addLookupExp(100, 'C', CmdWindowMiddle)
		addLookupExp(101, 'E', CmdWindowEnd)
		addLookupExp(102, 'F', CmdWindowForward)
		addLookupExp(103, 'H', CmdWindowSetHeight)
		addLookupExp(104, 'L', CmdWindowLeft)
		addLookupExp(105, 'M', CmdWindowScroll)
		addLookupExp(106, 'N', CmdWindowNew)
		addLookupExp(107, 'O', CmdNoop)
		addLookupExp(108, 'R', CmdWindowRight)
		addLookupExp(109, 'S', CmdNoop)
		addLookupExp(110, 'T', CmdWindowTop)
		addLookupExp(111, 'U', CmdWindowUpdate)

		// X prefix - exit }             {112}
		addLookupExp(112, 'A', CmdExitAbort)
		addLookupExp(113, 'F', CmdExitFail)
		addLookupExp(114, 'S', CmdExitSuccess)

		// Y prefix }        {115}
		// There aren't any in this table! }

		// Z prefix }        {115}
		// There aren't any in this table! }

		// ~ prefix - miscellaneous debugging commands}  {115}
		addLookupExp(115, 'D', CmdDump)
		addLookupExp(116, 'V', CmdValidate)

		// sentinel }                    {117}
		addLookupExp(117, '?', CmdNoSuch)

		// initialize lookupexp_ptr }
		// These magic numbers point to the start of each section in lookupexp table }
		LookupExpPtr[CmdPrefixAst] = 1
		LookupExpPtr[CmdPrefixA] = 1
		LookupExpPtr[CmdPrefixB] = 8
		LookupExpPtr[CmdPrefixC] = 15
		LookupExpPtr[CmdPrefixD] = 17
		LookupExpPtr[CmdPrefixE] = 22
		LookupExpPtr[CmdPrefixEo] = 28
		LookupExpPtr[CmdPrefixEq] = 31
		LookupExpPtr[CmdPrefixF] = 35
		LookupExpPtr[CmdPrefixFg] = 46
		LookupExpPtr[CmdPrefixI] = 52
		LookupExpPtr[CmdPrefixK] = 52
		LookupExpPtr[CmdPrefixL] = 64
		LookupExpPtr[CmdPrefixO] = 66
		LookupExpPtr[CmdPrefixP] = 69
		LookupExpPtr[CmdPrefixS] = 71
		LookupExpPtr[CmdPrefixT] = 80
		LookupExpPtr[CmdPrefixTc] = 89
		LookupExpPtr[CmdPrefixTf] = 92
		LookupExpPtr[CmdPrefixU] = 98
		LookupExpPtr[CmdPrefixW] = 99
		LookupExpPtr[CmdPrefixX] = 112
		LookupExpPtr[CmdPrefixY] = 115
		LookupExpPtr[CmdPrefixZ] = 115
		LookupExpPtr[CmdPrefixTilde] = 115
		LookupExpPtr[CmdNoSuch] = 117
	}
}

func startUp(argc int, argv []string) bool {
	const frameNameCmd = "COMMAND"
	const frameNameOops = "OOPS"
	const frameNameHeap = "HEAP"

	result := false

	// Get the command line.
	var commandLine string
	if argc > 1 {
		commandLine = strings.Join(argv[1:], " ")
	}

	if len(commandLine) > FileNameLen {
		ScreenMessage(MsgParameterTooLong)
		goto l99
	}

	// Open the files.
	if !FileCreateOpen(&commandLine, ParseCommand, &Files[1], &Files[2]) {
		goto l99
	}

	loadCommandTable(FileData.OldCmds)

	// Try to get started on the terminal.  If this fails assume carry on
	// in BATCH mode.
	LudwigMode = LudwigBatch
	if VduInit(&TerminalInfo, &TtControlC, &TtWinChanged) {
		InitialScrWidth = TerminalInfo.Width
		InitialScrHeight = TerminalInfo.Height
		InitialMarginRight = TerminalInfo.Width
		//        if (trmflags_v_hard & tt_capabilities) {
		//            ludwig_mode = ludwig_mode_type::ludwig_hardcopy;
		//        } else
		{
			LudwigMode = LudwigScreen
			VduNewIntroducer(CommandIntroducer)
		}
	}
	// Set the scr_msg_row as one more than the terminal height (which may
	// be zero). This avoids any need for special checks about Ludwig being
	// in Screen mode before clearing messages.

	ScrMsgRow = TerminalInfo.Height + 1

	// Create the three automatically defined frames: OOPS, COMMAND and LUDWIG.
	// Save pointers to COMMAND & OOPS  frames for use in later frame routines.

	if !FrameEdit(frameNameOops) {
		goto l99
	}
	if !FrameSetHeight(InitialScrHeight, true) {
		goto l99
	}
	FrameOops = CurrentFrame
	CurrentFrame = nil
	FrameOops.SpaceLimit = MaxSpace     // Big !
	FrameOops.SpaceLeft = MaxSpace - 50 // Big ! - space for <eop> line !!
	FrameOops.Options.Set(OptSpecialFrame)
	if !FrameEdit(frameNameCmd) {
		goto l99
	}
	FrameCmd = CurrentFrame
	CurrentFrame = nil
	FrameCmd.Options.Set(OptSpecialFrame)
	if !FrameEdit(frameNameHeap) {
		goto l99
	}
	FrameHeap = CurrentFrame
	CurrentFrame = nil
	FrameHeap.Options.Set(OptSpecialFrame)
	{
		if !FrameEdit(DefaultFrameName) {
			goto l99
		}
	}

	if LudwigMode == LudwigScreen {
		ScreenFixup()
	}

	// Load the key definitions.

	if LudwigMode == LudwigScreen {
		UserKeyInitialize()
	}

	// Hook our input and output files into the current frame.

	// with current_frame^ do
	if Files[1] != nil {
		CurrentFrame.InputFile = 1
		FilesFrames[1] = CurrentFrame
	}
	if Files[2] != nil {
		CurrentFrame.OutputFile = 2
		FilesFrames[2] = CurrentFrame
	}

	// Load the input file.

	if LudwigMode != LudwigBatch {
		ScreenMessage(MsgCopyrightAndLoadingFile)
		if LudwigMode == LudwigScreen {
			VduFlush()
		}
	}
	if !FilePage(CurrentFrame, &ExitAbort) {
		goto l99
	}
	if LudwigMode != LudwigBatch {
		ScreenClearMsgs(false)
	}
	if LudwigMode == LudwigScreen {
		ScreenFixup()
	}

	// Execute the user's initialization string.

	if FileData.Initial != "" {
		if LudwigMode == LudwigScreen {
			VduFlush()
		}
		tparam := &TParObject{}
		// with tparam^ do
		tparam.Len = len(FileData.Initial)
		tparam.Dlm = TpdExact
		tparam.Str.Assign(FileData.Initial)
		tparam.Nxt = nil
		tparam.Con = nil
		if !Execute(CmdFileExecute, LeadParamNone, 1, tparam, true) {
			if ExitAbort {
				// something is wrong, but let the user continue anyway!
				if LudwigMode != LudwigBatch {
					ScreenBeep()
				}
				ExitAbort = false
			}
		}
	}

	// Set the Abort Flag now.  This will suppress spurious start-up messages
	LudwigAborted = true
	result = true
l99:
	return result
}

func main() {
	SysInitSig()
	ValueInitializations()
	initialize()                        // Stuff VALUE can't do, like creating frames etc.
	if startUp(len(os.Args), os.Args) { // Parse command line, get files attached, etc.
		ExecuteImmed()
		SysExitSuccess()
	}
	if LudwigAborted {
		SysExitFailure()
	}
	SysExitSuccess()
}
