/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         VDU
//
// Description:  This module does all the complex control of the VDU type
//               screens that Ludwig demands. This is the ncurses version.

package ludwig

import (
	"os"
	"strconv"
	"time"

	nc "ludwig-go/internal/ncurses"
)

// Constants
const (
	BS  = 8
	NL  = 10
	CR  = 13
	SPC = 32
	DEL = 127

	OutMClearEOL    = 1
	NumControlChars = 33
)

// Key code ranges
const (
	MinNormalCode = 0
	MaxNormalCode = OrdMaxChar
)

// Ncurses key range constants
var (
	MinCursesKey    int
	MaxCursesKey    int
	NumNcursesKeys  int
	NcursesSubtract int
	MassagedMax     int
)

// Global variables
var (
	controlChars map[int]bool
	terminators  map[int]bool
	vduSetup     bool
	inInsertMode bool
	gCtrlC       *bool
	gWinChange   *bool
	stdscr       *nc.Window
	refreshDelay int
)

func init() {
	// Initialize control chars set
	controlChars = map[int]bool{
		0x00: true, 0x01: true, 0x02: true, 0x03: true,
		0x04: true, 0x05: true, 0x06: true, 0x07: true,
		0x08: true, 0x09: true, 0x0A: true, 0x0B: true,
		0x0C: true, 0x0D: true, 0x0E: true, 0x0F: true,
		0x10: true, 0x11: true, 0x12: true, 0x13: true,
		0x14: true, 0x15: true, 0x16: true, 0x17: true,
		0x18: true, 0x19: true, 0x1A: true, 0x1B: true,
		0x1C: true, 0x1D: true, 0x1E: true, 0x1F: true,
		0x7F: true,
	}
	terminators = make(map[int]bool)
	value := os.Getenv("LUD_REFRESH_DELAY")
	if value == "" {
		refreshDelay = 0
	} else {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			refreshDelay = 0
		} else {
			refreshDelay = parsed
		}
	}
}

// massageKey converts ncurses key codes to Ludwig key codes
func massageKey(keyCode int) int {
	if keyCode >= MinNormalCode && keyCode <= MaxNormalCode {
		return keyCode
	} else if keyCode >= MinCursesKey && keyCode <= MaxCursesKey {
		if keyCode == nc.KEY_BACKSPACE {
			return DEL
		}
		return keyCode
	}
	return 0
}

// unmassageKey converts Ludwig key codes back to ncurses key codes
func unmassageKey(key int) int {
	if key >= MinNormalCode && key <= MaxNormalCode {
		return key
	} else if key > MaxNormalCode && key <= MassagedMax {
		return key
	}
	return 0
}

// VduMoveCurs moves the cursor to the specified position (1-based)
func VduMoveCurs(x, y int) {
	stdscr.Move(y-1, x-1)
}

// VduFlush refreshes the screen
func VduFlush() {
	stdscr.Refresh()
	if refreshDelay > 0 {
		time.Sleep(time.Duration(refreshDelay) * time.Millisecond)
	}
}

// VduBeep produces a beep or flash
func VduBeep() {
	nc.Flash()
}

// VduClearEOL clears from cursor to end of line
func VduClearEOL() {
	stdscr.ClearToEOL()
}

// VduDisplayStr displays a string with optional clear to end of line
func VduDisplayStr(str string, opts int) {
	_, maxX := stdscr.MaxYX()
	_, curX := stdscr.CursorYX()
	maxlen := maxX - curX

	slen := len(str)
	hitMargin := false
	if slen >= maxlen {
		slen = maxlen
		hitMargin = true
	}

	stdscr.Print(str[:slen])

	if !hitMargin && (opts&OutMClearEOL) != 0 {
		VduClearEOL()
	}
}

// VduDisplayCh displays a single character
func VduDisplayCh(ch byte) {
	stdscr.AddChar(nc.Char(ch))
}

// VduClearScr clears the entire screen
func VduClearScr() {
	stdscr.Clear()
}

// VduClearEOS clears from cursor to end of screen
func VduClearEOS() {
	stdscr.ClearToBottom()
}

