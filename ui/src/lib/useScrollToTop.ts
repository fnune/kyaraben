import { useEffect } from 'react'

export function useScrollToTop() {
  useEffect(() => {
    document.getElementById('main-content')?.scrollTo(0, 0)
  }, [])
}
