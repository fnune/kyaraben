import { createContext, useContext, useState, type ReactNode } from 'react'

interface BottomBarContextValue {
  isVisible: boolean
  setVisible: (visible: boolean) => void
}

const BottomBarContext = createContext<BottomBarContextValue | null>(null)

export function BottomBarProvider({ children }: { children: ReactNode }) {
  const [isVisible, setVisible] = useState(false)

  return (
    <BottomBarContext.Provider value={{ isVisible, setVisible }}>
      {children}
    </BottomBarContext.Provider>
  )
}

export function useBottomBar() {
  const context = useContext(BottomBarContext)
  if (!context) {
    throw new Error('useBottomBar must be used within a BottomBarProvider')
  }
  return context
}