// VduScrollUp scrolls the screen up by n lines
func VduScrollUp(n int) {
	stdscr.ScrollOk(true)
	stdscr.Scroll(n)
	stdscr.ScrollOk(false)
}

// VduDeleteLines deletes n lines at current position
func VduDeleteLines(n int) {
	stdscr.InsDelLines(-n)
}

// VduInsertLines inserts n lines at current position
func VduInsertLines(n int) {
	stdscr.InsDelLines(n)
}

// VduInsertChars inserts n characters at current position
func VduInsertChars(n int) {
	for range n {
		stdscr.InsChar(nc.Char(' '))
	}
}

// VduDeleteChars deletes n characters at current position
func VduDeleteChars(n int) {
	for range n {
		stdscr.DelChar()
	}
}

// VduDisplayCrLf displays a carriage return / line feed
func VduDisplayCrLf() {
	y, _ := stdscr.CursorYX()
	maxY, _ := stdscr.MaxYX()

	if y == maxY-1 {
		VduScrollUp(1)
	} else {
		y++
	}

	stdscr.Move(y, 0)
	stdscr.Refresh()
}

// VduTakeBackKey pushes a key back to the input queue
func VduTakeBackKey(key int) {
	nc.UnGetChar(nc.Char(unmassageKey(key)))
}

// VduNewIntroducer sets up terminators for input
func VduNewIntroducer(key int) {
	terminators = make(map[int]bool)
	for k := range controlChars {
		terminators[k] = true
	}
	if key > 0 {
		terminators[key] = true
	}
}

// VduGetKey gets a single key from the user
func VduGetKey() int {
	nc.CursSet(1)
	VduFlush()
	var rawKey nc.Key
	for {
		rawKey = stdscr.GetChar()
		if rawKey != 0 {
			break
		}
	}

	if rawKey == nc.KEY_RESIZE && gWinChange != nil {
		*gWinChange = true
	}
	nc.CursSet(0)
	return massageKey(int(rawKey))
}

// VduGetInput gets a line of input from the user with a prompt
func VduGetInput(prompt string, get **StrObject, getLen int, outlen *int) {
	VduBold()
	VduDisplayStr(prompt, OutMClearEOL)
	VduNormal()

	// Fill get with spaces
	*get = NewBlankStrObject(MaxStrLen)

	_, curX := stdscr.CursorYX()
	maxY, maxX := stdscr.MaxYX()
	maxlen := min(MaxStrLen, maxX-curX)

	if getLen > maxlen {
		getLen = maxlen
	}

	*outlen = 0
	key := VduGetKey()

	for getLen > 0 && key != CR && key != NL {
		if *outlen > 0 && (key == BS || key == DEL) {
			getLen++
			*outlen--
			stdscr.AddChar(nc.Char(BS))
			stdscr.AddChar(nc.Char(SPC))
			stdscr.AddChar(nc.Char(BS))
		} else {
			if key < 0 || key > OrdMaxChar || controlChars[key] {
				VduBeep()
			} else {
				getLen--
				*outlen++
				(*get).Set(*outlen, byte(key))
				stdscr.AddChar(nc.Char(key))
			}
		}
		key = VduGetKey()
	}
	_ = maxY // Avoid unused variable warning
}

// VduInsertMode sets insert mode on or off
func VduInsertMode(turnOn bool) {
	inInsertMode = turnOn
}

// VduGetText gets text input from the user
func VduGetText(strLen int, str *StrObject, outlen *int) {
	// Fill str with spaces
	str.Fill(' ', 1, MaxStrLen)

	*outlen = 0
	_, curX := stdscr.CursorYX()
	maxY, maxX := stdscr.MaxYX()
	maxlen := maxX - curX

	if strLen > maxlen {
		strLen = maxlen
	}

	for strLen > 0 {
		key := VduGetKey()
		if key < 0 || key > OrdMaxChar || terminators[key] {
			VduTakeBackKey(key)
			strLen = 0
		} else {
			if inInsertMode {
				VduInsertChars(1)
			}
			stdscr.AddChar(nc.Char(key))
			stdscr.Refresh()
			*outlen++
			str.Set(*outlen, byte(key))
			strLen--
		}
	}
	_ = maxY // Avoid unused variable warning
}

