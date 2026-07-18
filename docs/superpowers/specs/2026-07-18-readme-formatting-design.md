# README Formatting Design

## Goal

Make every code or output example in `README.md` use a language-specific GitHub Markdown fence, and display the fork-maintenance notice as a GitHub information-style alert.

## Changes

- Wrap the existing two-sentence fork notice in a GitHub `[!NOTE]` alert.
- Use `bash` for installation and contributor command blocks.
- Use `text` for plain error and stacktrace output.
- Use `go` for constants and all Go examples.
- Replace the existing HTML `<pre>` and `<b>` markup with fenced `go` blocks. The highlighted calls become ordinary Go code because fenced syntax highlighting and inline HTML bolding cannot be combined reliably.
- Preserve all prose, example code, commands, and ordering apart from the Markdown container changes.

## Acceptance Checks

1. Every triple-backtick opening fence in `README.md` has `bash`, `text`, or `go`.
2. `README.md` contains no `<pre>`, `</pre>`, `<b>`, or `</b>` tags.
3. The fork notice begins with `> [!NOTE]` and both sentences remain unchanged.
4. Markdown fences are balanced.
