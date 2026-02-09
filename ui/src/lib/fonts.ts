export interface FontPair {
  readonly id: string
  readonly name: string
  readonly description: string
  readonly heading: string
  readonly body: string
  readonly mono: string
}

export const fontPairs: FontPair[] = [
  {
    id: 'system',
    name: 'System',
    description: 'Platform default fonts',
    heading:
      'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    body: 'ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    mono: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
  },
  {
    id: 'mono',
    name: 'Mono',
    description: 'Space Mono throughout, retro terminal character',
    heading: '"Space Mono", ui-monospace, monospace',
    body: '"Space Mono", ui-monospace, monospace',
    mono: '"Space Mono", ui-monospace, monospace',
  },
  {
    id: 'industrial',
    name: 'Industrial',
    description: 'Space Grotesk headings, IBM Plex Sans body',
    heading: '"Space Grotesk", sans-serif',
    body: '"IBM Plex Sans", sans-serif',
    mono: '"IBM Plex Mono", ui-monospace, monospace',
  },
  {
    id: 'swiss',
    name: 'Swiss',
    description: 'Inter throughout, neutral Helvetica-successor clarity',
    heading: '"Inter", sans-serif',
    body: '"Inter", sans-serif',
    mono: '"JetBrains Mono", ui-monospace, monospace',
  },
  {
    id: 'editorial',
    name: 'Editorial',
    description: 'DM Serif Display headings, DM Sans body',
    heading: '"DM Serif Display", serif',
    body: '"DM Sans", sans-serif',
    mono: '"DM Mono", ui-monospace, monospace',
  },
  {
    id: 'geometric',
    name: 'Geometric',
    description: 'Outfit throughout, pure Bauhaus circles and lines',
    heading: '"Outfit", sans-serif',
    body: '"Outfit", sans-serif',
    mono: '"JetBrains Mono", ui-monospace, monospace',
  },
  {
    id: 'broadcast',
    name: 'Broadcast',
    description: 'Anybody headings, Source Sans body, 80s TV graphics',
    heading: '"Anybody", sans-serif',
    body: '"Source Sans 3", sans-serif',
    mono: '"Source Code Pro", ui-monospace, monospace',
  },
  {
    id: 'technical',
    name: 'Technical',
    description: 'Chakra Petch headings, Saira body, angular computer manual',
    heading: '"Chakra Petch", sans-serif',
    body: '"Saira", sans-serif',
    mono: '"Share Tech Mono", ui-monospace, monospace',
  },
  {
    id: 'gothic',
    name: 'Gothic',
    description: 'Bricolage Grotesque throughout, closest to ITC Serif Gothic',
    heading: '"Bricolage Grotesque", sans-serif',
    body: '"Bricolage Grotesque", sans-serif',
    mono: '"JetBrains Mono", ui-monospace, monospace',
  },
  {
    id: 'instrument',
    name: 'Instrument',
    description: 'Instrument Serif headings, Instrument Sans body',
    heading: '"Instrument Serif", serif',
    body: '"Instrument Sans", sans-serif',
    mono: '"JetBrains Mono", ui-monospace, monospace',
  },
  {
    id: 'typeset',
    name: 'Typeset',
    description: 'DM Serif Display headings, IBM Plex Sans body, IBM Plex Mono',
    heading: '"DM Serif Display", serif',
    body: '"IBM Plex Sans", sans-serif',
    mono: '"IBM Plex Mono", ui-monospace, monospace',
  },
]

export function applyFontPair(pair: FontPair) {
  const root = document.documentElement
  root.style.setProperty('--t-font-heading', pair.heading)
  root.style.setProperty('--t-font-body', pair.body)
  root.style.setProperty('--t-font-mono', pair.mono)
}
