import { useEffect, type ReactNode } from 'react'
import { useBottomBar } from './BottomBarContext'

export function BottomBar({ children }: { children: ReactNode }) {
  const { setVisible } = useBottomBar()

  useEffect(() => {
    setVisible(true)
    return () => setVisible(false)
  }, [setVisible])

  return (
    <div className="fixed bottom-0 left-0 right-0 bg-gray-800/95 backdrop-blur border-t border-gray-700 px-6 py-4 z-40">
      <div className="flex items-center justify-between">{children}</div>
    </div>
  )
}
