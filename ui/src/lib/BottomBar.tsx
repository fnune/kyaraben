import type { ReactNode } from 'react'
import { BottomBarPortal } from './BottomBarSlot'

export const BOTTOM_BAR_HEIGHT = 'h-14'

export function BottomBar({ children }: { children: ReactNode }) {
  return (
    <BottomBarPortal>
      <div
        className={`bg-gray-800/95 backdrop-blur border-t border-gray-700 px-6 ${BOTTOM_BAR_HEIGHT} flex items-center`}
      >
        <div className="flex-1 flex items-center justify-between">{children}</div>
      </div>
    </BottomBarPortal>
  )
}
