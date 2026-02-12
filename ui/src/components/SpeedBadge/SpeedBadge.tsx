import { useEffect, useRef, useState } from 'react'

export interface SpeedBadgeProps {
  readonly speedBytes: number
  readonly show: boolean
}

export function SpeedBadge({ speedBytes, show }: SpeedBadgeProps) {
  const [displaySpeed, setDisplaySpeed] = useState(0)
  const [displayUnit, setDisplayUnit] = useState<'KB/s' | 'MB/s'>('MB/s')
  const speedTargetRef = useRef(0)
  const unitRef = useRef<'KB/s' | 'MB/s'>('MB/s')

  useEffect(() => {
    const mbps = Math.round(speedBytes / (1024 * 1024))
    const kbps = Math.round(speedBytes / 1024)
    const nextUnit = mbps > 0 ? 'MB/s' : 'KB/s'
    const nextValue = mbps > 0 ? mbps : kbps

    if (nextValue === 0) {
      speedTargetRef.current = 0
      return
    }

    if (unitRef.current !== nextUnit) {
      unitRef.current = nextUnit
      setDisplayUnit(nextUnit)
      setDisplaySpeed(Math.max(0, nextValue))
      speedTargetRef.current = Math.max(0, nextValue)
      return
    }

    speedTargetRef.current = Math.max(0, nextValue)
  }, [speedBytes])

  useEffect(() => {
    const timer = setInterval(() => {
      setDisplaySpeed((current) => {
        const target = speedTargetRef.current
        if (current === target) return current
        const delta = target - current
        const step = Math.sign(delta) * Math.max(1, Math.floor(Math.abs(delta) / 3))
        const next = current + step
        if ((delta > 0 && next > target) || (delta < 0 && next < target)) {
          return target
        }
        return next
      })
    }, 80)
    return () => clearInterval(timer)
  }, [])

  if (!show) return null

  return (
    <span
      className={`text-xs font-mono ${
        displaySpeed === 0 ? 'text-on-surface-dim' : 'text-status-ok'
      }`}
    >
      {displaySpeed} {displayUnit}
    </span>
  )
}
