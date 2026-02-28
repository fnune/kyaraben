import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './App'
import { ErrorBoundary } from './lib/ErrorBoundary'
import { HomeDirProvider } from './lib/HomeDirContext'
import './index.css'

const ZOOM_STORAGE_KEY = 'kyaraben-zoom-factor'
const DEFAULT_HANDHELD_ZOOM = 1.25

function isHandheldScreen(): boolean {
  // Steam Deck / Steam Deck OLED: 1280x800
  return window.screen.width <= 1280 && window.screen.height <= 800
}

function initZoom() {
  const saved = localStorage.getItem(ZOOM_STORAGE_KEY)
  if (saved) {
    const factor = parseFloat(saved)
    if (!Number.isNaN(factor) && factor >= 0.5 && factor <= 3) {
      window.electron.setZoomFactor(factor)
    }
  } else if (isHandheldScreen()) {
    window.electron.setZoomFactor(DEFAULT_HANDHELD_ZOOM)
  }

  let lastFactor = window.electron.getZoomFactor()
  setInterval(() => {
    const current = window.electron.getZoomFactor()
    if (current !== lastFactor) {
      lastFactor = current
      localStorage.setItem(ZOOM_STORAGE_KEY, current.toString())
    }
  }, 500)
}

initZoom()

const container = document.getElementById('root')
if (!container) {
  throw new Error('Root element not found')
}

createRoot(container).render(
  <StrictMode>
    <ErrorBoundary>
      <HomeDirProvider>
        <App />
      </HomeDirProvider>
    </ErrorBoundary>
  </StrictMode>,
)
