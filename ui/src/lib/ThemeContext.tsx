import { createContext, useContext, useEffect, useState } from 'react'
import { type FontPair, applyFontPair, fontPairs } from './fonts'
import { type ThemeDefinition, applyTheme, themes } from './themes'

interface ThemeContextValue {
  readonly theme: ThemeDefinition
  readonly fontPair: FontPair
  readonly setThemeId: (id: string) => void
  readonly setFontPairId: (id: string) => void
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [themeId, setThemeId] = useState('default')
  const [fontPairId, setFontPairId] = useState('system')

  const theme = themes.find((t) => t.id === themeId) ?? themes[0]
  const fontPair = fontPairs.find((f) => f.id === fontPairId) ?? fontPairs[0]

  useEffect(() => {
    applyTheme(theme.tokens)
  }, [theme])

  useEffect(() => {
    applyFontPair(fontPair)
  }, [fontPair])

  return (
    <ThemeContext.Provider value={{ theme, fontPair, setThemeId, setFontPairId }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider')
  return ctx
}