// VduKeyboardInit initializes keyboard mappings
func VduKeyboardInit(
	nrKeyNames *int,
	keyNameList *[]KeyNameRecord,
	keyIntroducers *[MaxSetRange + 1]bool,
	terminalInfo *TerminalInfoType,
) {
	*nrKeyNames = NumControlChars + NumNcursesKeys
	kl := make([]KeyNameRecord, *nrKeyNames+1)

	// Initialize control character names
	for i := 1; i < NumControlChars; i++ {
		kl[i].KeyCode = i - 1
	}
	kl[1].KeyName = "CONTROL-@"
	kl[2].KeyName = "CONTROL-A"
	kl[3].KeyName = "CONTROL-B"
	kl[4].KeyName = "CONTROL-C"
	kl[5].KeyName = "CONTROL-D"
	kl[6].KeyName = "CONTROL-E"
	kl[7].KeyName = "CONTROL-F"
	kl[8].KeyName = "CONTROL-G"
	kl[9].KeyName = "BACKSPACE"
	kl[10].KeyName = "TAB"
	kl[11].KeyName = "LINE-FEED"
	kl[12].KeyName = "CONTROL-K"
	kl[13].KeyName = "CONTROL-L"
	kl[14].KeyName = "RETURN"
	kl[15].KeyName = "CONTROL-N"
	kl[16].KeyName = "CONTROL-O"
	kl[17].KeyName = "CONTROL-P"
	kl[18].KeyName = "CONTROL-Q"
	kl[19].KeyName = "CONTROL-R"
	kl[20].KeyName = "CONTROL-S"
	kl[21].KeyName = "CONTROL-T"
	kl[22].KeyName = "CONTROL-U"
	kl[23].KeyName = "CONTROL-V"
	kl[24].KeyName = "CONTROL-W"
	kl[25].KeyName = "CONTROL-X"
	kl[26].KeyName = "CONTROL-Y"
	kl[27].KeyName = "CONTROL-Z"
	kl[28].KeyName = "CONTROL-["
	kl[29].KeyName = "CONTROL-\\"
	kl[30].KeyName = "CONTROL-]"
	kl[31].KeyName = "CONTROL-^"
	kl[32].KeyName = "CONTROL-_"
	kl[33].KeyCode = DEL
	kl[33].KeyName = "DELETE"

	// Initialize ncurses key names
	for i := MinCursesKey; i <= MaxCursesKey; i++ {
		kl[NumControlChars+1+i-MinCursesKey].KeyCode = massageKey(i)
	}

	kl[NumControlChars+1].KeyName = "BREAK"
	kl[NumControlChars+2].KeyName = "DOWN-ARROW"
	kl[NumControlChars+3].KeyName = "UP-ARROW"
	kl[NumControlChars+4].KeyName = "LEFT-ARROW"
	kl[NumControlChars+5].KeyName = "RIGHT-ARROW"
	kl[NumControlChars+6].KeyName = "HOME"
	kl[NumControlChars+7].KeyName = "BACKSPACE"
	kl[NumControlChars+8].KeyName = "FUNCTION-0"
	kl[NumControlChars+9].KeyName = "FUNCTION-1"
	kl[NumControlChars+10].KeyName = "FUNCTION-2"
	kl[NumControlChars+11].KeyName = "FUNCTION-3"
	kl[NumControlChars+12].KeyName = "FUNCTION-4"
	kl[NumControlChars+13].KeyName = "FUNCTION-5"
	kl[NumControlChars+14].KeyName = "FUNCTION-6"
	kl[NumControlChars+15].KeyName = "FUNCTION-7"
	kl[NumControlChars+16].KeyName = "FUNCTION-8"
	kl[NumControlChars+17].KeyName = "FUNCTION-9"
	kl[NumControlChars+18].KeyName = "FUNCTION-10"
	kl[NumControlChars+19].KeyName = "FUNCTION-11"
	kl[NumControlChars+20].KeyName = "FUNCTION-12"
	kl[NumControlChars+21].KeyName = "SHIFT-FUNCTION-1"
	kl[NumControlChars+22].KeyName = "SHIFT-FUNCTION-2"
	kl[NumControlChars+23].KeyName = "SHIFT-FUNCTION-3"
	kl[NumControlChars+24].KeyName = "SHIFT-FUNCTION-4"
	kl[NumControlChars+25].KeyName = "SHIFT-FUNCTION-5"
	kl[NumControlChars+26].KeyName = "SHIFT-FUNCTION-6"
	kl[NumControlChars+27].KeyName = "SHIFT-FUNCTION-7"
	kl[NumControlChars+28].KeyName = "SHIFT-FUNCTION-8"
	kl[NumControlChars+29].KeyName = "SHIFT-FUNCTION-9"
	kl[NumControlChars+30].KeyName = "SHIFT-FUNCTION-10"
	kl[NumControlChars+31].KeyName = "SHIFT-FUNCTION-11"
	kl[NumControlChars+32].KeyName = "SHIFT-FUNCTION-12"
	kl[NumControlChars+33].KeyName = "FUNCTION-25"
	kl[NumControlChars+34].KeyName = "FUNCTION-26"
	kl[NumControlChars+35].KeyName = "FUNCTION-27"
	kl[NumControlChars+36].KeyName = "FUNCTION-28"
	kl[NumControlChars+37].KeyName = "FUNCTION-29"
	kl[NumControlChars+38].KeyName = "FUNCTION-30"
	kl[NumControlChars+39].KeyName = "FUNCTION-31"
	kl[NumControlChars+40].KeyName = "FUNCTION-32"
	kl[NumControlChars+41].KeyName = "FUNCTION-33"
	kl[NumControlChars+42].KeyName = "FUNCTION-34"
	kl[NumControlChars+43].KeyName = "FUNCTION-35"
	kl[NumControlChars+44].KeyName = "FUNCTION-36"
	kl[NumControlChars+45].KeyName = "FUNCTION-37"
	kl[NumControlChars+46].KeyName = "FUNCTION-38"
	kl[NumControlChars+47].KeyName = "FUNCTION-39"
	kl[NumControlChars+48].KeyName = "FUNCTION-40"
	kl[NumControlChars+49].KeyName = "FUNCTION-41"
	kl[NumControlChars+50].KeyName = "FUNCTION-42"
	kl[NumControlChars+51].KeyName = "FUNCTION-43"
	kl[NumControlChars+52].KeyName = "FUNCTION-44"
	kl[NumControlChars+53].KeyName = "FUNCTION-45"
	kl[NumControlChars+54].KeyName = "FUNCTION-46"
	kl[NumControlChars+55].KeyName = "FUNCTION-47"
	kl[NumControlChars+56].KeyName = "FUNCTION-48"
	kl[NumControlChars+57].KeyName = "FUNCTION-49"
	kl[NumControlChars+58].KeyName = "FUNCTION-50"
	kl[NumControlChars+59].KeyName = "FUNCTION-51"
	kl[NumControlChars+60].KeyName = "FUNCTION-52"
	kl[NumControlChars+61].KeyName = "FUNCTION-53"
	kl[NumControlChars+62].KeyName = "FUNCTION-54"
	kl[NumControlChars+63].KeyName = "FUNCTION-55"
	kl[NumControlChars+64].KeyName = "FUNCTION-56"
	kl[NumControlChars+65].KeyName = "FUNCTION-57"
	kl[NumControlChars+66].KeyName = "FUNCTION-58"
	kl[NumControlChars+67].KeyName = "FUNCTION-59"
	kl[NumControlChars+68].KeyName = "FUNCTION-60"
	kl[NumControlChars+69].KeyName = "FUNCTION-61"
	kl[NumControlChars+70].KeyName = "FUNCTION-62"
	kl[NumControlChars+71].KeyName = "FUNCTION-63"
	kl[NumControlChars+72].KeyName = "DELETE-LINE"
	kl[NumControlChars+73].KeyName = "INSERT-LINE"
	kl[NumControlChars+74].KeyName = "DELETE-CHAR"
	kl[NumControlChars+75].KeyName = "INSERT-CHAR"
	kl[NumControlChars+76].KeyName = "EIC"
	kl[NumControlChars+77].KeyName = "CLEAR"
	kl[NumControlChars+78].KeyName = "CLEAR-EOS"
	kl[NumControlChars+79].KeyName = "CLEAR-EOL"
	kl[NumControlChars+80].KeyName = "SCROLL-FORWARD"
	kl[NumControlChars+81].KeyName = "SCROLL-REVERSE"
	kl[NumControlChars+82].KeyName = "PAGE-DOWN"
	kl[NumControlChars+83].KeyName = "PAGE-UP"
	kl[NumControlChars+84].KeyName = "SET-TAB"
	kl[NumControlChars+85].KeyName = "CLEAR-TAB"
	kl[NumControlChars+86].KeyName = "CLEAR-ALL-TABS"
	kl[NumControlChars+87].KeyName = "SEND"
	kl[NumControlChars+88].KeyName = "SOFT-RESET"
	kl[NumControlChars+89].KeyName = "RESET"
	kl[NumControlChars+90].KeyName = "PRINT"
	kl[NumControlChars+91].KeyName = "LOWER-LEFT"
	kl[NumControlChars+92].KeyName = "KEY-A1"
	kl[NumControlChars+93].KeyName = "KEY-A3"
	kl[NumControlChars+94].KeyName = "KEY-B2"
	kl[NumControlChars+95].KeyName = "KEY-C1"
	kl[NumControlChars+96].KeyName = "KEY-C3"
	kl[NumControlChars+97].KeyName = "BACK-TAB"
	kl[NumControlChars+98].KeyName = "BEGIN"
	kl[NumControlChars+99].KeyName = "CANCEL"
	kl[NumControlChars+100].KeyName = "CLOSE"
	kl[NumControlChars+101].KeyName = "COMMAND"
	kl[NumControlChars+102].KeyName = "COPY"
	kl[NumControlChars+103].KeyName = "CREATE"
	kl[NumControlChars+104].KeyName = "END"
	kl[NumControlChars+105].KeyName = "EXIT"
	kl[NumControlChars+106].KeyName = "FIND"
	kl[NumControlChars+107].KeyName = "HELP"
	kl[NumControlChars+108].KeyName = "MARK"
	kl[NumControlChars+109].KeyName = "MESSAGE"
	kl[NumControlChars+110].KeyName = "MOVE"
	kl[NumControlChars+111].KeyName = "NEXT"
	kl[NumControlChars+112].KeyName = "OPEN"
	kl[NumControlChars+113].KeyName = "OPTIONS"
	kl[NumControlChars+114].KeyName = "PREVIOUS"
	kl[NumControlChars+115].KeyName = "REDO"
	kl[NumControlChars+116].KeyName = "REFERENCE"
	kl[NumControlChars+117].KeyName = "REFRESH"
	kl[NumControlChars+118].KeyName = "REPLACE"
	kl[NumControlChars+119].KeyName = "RESTART"
	kl[NumControlChars+120].KeyName = "RESUME"
	kl[NumControlChars+121].KeyName = "SAVE"
	kl[NumControlChars+122].KeyName = "SHIFT-BEGIN"
	kl[NumControlChars+123].KeyName = "SHIFT-CANCEL"
	kl[NumControlChars+124].KeyName = "SHIFT-COMMAND"
	kl[NumControlChars+125].KeyName = "SHIFT-COPY"
	kl[NumControlChars+126].KeyName = "SHIFT-CREATE"
	kl[NumControlChars+127].KeyName = "SHIFT-DELETE-CHAR"
	kl[NumControlChars+128].KeyName = "SHIFT-DELETE-LINE"
	kl[NumControlChars+129].KeyName = "SELECT"
	kl[NumControlChars+130].KeyName = "SEND"
	kl[NumControlChars+131].KeyName = "SHIFT-CLEAR-EOL"
	kl[NumControlChars+132].KeyName = "SHIFT-EXIT"
	kl[NumControlChars+133].KeyName = "SHIFT-FIND"
	kl[NumControlChars+134].KeyName = "SHIFT-HELP"
	kl[NumControlChars+135].KeyName = "SHIFT-HOME"
	kl[NumControlChars+136].KeyName = "SHIFT-INSERT-CHAR"
	kl[NumControlChars+137].KeyName = "SHIFT-LEFT"
	kl[NumControlChars+138].KeyName = "SHIFT-MESSAGE"
	kl[NumControlChars+139].KeyName = "SHIFT-MOVE"
	kl[NumControlChars+140].KeyName = "SHIFT-NEXT"
	kl[NumControlChars+141].KeyName = "SHIFT-OPTIONS"
	kl[NumControlChars+142].KeyName = "SHIFT-PREVIOUS"
	kl[NumControlChars+143].KeyName = "SHIFT-PRINT"
	kl[NumControlChars+144].KeyName = "SHIFT-REDO"
	kl[NumControlChars+145].KeyName = "SHIFT-REPLACE"
	kl[NumControlChars+146].KeyName = "SHIFT-RIGHT"
	kl[NumControlChars+147].KeyName = "SHIFT-RESUME"
	kl[NumControlChars+148].KeyName = "SHIFT-SAVE"
	kl[NumControlChars+149].KeyName = "SHIFT-SUSPEND"
	kl[NumControlChars+150].KeyName = "SHIFT-UNDO"
	kl[NumControlChars+151].KeyName = "SUSPEND"
	kl[NumControlChars+152].KeyName = "UNDO"
	kl[NumControlChars+153].KeyName = "MOUSE"
	kl[NumControlChars+154].KeyName = "WINDOW-RESIZE-EVENT"
	kl[NumControlChars+155].KeyName = "SOME-OTHER-EVENT"

	// Clear key introducers
	for i := range keyIntroducers {
		keyIntroducers[i] = false
	}
	*keyNameList = make([]KeyNameRecord, *nrKeyNames)
	copy(*keyNameList, kl)
}

