# Claude instructions for kyaraben

NEVER AUTHOR COMMITS AS THE AI, OR ADD CO-AUTHORED-BY. ATTRIBUTE THE USER.

To quickly learn about kyaraben:

- `site/src/content/docs/index.mdx`: what kyaraben is and how it works
- `site/src/content/docs/getting-started.mdx`: installation and first-run walkthrough
- `site/src/content/docs/using-the-app.mdx`: app reference and guarantees
- `site/src/content/docs/using-the-cli.mdx`: CLI reference

Read and follow the contributing guidelines at `site/src/content/docs/contributing.mdx`.

Key points:

- Make code self-evident. Comments explaining "what" are banned. Comments explaining "why" are acceptable as a last resort if the code cannot be made clearer.
- Use sentence-case in headings. No bold text. No em-dashes. No emoji.
- Pass dependencies explicitly. No hidden instantiation. Follow SOLID principles. Good dependency injection enables using fakes over mocks in tests.
- Use `getByRole`, `getByLabel`, `getByText` in Playwright tests.
- Run `just check` before committing.
