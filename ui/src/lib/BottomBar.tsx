import { type ReactNode, useEffect } from 'react'
import { useBottomBar } from './BottomBarContext'

export const BOTTOM_BAR_HEIGHT = 'h-16'
export const SIDEBAR_BOTTOM_PADDING = 'min-[720px]:pb-16'

export function BottomBar({ children }: { children: ReactNode }) {
  const { setVisible } = useBottomBar()

  useEffect(() => {
    setVisible(true)
    return () => setVisible(false)
  }, [setVisible])

  return (
    <div
      className={`fixed bottom-0 left-0 right-0 bg-gray-800/95 backdrop-blur border-t border-gray-700 px-6 ${BOTTOM_BAR_HEIGHT} z-40 flex items-center`}
    >
      <div className="flex-1 flex items-center justify-between">{children}</div>
    </div>
  )
}