// VduInit initializes the VDU system
func VduInit(terminalInfo *TerminalInfoType, ctrlCFlag *bool, winchangeFlag *bool) bool {
	gCtrlC = ctrlCFlag
	gWinChange = winchangeFlag
	terminalInfo.Name = ""
	terminalInfo.Width = 80
	terminalInfo.Height = 4

	if SysIsTTY() {
		var err error
		stdscr, err = nc.Init()
		if err == nil {
			vduSetup = true
			nc.Raw(true)
			nc.Echo(false)
			nc.NewLines(false)
			nc.CursSet(0)
			stdscr.IntrFlush(false)
			stdscr.Keypad(true)
			stdscr.Idlok(true)
			stdscr.Idcok(true)
			stdscr.ScrollOk(false)

			// Initialize ncurses key range constants after Init
			MinCursesKey = 257
			MaxCursesKey = nc.KEY_MAX
			NumNcursesKeys = (MaxCursesKey - MinCursesKey) + 1
			NcursesSubtract = MinCursesKey - 1
			MassagedMax = MaxCursesKey

			maxY, maxX := stdscr.MaxYX()
			terminalInfo.Width = maxX
			terminalInfo.Height = maxY
			terminalInfo.Name = os.Getenv("TERM")

			VduClearScr()
		}
		return true
	}
	return false
}

// VduFree cleans up the VDU system
func VduFree() {
	if vduSetup {
		VduScrollUp(1)
		maxY, _ := stdscr.MaxYX()
		VduMoveCurs(1, maxY)
		VduFlush()
		nc.End()
	}
}

// VduGetNewDimensions gets the new screen dimensions after resize
func VduGetNewDimensions(newX *int, newY *int) {
	nc.End()
	stdscr.Refresh()
	maxY, maxX := stdscr.MaxYX()
	*newX = maxX
	*newY = maxY
}

// VduBold turns on bold attribute
func VduBold() {
	stdscr.AttrOn(nc.A_BOLD)
	stdscr.AttrOff(nc.A_DIM)
}

// VduDim turns on dim attribute
func VduDim() {
	stdscr.AttrOff(nc.A_BOLD)
	stdscr.AttrOn(nc.A_DIM)
}

// VduNormal turns off all attributes
func VduNormal() {
	stdscr.AttrOff(nc.A_BOLD)
	stdscr.AttrOff(nc.A_DIM)
}
