# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Ludwig is a text editor originally developed at the University of Adelaide in Pascal, ported to Go. It supports interactive screen editing, hardcopy terminal mode, and batch mode processing. This is a Go port; the original Pascal source is at [cjbarter/ludwig](https://github.com/cjbarter/ludwig).

## Build & Test Commands

Uses `task` (Taskfile.yml) as the build tool, with Go as a fallback:

```sh
task build            # Debug binary (bin/ludwig-debug) with debugger symbols
task build-release    # Release binary (bin/ludwig) stripped
task test             # Unit tests: go test ./internal/ludwig
task coverage         # Coverage report (opens HTML)
task system-test      # Python/pytest system tests (requires system-test/ clone)
task check            # All: build, build-release, build-help, test, system-test
```

Run a single test:
```sh
go test -run TestFunctionName ./internal/ludwig
```

System tests require cloning [ludwig-system-test](https://github.com/clstrfsck/ludwig-system-test) into `system-test/` and having Python, pytest, and pexpect installed. Use `LUDWIG_EXE` env var to point to a custom binary path.

## Architecture

All editor logic lives in a single package: `internal/ludwig/`. Entry points are in `cmd/ludwig/` (editor) and `cmd/ludwighlpbld/` (help file indexer).

### Core Data Structure Hierarchy

**Frame → Group → Line** is the central hierarchy:

- **FrameObject** (`frame.go`): An editor window/buffer. Contains line groups, cursor (Dot), named marks (a-z), tab stops, margins, and frame options (auto-indent, auto-wrap, etc.).
- **GroupObject** (`group.go`): A contiguous range of lines within a frame. Groups are doubly-linked within a frame and track line numbering.
- **LineHdrObject** (`line.go`): A single text line. Doubly-linked within its group. Contains a `StrObject` for text content and a list of marks attached to the line.
- **StrObject** (`strobject.go`): Variable-size byte array with **1-based indexing**. Allocates in quantized chunks (multiples of 10). This is the fundamental text storage unit.
- **MarkObject** (`mark.go`): A position (line + column, both 1-based). Used for cursors, named marks, and region boundaries. Marks auto-adjust when text is inserted/deleted.
- **SpanObject** (`span.go`): A named text region bounded by two marks. Frames are associated with spans. Spans can hold compiled command code.

### Key Subsystems

- **Text operations** (`text.go`): Insert, delete, move/copy text. Core editing primitives.
- **Line operations** (`line.go`): Insert, delete, split, merge lines. Manages group boundaries.
- **Pattern matching** (`patparse.go`, `dfa.go`, `nfa.go`): NFA/DFA-based regex-like pattern matching for search/replace.
- **Command execution** (`exec.go`, `code.go`): Command interpreter. Commands are an enum (`Commands` type) with 100+ types. Two command sets: "old" and "new" key bindings.
- **File I/O** (`filesys.go`): Reading/writing files with backup support.
- **Screen management** (`screen.go`, `vdu*.go`): Terminal rendering via ANSI escape sequences.

### Design Characteristics

- **Global state**: Editor state uses package-level globals (`CurrentFrame`, `ScrFrame`, `FirstSpan`, etc.).
- **Doubly-linked lists**: Used throughout for frames, groups, lines, marks, and spans (FLink/BLink pattern).
- **1-based indexing**: Columns, line positions, and StrObject indices are all 1-based, matching the original Pascal convention.
- **Trailing parameters (TParObject)**: Commands accept optional string parameters via a linked list structure with sequential (Next) and continuation (Con) linkage.  `first.Next` is the second trailing parameter, and `first.Con` is the second line of the first parameter.

## Testing

Unit tests use `github.com/stretchr/testify/assert` with table-driven patterns. Test helpers like `createTestFrame` and `createTestLine` set up editor state for isolated testing. Unit test coverage is low (~2.8%); system tests provide the primary safety net (348 passing).
