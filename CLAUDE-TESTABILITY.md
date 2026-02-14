# Improving Encapsulation and Testability

## Current State

The codebase is a faithful Pascal-to-Go port: ~25k lines in a single `internal/ludwig`
package with **44 exported package-level globals** (in `var.go`), **~308 exported free
functions**, and only ~31 methods (mostly on `StrObject` and `FrameOptions`). Everything
is tightly coupled through shared mutable global state.

Key metrics:

- `CurrentFrame` is referenced **~1,057 times** across 25 files
- Screen functions (`ScreenMessage`, `ScreenRedraw`, etc.) are called from **19 files**
- `LudwigMode` is checked in **11 files**
- `ScrFrame`/`ScrTopLine`/`ScrBotLine` etc. are referenced in **11 files** (~251 occurrences)
- `goto` is used in **14 files** for error handling
- Unit test coverage is ~2.8%; **348 system tests** provide the primary safety net

## Problems

### 1. Global state makes unit testing nearly impossible

Every function implicitly depends on globals like `CurrentFrame`, `ScrFrame`, `LudwigMode`,
`EditMode`, etc. Tests must carefully set up (and tear down) global state, and tests cannot
run in parallel. The existing test helpers (e.g., `createTestFrame()`) work around this, but
many functions also read globals that aren't set by the helpers.

Example from `exec.go:57`:

```go
// FIXME: Should this be "frame" rather than "CurrentFrame"
if !LineFromNumber(CurrentFrame, lineNr+count-1, lastLine) {
```

The function takes `frame` as a parameter but still reaches for the `CurrentFrame` global.

### 2. No separation between core logic and I/O

Text manipulation functions directly call screen update functions. For example, `TextInsert`
in `text.go` calls `ScreenDrawLine`, `ScreenRedraw`, etc. This means:

- You can't test text operations without the screen subsystem being initialized
- You can't reuse text logic in a headless context without stubbing the entire screen layer

### 3. Single flat package

All ~50 source files share one namespace. Internal helpers are indistinguishable from the
public API. The dot-import in `cmd/ludwig/main.go` (`import . "ludwig-go/internal/ludwig"`)
pulls every exported symbol into scope, confirming there is no encapsulation boundary.

### 4. Pascal idioms that hurt Go testability

- **Output parameters via `**T`**: Functions like
  `LinesCreate(count int, first **LineHdrObject, last **LineHdrObject) bool` use
  double-pointer output params instead of returning values. This makes them harder to
  compose and test.
- **`goto` for error handling**: `exec.go` and 13 other files use `goto` labels instead of
  early returns or error values.
- **Boolean success returns**: Nearly all functions return `bool` for success/failure instead
  of `error`, losing diagnostic information.

## Recommended Improvements

An incremental approach, where each phase is independently valuable and verifiable against
the system tests.

### Phase 1: Introduce an EditorState struct

**Goal**: Eliminate package-level mutable globals so tests can create isolated editor
instances.

**Approach**: Consolidate the globals from `var.go` into a struct:

```go
type EditorState struct {
    CurrentFrame *FrameObject
    FrameOops    *FrameObject
    FrameCmd     *FrameObject
    FrameHeap    *FrameObject

    EditMode     ModeType
    PreviousMode ModeType
    LudwigMode   LudwigModeType

    FirstSpan    *SpanObject
    Files        [MaxFiles + 1]*FileObject
    FilesFrames  [MaxFiles + 1]*FrameObject
    FgiFile      int
    FgoFile      int

    ScrFrame     *FrameObject
    ScrTopLine   *LineHdrObject
    ScrBotLine   *LineHdrObject
    ScrMsgRow    int
    ScrNeedsFix  bool

    LudwigAborted bool
    ExitAbort     bool
    Hangup        bool

    // Compiler state
    CompilerCode [MaxCode + 1]CodeObject
    CodeList     *CodeHeader
    CodeTop      int

    // Command tables
    Lookup       [OrdMaxChar + MaxSpecialKeys + 1]CommandObject
    LookupExp    [ExpandLim + 1]LookupExpType
    CmdAttrib    [CmdNoSuch + 1]CmdAttribRec
    ExecLevel    int

    // ... remaining globals
}
```

**Migration path**:

1. Create the `EditorState` struct with all fields from `var.go`
2. Create a package-level `var defaultState = &EditorState{}`
3. Convert existing global references to use `defaultState` (mechanical find/replace)
4. Gradually convert free functions to methods on `*EditorState`, or add `*EditorState`
   as a first parameter, file by file
5. Verify with system tests after each file conversion

