# Contributing to kyaraben

This document describes technical preferences and conventions for contributing
to kyaraben.

## Languages and typing

We use Go for the backend and CLI, chosen for its simplicity and
straightforward tooling. The UI is TypeScript. Both languages offer static
typing, which we rely on heavily.

When defining protocols or data interchange formats, strict typing is
essential. JSON schemas serve as the source of truth, with types generated for
both Go and TypeScript. This ensures the contract between components is
explicit and enforced at compile time.

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

## Dependency management

Pass dependencies explicitly. There should be no hidden instantiation deep in
the call stack. Expensive instantiations happen at the composition root,
typically main.go, and dependencies are threaded down through constructors.

Define dependencies as interfaces where substitution is needed, following the
Dependency Inversion principle. This makes testing straightforward: swap real
implementations for fakes at construction time.

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
