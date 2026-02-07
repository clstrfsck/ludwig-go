# Ncurses CGO Wrapper

This package provides a minimal CGO wrapper for the ncurses library, replacing the previous dependency on goncurses.

## Overview

The wrapper provides direct access to ncurses functions through CGO, eliminating the need for third-party dependencies while maintaining compatibility with the existing VDU interface.

## Key Features

- Direct ncurses bindings using CGO
- Minimal interface focused on Ludwig's requirements
- Compatible with existing VDU API
- No external Go dependencies (only ncurses C library required)

## Functions Provided

### Initialization and Cleanup
- `Init()` - Initialize ncurses and return the standard screen window
- `End()` - Clean up and close ncurses

### Terminal Configuration
- `Raw(bool)` - Enable/disable raw mode
- `Echo(bool)` - Enable/disable character echoing
- `NewLines(bool)` - Enable/disable newline translation
- `Flash()` - Flash the screen (visual bell)

### Input
- `UnGetChar(Char)` - Push a character back to input queue
- `Window.GetChar()` - Get a character from the window

### Window Methods
- `Move(y, x int)` - Move cursor to position
- `Refresh()` - Refresh the window display
- `ClearToEOL()` - Clear from cursor to end of line
- `ClearToBottom()` - Clear from cursor to bottom of window
- `Clear()` - Clear entire window
- `Print(string)` - Print a string
- `AddChar(Char)` - Add a character at cursor
- `ScrollOk(bool)` - Enable/disable scrolling
- `Scroll(int)` - Scroll window by n lines
- `InsDelLines(int)` - Insert or delete lines
- `InsChar(Char)` - Insert character at cursor
- `DelChar()` - Delete character at cursor
- `CursorYX()` - Get current cursor position
- `MaxYX()` - Get window dimensions
- `Keypad(bool)` - Enable/disable keypad mode
- `AttrOn(int)` - Enable text attributes
- `AttrOff(int)` - Disable text attributes
- `IntrFlush(bool)` - Control interrupt flush
- `Idlok(bool)` - Enable/disable hardware insert/delete line
- `Idcok(bool)` - Enable/disable hardware insert/delete character

## Key Constants

### Attributes
- `A_BOLD` - Bold text attribute
- `A_REVERSE` - Reverse video attribute

### Special Keys
- `KEY_BACKSPACE` - Backspace key
- `KEY_RESIZE` - Terminal resize event
- `KEY_DOWN`, `KEY_UP`, `KEY_LEFT`, `KEY_RIGHT` - Arrow keys
- `KEY_HOME`, `KEY_END` - Home and End keys
- `KEY_F0` - Function key F0 (F1-F63 computed from this)
- `KEY_DC`, `KEY_IC` - Delete/Insert character
- `KEY_NPAGE`, `KEY_PPAGE` - Page Down/Up
- `KEY_BTAB` - Back tab

## Implementation Details

### Macro Wrappers

Some ncurses functions are implemented as macros in C, which cannot be directly called from CGO. For these, we provide helper functions in the C preamble:

- `get_key_max()` - Wrapper for `KEY_MAX` macro
- `get_yx()` - Wrapper for `getyx()` macro
- `get_maxyx()` - Wrapper for `getmaxyx()` macro

### CGO Configuration

The package uses the following CGO directives:
```go
#cgo LDFLAGS: -lncurses
```

This links against the ncurses library, which must be installed on the system.

## System Requirements

- ncurses development library (libncurses-dev or ncurses-devel)
- C compiler (gcc, clang)
- CGO enabled (default for Go)

## Building

The package builds automatically with the rest of the ludwig-go project:

```bash
go build ./cmd/ludwig
```

Or using the task runner:

```bash
task build
```

## Migration from goncurses

The wrapper provides a compatible API with the previous goncurses implementation:

| goncurses | ncurses wrapper |
|-----------|----------------|
| `gc.Init()` | `nc.Init()` |
| `gc.End()` | `nc.End()` |
| `gc.Flash()` | `nc.Flash()` |
| `gc.UnGetChar()` | `nc.UnGetChar()` |
| `gc.Char` | `nc.Char` |
| `gc.Key` | `nc.Key` |
| `gc.KEY_*` | `nc.KEY_*` |
| `gc.A_BOLD` | `nc.A_BOLD` |
| Window methods work the same |

## Notes

- The wrapper intentionally provides only the functions needed by Ludwig
- Error handling follows Go conventions (error returns where applicable)
- The `Window` type wraps the C `WINDOW*` pointer
- Character types use Go's `rune` type for proper Unicode support
