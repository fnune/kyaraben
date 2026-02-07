import { Button } from '@/lib/Button'

export interface ApplyAfterUpdateBannerProps {
  readonly onApply: () => void
  readonly onDismiss: () => void
}

export function ApplyAfterUpdateBanner({ onApply, onDismiss }: ApplyAfterUpdateBannerProps) {
  return (
    <div className="bg-amber-900/50 border-b border-amber-700/50 px-4 py-3">
      <div className="flex items-center justify-between gap-4">
        <div className="flex-1 min-w-0">
          <p className="text-sm text-amber-100">
            Kyaraben was updated. Run Apply to get the latest emulator configs.
          </p>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <Button variant="secondary" onClick={onDismiss}>
            Dismiss
          </Button>
          <Button onClick={onApply}>Apply now</Button>
        </div>
      </div>
    </div>
  )
}
