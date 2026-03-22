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

	"ludwig-go/internal/highlight"
	nc "ludwig-go/internal/ncurses"
)

// Constants
const (
	BS  = 8
	NL  = 10
	CR  = 13
	SPC = 32
	DEL = 127

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
	terminators  map[int]bool
	vduSetup     bool
	inInsertMode bool
	stdscr       *nc.Window
	refreshDelay int
)

func init() {
	terminators = make(map[int]bool)
	if v, err := strconv.Atoi(os.Getenv("LUD_REFRESH_DELAY")); err == nil && v > 0 {
		refreshDelay = v
	}
}

func isControlChar(key int) bool {
	return (key >= 0 && key <= 31) || key == 127
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
func VduDisplayStr(str string, clearToEol bool) {
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

	if !hitMargin && clearToEol {
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
	for k := range 32 {
		terminators[k] = true
	}
	terminators[127] = true
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

	if rawKey == nc.KEY_RESIZE {
		TtWinChanged = true
	}
	nc.CursSet(0)
	return massageKey(int(rawKey))
}

// VduGetInput gets a line of input from the user with a prompt
func VduGetInput(prompt string, get **StrObject, getLen int, outlen *int) {
	VduBold()
	VduDisplayStr(prompt, true)
	VduNormal()

	// Fill get with spaces
	*get = NewBlankStrObject(MaxStrLen)

	_, curX := stdscr.CursorYX()
	_, maxX := stdscr.MaxYX()
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
			if key < 0 || key > OrdMaxChar || isControlChar(key) {
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
	_, maxX := stdscr.MaxYX()
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
	// Control character names: "CONTROL-X" with exceptions for special keys
	ctrlOverrides := map[int]string{9: "BACKSPACE", 10: "TAB", 11: "LINE-FEED", 14: "RETURN"}
	for i := 1; i <= 32; i++ {
		if name, ok := ctrlOverrides[i]; ok {
			kl[i].KeyName = name
		} else {
			kl[i].KeyName = "CONTROL-" + string(rune('@'+i-1))
		}
	}
	kl[33].KeyCode = DEL
	kl[33].KeyName = "DELETE"

	// Initialize ncurses key names
	for i := MinCursesKey; i <= MaxCursesKey; i++ {
		kl[NumControlChars+1+i-MinCursesKey].KeyCode = massageKey(i)
	}

	for i, name := range []string{"BREAK", "DOWN-ARROW", "UP-ARROW", "LEFT-ARROW", "RIGHT-ARROW", "HOME", "BACKSPACE"} {
		kl[NumControlChars+1+i].KeyName = name
	}
	// FUNCTION-0 through FUNCTION-12 (indices NumControlChars+8 to NumControlChars+20)
	for n := 0; n <= 12; n++ {
		kl[NumControlChars+8+n].KeyName = "FUNCTION-" + strconv.Itoa(n)
	}
	// SHIFT-FUNCTION-1 through SHIFT-FUNCTION-12 (indices NumControlChars+21 to NumControlChars+32)
	for n := 1; n <= 12; n++ {
		kl[NumControlChars+20+n].KeyName = "SHIFT-FUNCTION-" + strconv.Itoa(n)
	}
	// FUNCTION-25 through FUNCTION-63 (indices NumControlChars+33 to NumControlChars+71)
	for n := 25; n <= 63; n++ {
		kl[NumControlChars+8+n].KeyName = "FUNCTION-" + strconv.Itoa(n)
	}
	// Terminal-specific key names (indices NumControlChars+72 onward)
	for i, name := range []string{
		"DELETE-LINE", "INSERT-LINE", "DELETE-CHAR", "INSERT-CHAR",
		"EIC", "CLEAR", "CLEAR-EOS", "CLEAR-EOL",
		"SCROLL-FORWARD", "SCROLL-REVERSE", "PAGE-DOWN", "PAGE-UP",
		"SET-TAB", "CLEAR-TAB", "CLEAR-ALL-TABS", "SEND",
		"SOFT-RESET", "RESET", "PRINT", "LOWER-LEFT",
		"KEY-A1", "KEY-A3", "KEY-B2", "KEY-C1", "KEY-C3",
		"BACK-TAB", "BEGIN", "CANCEL", "CLOSE", "COMMAND",
		"COPY", "CREATE", "END", "EXIT", "FIND", "HELP",
		"MARK", "MESSAGE", "MOVE", "NEXT", "OPEN", "OPTIONS",
		"PREVIOUS", "REDO", "REFERENCE", "REFRESH", "REPLACE",
		"RESTART", "RESUME", "SAVE",
		"SHIFT-BEGIN", "SHIFT-CANCEL", "SHIFT-COMMAND", "SHIFT-COPY",
		"SHIFT-CREATE", "SHIFT-DELETE-CHAR", "SHIFT-DELETE-LINE",
		"SELECT", "SEND", "SHIFT-CLEAR-EOL", "SHIFT-EXIT",
		"SHIFT-FIND", "SHIFT-HELP", "SHIFT-HOME", "SHIFT-INSERT-CHAR",
		"SHIFT-LEFT", "SHIFT-MESSAGE", "SHIFT-MOVE", "SHIFT-NEXT",
		"SHIFT-OPTIONS", "SHIFT-PREVIOUS", "SHIFT-PRINT", "SHIFT-REDO",
		"SHIFT-REPLACE", "SHIFT-RIGHT", "SHIFT-RESUME", "SHIFT-SAVE",
		"SHIFT-SUSPEND", "SHIFT-UNDO", "SUSPEND", "UNDO",
		"MOUSE", "WINDOW-RESIZE-EVENT", "SOME-OTHER-EVENT",
	} {
		kl[NumControlChars+72+i].KeyName = name
	}

	// Clear key introducers
	for i := range keyIntroducers {
		keyIntroducers[i] = false
	}
	*keyNameList = make([]KeyNameRecord, *nrKeyNames)
	copy(*keyNameList, kl)
}

// VduInit initializes the VDU system
func VduInit(terminalInfo *TerminalInfoType) (map[string]int, bool) {
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

			colors := highlight.ColorInit(FileData.Highlighting)
			VduClearScr()
			return colors, true
		}
	}
	return nil, false
}

// VduFree cleans up the VDU system
func VduFree() {
	if vduSetup {
		vduSetup = false
		VduScrollUp(1)
		maxY, _ := stdscr.MaxYX()
		VduMoveCurs(1, maxY)
		VduFlush()
		nc.EndWin()
		highlight.ColorReset()
	}
}

// VduGetNewDimensions gets the new screen dimensions after resize
func VduGetNewDimensions(newX *int, newY *int) {
	nc.EndWin()
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
