export interface ThemeTokens {
  readonly surface: string
  readonly surfaceAlt: string
  readonly surfaceRaised: string
  readonly onSurface: string
  readonly onSurfaceSecondary: string
  readonly onSurfaceMuted: string
  readonly onSurfaceDim: string
  readonly onSurfaceFaint: string
  readonly accent: string
  readonly accentHover: string
  readonly accentMuted: string
  readonly outline: string
  readonly outlineStrong: string
  readonly logoFilter: string
  readonly colorScheme: 'dark' | 'light'
  readonly radiusCard: string
  readonly radiusElement: string
  readonly radiusControl: string
}

export interface ThemeDefinition {
  readonly id: string
  readonly name: string
  readonly description: string
  readonly tokens: ThemeTokens
}

export const themes: ThemeDefinition[] = [
  {
    id: 'default',
    name: 'Default',
    description: 'Current dark theme',
    tokens: {
      surface: '#111827',
      surfaceAlt: '#1f2937',
      surfaceRaised: '#374151',
      onSurface: '#f3f4f6',
      onSurfaceSecondary: '#d1d5db',
      onSurfaceMuted: '#9ca3af',
      onSurfaceDim: '#6b7280',
      onSurfaceFaint: '#4b5563',
      accent: '#3b82f6',
      accentHover: '#2563eb',
      accentMuted: 'rgba(59, 130, 246, 0.1)',
      outline: '#374151',
      outlineStrong: '#4b5563',
      logoFilter:
        'brightness(0) saturate(100%) invert(70%) sepia(0%) saturate(0%) hue-rotate(0deg)',
      colorScheme: 'dark',
      radiusCard: '0.75rem',
      radiusElement: '0.5rem',
      radiusControl: '0.375rem',
    },
  },
  {
    id: 'famicom',
    name: 'Famicom',
    description: 'Warm burgundy and cream, 1983',
    tokens: {
      surface: '#1a1014',
      surfaceAlt: '#281c22',
      surfaceRaised: '#382830',
      onSurface: '#f0e6d6',
      onSurfaceSecondary: '#d4c4b0',
      onSurfaceMuted: '#a89888',
      onSurfaceDim: '#786860',
      onSurfaceFaint: '#584840',
      accent: '#cc1111',
      accentHover: '#e02020',
      accentMuted: 'rgba(204, 17, 17, 0.12)',
      outline: '#382830',
      outlineStrong: '#483840',
      logoFilter:
        'brightness(0) saturate(100%) invert(80%) sepia(10%) saturate(200%) hue-rotate(340deg)',
      colorScheme: 'dark',
      radiusCard: '0.5rem',
      radiusElement: '0.375rem',
      radiusControl: '0.25rem',
    },
  },
  {
    id: 'concrete',
    name: 'Concrete',
    description: 'Industrial grey, Commodore blue',
    tokens: {
      surface: '#d0ccc6',
      surfaceAlt: '#c4c0ba',
      surfaceRaised: '#b8b4ae',
      onSurface: '#1a1a18',
      onSurfaceSecondary: '#3a3a36',
      onSurfaceMuted: '#5a5a56',
      onSurfaceDim: '#7a7a76',
      onSurfaceFaint: '#9a9a96',
      accent: '#4a52a8',
      accentHover: '#3d459c',
      accentMuted: 'rgba(74, 82, 168, 0.1)',
      outline: '#aaa6a0',
      outlineStrong: '#9a968f',
      logoFilter:
        'brightness(0) saturate(100%) invert(10%) sepia(0%) saturate(0%) hue-rotate(0deg)',
      colorScheme: 'light',
      radiusCard: '0.125rem',
      radiusElement: '0.125rem',
      radiusControl: '0.0625rem',
    },
  },
  {
    id: 'karesansui',
    name: 'Karesansui',
    description: 'Dry landscape garden, stone and vermillion',
    tokens: {
      surface: '#e6e2da',
      surfaceAlt: '#dcd8d0',
      surfaceRaised: '#d0ccc4',
      onSurface: '#1a1815',
      onSurfaceSecondary: '#3a3835',
      onSurfaceMuted: '#6a6560',
      onSurfaceDim: '#8a8580',
      onSurfaceFaint: '#aaa5a0',
      accent: '#c41e3a',
      accentHover: '#a8182f',
      accentMuted: 'rgba(196, 30, 58, 0.08)',
      outline: '#c4c0b8',
      outlineStrong: '#b0aca4',
      logoFilter:
        'brightness(0) saturate(100%) invert(8%) sepia(5%) saturate(200%) hue-rotate(20deg)',
      colorScheme: 'light',
      radiusCard: '0.75rem',
      radiusElement: '0.5rem',
      radiusControl: '0.375rem',
    },
  },
  {
    id: 'commodore',
    name: 'Commodore',
    description: 'C64 blue and brown, 1982',
    tokens: {
      surface: '#2a2176',
      surfaceAlt: '#332a82',
      surfaceRaised: '#3d3490',
      onSurface: '#b8b4ff',
      onSurfaceSecondary: '#9a96dd',
      onSurfaceMuted: '#7c78bb',
      onSurfaceDim: '#5e5a99',
      onSurfaceFaint: '#464280',
      accent: '#7c71da',
      accentHover: '#8e85e4',
      accentMuted: 'rgba(124, 113, 218, 0.15)',
      outline: '#3d3490',
      outlineStrong: '#4a42a0',
      logoFilter:
        'brightness(0) saturate(100%) invert(70%) sepia(40%) saturate(300%) hue-rotate(210deg) brightness(1.1)',
      colorScheme: 'dark',
      radiusCard: '0.25rem',
      radiusElement: '0.125rem',
      radiusControl: '0.125rem',
    },
  },
  {
    id: 'circuit',
    name: 'Circuit',
    description: 'PCB solder-mask green on warm grey',
    tokens: {
      surface: '#cbc8be',
      surfaceAlt: '#bfbcb2',
      surfaceRaised: '#b3b0a6',
      onSurface: '#1a1c18',
      onSurfaceSecondary: '#36382e',
      onSurfaceMuted: '#565850',
      onSurfaceDim: '#767870',
      onSurfaceFaint: '#969890',
      accent: '#2d7a4f',
      accentHover: '#246840',
      accentMuted: 'rgba(45, 122, 79, 0.1)',
      outline: '#a6a39a',
      outlineStrong: '#96938a',
      logoFilter:
        'brightness(0) saturate(100%) invert(10%) sepia(0%) saturate(0%) hue-rotate(0deg)',
      colorScheme: 'light',
      radiusCard: '0.25rem',
      radiusElement: '0.125rem',
      radiusControl: '0.125rem',
    },
  },
  {
    id: 'slate',
    name: 'Slate',
    description: 'Cool blue-grey with teal, late 1980s',
    tokens: {
      surface: '#1a1f2e',
      surfaceAlt: '#242a3c',
      surfaceRaised: '#30384a',
      onSurface: '#e0e4ee',
      onSurfaceSecondary: '#b8bccc',
      onSurfaceMuted: '#8a8eaa',
      onSurfaceDim: '#606488',
      onSurfaceFaint: '#464a66',
      accent: '#2aabb8',
      accentHover: '#36bfcc',
      accentMuted: 'rgba(42, 170, 187, 0.12)',
      outline: '#30384a',
      outlineStrong: '#3e4658',
      logoFilter:
        'brightness(0) saturate(100%) invert(70%) sepia(10%) saturate(200%) hue-rotate(180deg)',
      colorScheme: 'dark',
      radiusCard: '0.5rem',
      radiusElement: '0.375rem',
      radiusControl: '0.25rem',
    },
  },
]

