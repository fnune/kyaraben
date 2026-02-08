import type { ReactNode } from 'react'

export interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
}

export function Modal({ open, onClose, title, children }: ModalProps) {
  if (!open) return null

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
      className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4"
      onClick={onClose}
      onKeyDown={(e) => e.key === 'Escape' && onClose()}
    >
      <div
        role="document"
        className="bg-surface-alt rounded-element shadow-xl max-w-lg w-full max-h-[90vh] flex flex-col border border-outline"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between p-6 pb-4 shrink-0">
          <h2 id="modal-title" className="font-heading text-lg font-semibold text-on-surface">
            {title}
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="text-on-surface-dim hover:text-on-surface-secondary"
          >
            &times;
          </button>
        </div>
        <div className="px-6 pb-6 overflow-y-auto">{children}</div>
      </div>
    </div>
  )
}
