import { createContext, type ReactNode, useContext, useId, useSyncExternalStore } from 'react'
import { createPortal } from 'react-dom'

interface BottomBarSlotContextValue {
  subscribe: (callback: () => void) => () => void
  getContainer: () => HTMLDivElement | null
  setContainer: (el: HTMLDivElement | null) => void
}

const BottomBarSlotContext = createContext<BottomBarSlotContextValue | null>(null)

export function BottomBarSlotProvider({ children }: { children: ReactNode }) {
  let container: HTMLDivElement | null = null
  const listeners = new Set<() => void>()

  const value: BottomBarSlotContextValue = {
    subscribe: (callback) => {
      listeners.add(callback)
      return () => listeners.delete(callback)
    },
    getContainer: () => container,
    setContainer: (el) => {
      container = el
      for (const listener of listeners) listener()
    },
  }

  return <BottomBarSlotContext.Provider value={value}>{children}</BottomBarSlotContext.Provider>
}

export function BottomBarSlot() {
  const context = useContext(BottomBarSlotContext)
  if (!context) throw new Error('BottomBarSlot must be used within BottomBarSlotProvider')

  return <div ref={context.setContainer} />
}

export function BottomBarPortal({ children }: { children: ReactNode }) {
  const context = useContext(BottomBarSlotContext)
  const id = useId()

  const container = useSyncExternalStore(
    context?.subscribe ?? (() => () => {}),
    context?.getContainer ?? (() => null),
  )

  if (!container) return null
  return createPortal(<div key={id}>{children}</div>, container)
}