export function applyTheme(tokens: ThemeTokens) {
  const root = document.documentElement
  root.style.setProperty('--t-surface', tokens.surface)
  root.style.setProperty('--t-surface-alt', tokens.surfaceAlt)
  root.style.setProperty('--t-surface-raised', tokens.surfaceRaised)
  root.style.setProperty('--t-on-surface', tokens.onSurface)
  root.style.setProperty('--t-on-surface-secondary', tokens.onSurfaceSecondary)
  root.style.setProperty('--t-on-surface-muted', tokens.onSurfaceMuted)
  root.style.setProperty('--t-on-surface-dim', tokens.onSurfaceDim)
  root.style.setProperty('--t-on-surface-faint', tokens.onSurfaceFaint)
  root.style.setProperty('--t-accent', tokens.accent)
  root.style.setProperty('--t-accent-hover', tokens.accentHover)
  root.style.setProperty('--t-accent-muted', tokens.accentMuted)
  root.style.setProperty('--t-outline', tokens.outline)
  root.style.setProperty('--t-outline-strong', tokens.outlineStrong)
  root.style.setProperty('--t-logo-filter', tokens.logoFilter)
  root.style.setProperty('color-scheme', tokens.colorScheme)
  root.style.setProperty('--t-radius-card', tokens.radiusCard)
  root.style.setProperty('--t-radius-element', tokens.radiusElement)
  root.style.setProperty('--t-radius-control', tokens.radiusControl)
}
