# Documentation style guide

Rules for writing content in the kyaraben documentation site.

## Prose

- Use sentence-case in headings ("Getting started", not "Getting Started")
- Do not use bold text
- Do not use em-dashes
- Do not use emoji
- Use second person ("you", not "the user" or "one")
- Use present tense ("kyaraben creates", not "kyaraben will create")
- Use imperative mood for instructions ("run the command", not "you should run
  the command")
- Use backticks for domain concepts: `System`, `Emulator`, `Provision`,
  `UserStore`, `Manifest`, `EmulatorConfig`, `KyarabenConfig`
- Use backticks for CLI commands, file paths, config keys, and code
- Link to reference pages for config details instead of duplicating them
- Keep paragraphs short

## Structure

- Every page has a frontmatter title and description
- Code blocks always have a language tag
- CLI examples show both the command and representative output
- Config examples show only the relevant keys, not the full file
- Admonitions use Starlight's `:::note`, `:::tip`, `:::caution`, `:::danger`
  syntax. Use `:::caution[Work in progress]` for features that are not yet
  stable.
- For features that are planned but not implemented, state this directly
  rather than describing them as if they exist

## Screenshots

- Screenshots live in `src/assets/screenshots/`
- Naming convention: `v<version>-<page>-<description>.png`
- Placeholder locations are marked with `{/* Screenshot: filename.png */}`
- See `src/assets/screenshots/README.md` for details

## Formatting

Markdown and MDX files are formatted with prettier. Run `npm run fmt` from
the `site/` directory, or rely on the pre-commit hook.
