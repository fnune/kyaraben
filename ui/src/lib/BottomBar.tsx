import type { ReactNode } from 'react'
import { BottomBarPortal } from './BottomBarSlot'

export const BOTTOM_BAR_HEIGHT = 'h-14'

export function BottomBar({ children }: { children: ReactNode }) {
  return (
    <BottomBarPortal>
      <div
        className={`bg-surface-alt/95 backdrop-blur-sm border-t-2 border-t-accent px-6 ${BOTTOM_BAR_HEIGHT} flex items-center`}
      >
        <div className="flex-1 min-w-0 flex items-center justify-between">{children}</div>
      </div>
    </BottomBarPortal>
  )
}
