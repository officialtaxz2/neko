import { useEffect, useState } from 'react'
import { useStore } from 'zustand'
import { useNeko } from '@/state/provider'
import { fullscreenElement, requestFullscreen, setFallback, stageStore } from '@/state/stage'

// fullscreen of the picture; while fullscreen the keyboard lock lets Esc/Alt+Tab etc. reach the remote.
// the fork: also WebKit's prefixed fullscreen and the page-filling fallback (state/stage.ts), left by
// its own exit button
export function useFullscreen() {
  const neko = useNeko()
  const fallback = useStore(stageStore(neko), (s) => s.fallback)
  const [element, setElement] = useState(false)
  useEffect(() => {
    const onFs = () => {
      const fs = !!fullscreenElement()
      setElement(fs)
      const kb = (navigator as any).keyboard
      fs ? kb?.lock?.().catch(() => {}) : kb?.unlock?.()
    }
    document.addEventListener('fullscreenchange', onFs)
    document.addEventListener('webkitfullscreenchange', onFs)
    return () => {
      document.removeEventListener('fullscreenchange', onFs)
      document.removeEventListener('webkitfullscreenchange', onFs)
    }
  }, [])
  // the fallback hides the page's scrollbars while it covers the page
  useEffect(() => {
    if (!fallback) return
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      document.body.style.overflow = overflow
    }
  }, [fallback])
  return {
    fullscreen: element || fallback,
    fallback,
    request: () => void requestFullscreen(neko),
    exit: () => setFallback(neko, false),
  }
}
