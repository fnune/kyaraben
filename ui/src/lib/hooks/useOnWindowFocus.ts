import { useEffect, useRef } from 'react'

export function useOnWindowFocus(onFocus: () => void, debounceMs = 500) {
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const onFocusRef = useRef(onFocus)
  const isFirstFocusRef = useRef(true)
  onFocusRef.current = onFocus

  useEffect(() => {
    const handleFocus = () => {
      if (isFirstFocusRef.current) {
        isFirstFocusRef.current = false
        return
      }

      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }

      debounceRef.current = setTimeout(() => {
        onFocusRef.current()
      }, debounceMs)
    }

    window.addEventListener('focus', handleFocus)
    return () => {
      window.removeEventListener('focus', handleFocus)
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [debounceMs])
}
