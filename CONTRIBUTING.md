# Contributing to kyaraben

This document describes technical preferences and conventions for contributing
to kyaraben.

## Languages and typing

We use Go for the backend and CLI, chosen for its simplicity and
straightforward tooling. The UI is TypeScript. Both languages offer static
typing, which we rely on heavily.

When defining protocols or data interchange formats, strict typing is
essential. Go types in `internal/daemon/types.go` serve as the source of truth
for the daemon protocol. TypeScript types should be generated from Go types to
ensure the contract between components is explicit and enforced at compile
time. JSON schemas may be generated from Go for documentation or runtime
validation.

## Testing

We follow Martin Fowler's distinction between fakes and mocks. Fakes are
working implementations with shortcuts, such as an in-memory store instead of a
real database. Mocks verify that specific methods were called with specific
arguments, which couples tests to implementation details. Prefer fakes.

Test harnesses matter more than individual test cases. A well-designed harness
that can spin up isolated environments, run commands, and assert on outcomes
makes writing new tests trivial. Invest in the harness.

Unit tests cover pure logic. Integration tests use fakes for external
dependencies like the Nix client. End-to-end tests invoke the real system,
including actual Nix builds. E2E tests are slower but validate the full flow.

For UI E2E tests, follow the Playwright best practices at
https://playwright.dev/docs/best-practices. Prefer user-facing selectors like
`getByRole`, `getByLabel`, and `getByText` over CSS selectors or test IDs. Test
what users see and interact with, not implementation details.

## Simplicity

Start with the simplest solution that works. Do not build abstractions until
you need them. If a feature is not required for the current milestone, do not
design for it yet.

When a dependency introduces breaking changes, treat the new version as a new
entity rather than adding complex version handling. This keeps the codebase
simple at the cost of some duplication, which is an acceptable trade-off.

Explicit configuration is better than implicit defaults. When something can be
configured, make the user state their choice rather than guessing.

## Domain modeling

Use domain-driven design principles. Name things precisely using the language
of the domain. When writing documentation, use backticks when referring to
domain concepts to distinguish them from ordinary English words.

Keep the domain model clean. Implementation details such as serialization
formats or database schemas should not leak into the model.

## Adding systems and emulators

Systems and emulators are defined in separate packages that implement
interfaces from `internal/model/definitions.go`. This keeps each definition
self-contained and provides compile-time safety for missing methods.

To add a new system (e.g., N64):

1. Add the `SystemID` constant to `internal/model/system.go`
2. Create `internal/systems/n64/n64.go` implementing `model.SystemDefinition`
3. Add the system to `internal/registry/all.go`

To add a new emulator (e.g., mupen64plus):

1. Add the `EmulatorID` constant to `internal/model/emulator.go`
2. Create `internal/emulators/mupen64plus/mupen64plus.go` implementing
   `model.EmulatorDefinition` (includes both the emulator metadata and the
   config generator)
3. Add the emulator to `internal/registry/all.go`

The `TestAllDefinitions` test in `internal/registry/registry_test.go` validates
that all definitions have required fields and that cross-references are valid
(default emulators exist and support their systems, emulators reference
existing systems).

## Dependency management

Pass dependencies explicitly. There should be no hidden instantiation deep in
the call stack. Expensive instantiations happen at the composition root,
typically main.go, and dependencies are threaded down through constructors.

Define dependencies as interfaces where substitution is needed, following the
Dependency Inversion principle. This makes testing straightforward: swap real
implementations for fakes at construction time.

## Logging

The Go daemon writes to `~/.local/state/kyaraben/kyaraben.log`. Use the logging
package (`internal/logging`) for all Go log output. The daemon owns this log
file; other processes should not write to it directly to avoid race conditions.

Create a component-scoped logger at package level:

```go
var log = logging.New("nix")

func DoSomething() {
    log.Info("starting operation")
    log.Debug("details: %v", details)
    log.Error("operation failed: %v", err)
}
```

The component name appears in log entries: `[INFO] [nix] starting operation`.
This makes it clear which part of the system produced each log entry.

Use appropriate log levels:

- `Debug`: detailed information useful for debugging, not needed in normal
  operation
- `Info`: significant events in normal operation (starting builds, completing
  operations)
- `Error`: error conditions that should be investigated

Keep log messages concise and actionable. Include relevant context (file paths,
operation names, error details) but avoid verbose output that makes logs hard
to read.

The UI (Electron) writes to two destinations:

1. Through the daemon via a `log` protocol command, which the daemon writes to
   the unified log file. This provides a single log file with all system
   activity for holistic debugging.
2. Directly to its own file at `~/.local/state/kyaraben/kyaraben-ui.log`. This
   captures UI-specific logs including protocol-level details (what was sent,
   what was received) for debugging UI issues in isolation.

This dual-sink approach enables:

- Holistic debugging: review `kyaraben.log` to see all daemon activity plus
  UI-originating events in chronological order
- UI-specific debugging: review `kyaraben-ui.log` to see UI operations,
  protocol messages, and UI-side errors in isolation

Log files:

- `~/.local/state/kyaraben/kyaraben.log`: unified log (daemon writes, UI
  writes through daemon)
- `~/.local/state/kyaraben/kyaraben-ui.log`: UI-only log (UI writes directly)

## Code style

Make code self-evident. Do not write comments explaining what code does; if the
code needs explanation, rewrite it to be clearer.

Use sentence-case in headings. Do not use bold text in documentation. Avoid
em-dashes for punctuation.

## Commit messages

Never commit without explicit user approval of the message. Draft the message,
present it, and iterate until approved.

The format is a brief actionable title in imperative mood, followed by a body
explaining what changed and why, followed by a test plan:

```
Brief, actionable description

What changed and why. Use paragraphs, lists, or both as appropriate.

## Test plan

Reproducible verification steps with brief descriptions of what to check.
```

Use imperative mood in titles: "Add", "Fix", "Remove", not "Added", "Fixes",
"Removed". No trailing periods on list items.

Use backticks for code references and fenced code blocks with language tags. Do
not use bold text. Links must be internet-accessible; never use local paths.

Never include LLM attribution, Co-Authored-By for AI, or any indication of AI
involvement. Avoid value judgements like "improve", "better", "cleaner",
"robust", or "elegant". Avoid vague terms like "various", "several
improvements", or "minor fixes". Do not make unsubstantiated claims about
performance or behavior; if claiming an improvement, include proof such as
benchmarks or observability data.

Bad: "Improve error handling for better reliability" Good: "Add retry logic for
transient network failures"

## Process

Discuss approaches before committing to them. When facing a design decision,
write down the options and trade-offs before picking one. Document decisions
and rationale so future contributors understand why things are the way they
are.

Keep documentation as living documents. When the code changes, update the
relevant docs. Stale documentation is worse than no documentation.
