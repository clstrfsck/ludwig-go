# Rust or Golang

I asked Claude Code whether it would be better to improve this golang port, or start work on a Rust re-implementation.  Here's how that went.

## The question

> It's my intention to eventually write a Rust implementation of this editor.  The Rust implementation would be a complete rewrite though, rather than a conversion from the existing source code.
>
> In your view, do you think it is worth investing the time to improve the golang codebase, or would it be preferable to start work on the Rust implementation sooner?

## Response

> ⏺ Honestly, I'd skip most of the Go refactoring and start the Rust rewrite sooner. Here's my reasoning:
>
> - **The refactoring effort has a limited lifespan**. Phases 1-2 alone would touch thousands of call sites across 25+ files. That's weeks of careful, mechanical work that gets discarded when the Rust version is ready.
>
> - **The Go code is already a useful reference as-is**. For a complete rewrite, you need to understand the algorithms and semantics, not have clean    Go code to port from. The current code — even with its globals and Pascal idioms — is perfectly readable for understanding how the Frame→Group→Line hierarchy works, how the NFA/DFA pattern matcher is built, how marks auto-adjust, etc. You've already been through it once for the Pascal→Go port, so the domain knowledge is there.
>
> - **The system tests are the real asset, and they're language-agnostic**. The 348 pytest/pexpect tests exercise the editor's external behavior through terminal interaction. They'll work against a Rust binary just as well — just point LUDWIG_EXE at it. That's your safety net for the rewrite, not clean Go internals.
>
> - **The Rust rewrite will have better structure from day one**. Rust's ownership model and module system will naturally push you toward the encapsulation that's hard to retrofit into the Go code. The problems in TESTABILITY.md (global mutable state, no I/O separation, flat namespace) are things Rust essentially forces you to solve up front.
>
> - **The one thing I'd consider doing first** is any targeted work that makes the system tests more comprehensive or easier to run against a different binary. That investment carries over directly to the Rust implementation. If there are behavioral edge cases you've discovered in the Go port that aren't covered by system tests, capturing those as new test cases would be more valuable than refactoring the Go internals.
