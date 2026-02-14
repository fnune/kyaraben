import type { ReactNode } from 'react'

export function Shimmer() {
  return (
    <div className="h-full w-1/3 bg-linear-to-r from-transparent via-accent to-transparent animate-[shimmer_1.5s_infinite]" />
  )
}

export function ProgressBar({ percent }: { readonly percent: number }) {
  return (
    <>
      <div
        className="absolute inset-y-0 left-0 bg-accent rounded-full transition-all duration-300"
        style={{ width: `${percent}%` }}
      />
      <div className="absolute inset-0 overflow-hidden">
        <div className="h-full w-1/3 bg-linear-to-r from-transparent via-white/20 to-transparent animate-[shimmer_1.5s_infinite]" />
      </div>
    </>
  )
}

export function ProgressRail({
  children,
  className,
}: {
  readonly children: ReactNode
  readonly className?: string
}) {
  return (
    <div className={className}>
      <div className="relative overflow-hidden h-full">{children}</div>
    </div>
  )
}