**Impact**: Tests can create independent `EditorState` instances. No more global setup/teardown
concerns. Tests can run in parallel.

### Phase 2: Extract a Screen interface

**Goal**: Decouple text/command logic from terminal I/O so core operations can be tested
without ncurses.

**Approach**: Define an interface for screen operations:

```go
type Screen interface {
    Message(msg string)
    DrawLine(line *LineHdrObject, row, col int)
    Redraw(startRow, finishRow int)
    FreeBottomLine()
    Update()
    MoveCurs(col, row int)
    // ... remaining screen operations (~10-15 methods total)
}
```

Add a `Screen` field to `EditorState`. The real implementation wraps the current ncurses code
(`screen.go` + `vdu.go`). For tests, provide a no-op or recording implementation:

```go
type NullScreen struct{}
func (NullScreen) Message(msg string)                      {}
func (NullScreen) DrawLine(line *LineHdrObject, row, col int) {}
// ...
```

**Migration path**:

1. Audit all `Screen*` and `Vdu*` calls from non-screen files to determine the interface
   surface
2. Define the `Screen` interface
3. Wrap existing screen/vdu code in a concrete `TerminalScreen` struct
4. Add `Screen` to `EditorState` and convert callers to use it
5. Create `NullScreen` for tests

**Impact**: `text.go`, `frame.go`, `exec.go`, and other core files become testable without
terminal initialization.

### Phase 3: Return values instead of output parameters

**Goal**: Replace Pascal-style double-pointer output parameters with idiomatic Go returns.

**Examples**:

```go
// Before:
func LinesCreate(count int, first **LineHdrObject, last **LineHdrObject) bool

// After:
func LinesCreate(count int) (first, last *LineHdrObject, ok bool)
```

```go
// Before:
func LineToNumber(line *LineHdrObject, lineNr *int) bool

// After:
func LineToNumber(line *LineHdrObject) (lineNr int, ok bool)
```

**Migration path**: Convert function-by-function. Each conversion is a small, verifiable
change. Start with the most-called functions to get the biggest readability improvement.

**Impact**: More idiomatic, composable, and testable code. Eliminates a common source of
confusion for Go developers.

### Phase 4: Replace `goto` with structured control flow

**Goal**: Remove `goto` statements from the 14 files that use them, improving readability
and making functions easier to extract and test.

**Example from `exec.go`**:

```go
// Before:
func ExecComputeLineRange(...) bool {
    result := false
    // ... many cases with goto l99
l99:
    return result
}

// After:
func ExecComputeLineRange(...) bool {
    switch rept {
    case LeadParamNone, LeadParamPlus, LeadParamPInt:
        // ... returns true/false directly
    case LeadParamMinus, LeadParamNInt:
        // ... returns true/false directly
    }
    return false
}
```

**Files affected**: `exec.go`, `execimmed.go`, `frame.go`, `fyle.go`, `screen.go`,
`swap.go`, `quit.go`, `code.go`, `charcmd.go`, `caseditto.go`, `newword.go`,
`nextbridge.go`, `eqsgetrep.go`, `opsys.go`.

**Migration path**: Convert one function at a time, verify with system tests.

### Phase 5: Split the package (longer term)

**Goal**: Break the monolithic `internal/ludwig` package into focused sub-packages with
clear dependency boundaries.

**Candidate packages**:

| Package | Contents | Dependencies |
| ------- | -------- | ------------ |
| `ludwig/text` | `StrObject`, text/line/group operations | None (pure data) |
| `ludwig/pattern` | NFA/DFA pattern matching (`patparse.go`, `dfa.go`, `nfa.go`) | `ludwig/text` only |
| `ludwig/screen` | `Screen` interface + `TerminalScreen` impl | `ludwig/text` |
| `ludwig/editor` | `EditorState`, command execution, file I/O | All above |

The pattern matching subsystem is the easiest first candidate — it has minimal coupling to
global state and already has decent test coverage.

**Migration path**:

1. Extract `ludwig/text` (StrObject and line operations) first
2. Extract `ludwig/pattern` (already nearly self-contained)
3. Extract `ludwig/screen` (after Phase 2 interface exists)
4. Everything remaining becomes `ludwig/editor`

## Prioritization

**Start with Phase 1 + Phase 2 together** because:

- They unlock testability for the entire codebase
- The 348 system tests provide a safety net for the refactoring
- Each converted function immediately becomes independently testable
- They are prerequisites for Phase 5

**Phase 3 and Phase 4** can be done opportunistically — whenever a function is being
modified for other reasons, convert its signature and control flow at the same time.

**Phase 5** becomes practical once Phases 1 and 2 are complete and provides the final
encapsulation boundary.
