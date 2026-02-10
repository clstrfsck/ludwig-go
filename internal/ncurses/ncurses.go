// Package ncurses provides a minimal cgo wrapper for the ncurses library
package ncurses

/*
#cgo LDFLAGS: -lncurses
#include <ncurses.h>
#include <stdlib.h>

// Helper function to get KEY_MAX
static int get_key_max() {
	return KEY_MAX;
}

// Helper functions for getyx and getmaxyx macros
static void get_yx(WINDOW *win, int *y, int *x) {
	getyx(win, *y, *x);
}

static void get_maxyx(WINDOW *win, int *y, int *x) {
	getmaxyx(win, *y, *x);
}
*/
import "C"
import (
	"unsafe"
)

// Window represents an ncurses window
type Window struct {
	win *C.WINDOW
}

// Key represents a key code
type Key int

// Char represents a character
type Char rune

// Attribute constants
const (
	A_BOLD    int = C.A_BOLD
	A_REVERSE int = C.A_REVERSE
)

// Key constants
const (
	KEY_BACKSPACE = C.KEY_BACKSPACE
	KEY_RESIZE    = C.KEY_RESIZE
	KEY_DOWN      = C.KEY_DOWN
	KEY_UP        = C.KEY_UP
	KEY_LEFT      = C.KEY_LEFT
	KEY_RIGHT     = C.KEY_RIGHT
	KEY_HOME      = C.KEY_HOME
	KEY_BREAK     = C.KEY_BREAK
	KEY_F0        = C.KEY_F0
	KEY_DC        = C.KEY_DC
	KEY_IC        = C.KEY_IC
	KEY_NPAGE     = C.KEY_NPAGE
	KEY_PPAGE     = C.KEY_PPAGE
	KEY_END       = C.KEY_END
	KEY_BTAB      = C.KEY_BTAB
)

var (
	// KEY_MAX is the maximum key value
	KEY_MAX int
	// stdscr is the standard screen window
	stdscr *Window
)

// Init initializes ncurses and returns the standard screen window
func Init() (*Window, error) {
	cwin := C.initscr()
	if cwin == nil {
		return nil, ErrInitFailed
	}
	KEY_MAX = int(C.get_key_max())
	stdscr = &Window{win: cwin}
	return stdscr, nil
}

// End cleans up ncurses
func End() {
	C.endwin()
}

// Raw sets raw mode
func Raw(enable bool) {
	if enable {
		C.raw()
	} else {
		C.noraw()
	}
}

// Echo sets echo mode
func Echo(enable bool) {
	if enable {
		C.echo()
	} else {
		C.noecho()
	}
}

// NewLines sets newline translation mode
func NewLines(enable bool) {
	if enable {
		C.nl()
	} else {
		C.nonl()
	}
}

// Flash flashes the screen
func Flash() {
	C.flash()
}

// CursSet sets the cursor visibility
func CursSet(visibility int) {
	C.curs_set(C.int(visibility))
}

// UnGetChar pushes a character back to the input queue
func UnGetChar(ch Char) {
	C.ungetch(C.int(ch))
}

// ErrInitFailed is returned when ncurses initialization fails
type InitError struct{}

func (e *InitError) Error() string {
	return "ncurses initialization failed"
}

var ErrInitFailed = &InitError{}

// Window methods

// Move moves the cursor to the specified position
func (w *Window) Move(y, x int) {
	C.wmove(w.win, C.int(y), C.int(x))
}

// Refresh refreshes the window
func (w *Window) Refresh() {
	C.wrefresh(w.win)
}

// ClearToEOL clears from cursor to end of line
func (w *Window) ClearToEOL() {
	C.wclrtoeol(w.win)
}

// Print prints a string
func (w *Window) Print(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.waddstr(w.win, cstr)
}

// AddChar adds a character at the cursor position
func (w *Window) AddChar(ch Char) {
	C.waddch(w.win, C.chtype(ch))
}

// Clear clears the entire window
func (w *Window) Clear() {
	C.wclear(w.win)
}

// ClearToBottom clears from cursor to bottom of window
func (w *Window) ClearToBottom() {
	C.wclrtobot(w.win)
}

// ScrollOk sets scrolling mode
func (w *Window) ScrollOk(enable bool) {
	if enable {
		C.scrollok(w.win, C.bool(true))
	} else {
		C.scrollok(w.win, C.bool(false))
	}
}

// Scroll scrolls the window by n lines
func (w *Window) Scroll(n int) {
	C.wscrl(w.win, C.int(n))
}

// InsDelLines inserts or deletes lines (negative n deletes)
func (w *Window) InsDelLines(n int) {
	C.winsdelln(w.win, C.int(n))
}

// InsChar inserts a character at cursor position
func (w *Window) InsChar(ch Char) {
	C.winsch(w.win, C.chtype(ch))
}

// DelChar deletes character at cursor position
func (w *Window) DelChar() {
	C.wdelch(w.win)
}

// CursorYX returns the current cursor position
func (w *Window) CursorYX() (y, x int) {
	var cy, cx C.int
	C.get_yx(w.win, &cy, &cx)
	return int(cy), int(cx)
}

// MaxYX returns the maximum window size
func (w *Window) MaxYX() (y, x int) {
	var my, mx C.int
	C.get_maxyx(w.win, &my, &mx)
	return int(my), int(mx)
}

// GetChar gets a character from the window
func (w *Window) GetChar() Key {
	ch := C.wgetch(w.win)
	return Key(ch)
}

// Keypad enables or disables keypad mode
func (w *Window) Keypad(enable bool) {
	if enable {
		C.keypad(w.win, C.bool(true))
	} else {
		C.keypad(w.win, C.bool(false))
	}
}

// AttrOn turns on the specified attributes
func (w *Window) AttrOn(attr int) {
	C.wattron(w.win, C.int(attr))
}

// AttrOff turns off the specified attributes
func (w *Window) AttrOff(attr int) {
	C.wattroff(w.win, C.int(attr))
}

// IntrFlush controls interrupt flush
func (w *Window) IntrFlush(enable bool) {
	if enable {
		C.intrflush(w.win, C.bool(true))
	} else {
		C.intrflush(w.win, C.bool(false))
	}
}

// Idlok enables or disables use of hardware insert/delete line feature
func (w *Window) Idlok(enable bool) {
	if enable {
		C.idlok(w.win, C.bool(true))
	} else {
		C.idlok(w.win, C.bool(false))
	}
}

// Idcok enables or disables use of hardware insert/delete character feature
func (w *Window) Idcok(enable bool) {
	if enable {
		C.idcok(w.win, C.bool(true))
	} else {
		C.idcok(w.win, C.bool(false))
	}
}
