import { useEffect, useState } from 'react'
import { useNeko } from '@/state/provider'
import { describe, type DebugCounters } from '@/state/debug'

const TOUCH_EVENTS = ['touchstart', 'touchmove', 'touchend', 'touchcancel'] as const

// the fork: the `?debug` text over the picture (state/debug.ts), refreshed every second. It only
// reads: no pointer events, touches are counted on their way to the picture, never stopped
export function DebugOverlay() {
  const neko = useNeko()
  const [text, setText] = useState('')

  useEffect(() => {
    const counters: DebugCounters = { touches: 0, lastTouch: '', streamAt: 0, firstFrameAt: 0 }
    const onTouch = (e: Event) => {
      counters.touches++
      counters.lastTouch = e.type
    }
    // a new stream on the element (srcObject or an HLS source) starts with loadstart
    const onLoadStart = () => {
      counters.streamAt = Date.now()
      counters.firstFrameAt = 0
    }
    const onLoadedData = () => {
      if (counters.streamAt && !counters.firstFrameAt) counters.firstFrameAt = Date.now()
    }
    let video: HTMLVideoElement | null = null
    const follow = (next: HTMLVideoElement | null) => {
      if (next === video) return
      video?.removeEventListener('loadstart', onLoadStart)
      video?.removeEventListener('loadeddata', onLoadedData)
      video = next
      video?.addEventListener('loadstart', onLoadStart)
      video?.addEventListener('loadeddata', onLoadedData)
    }

    const offMessage = neko.client.events.on('message', (event, payload) => {
      if (event === 'signal/video') counters.video = { id: String(payload?.id ?? ''), auto: !!payload?.auto }
    })

    for (const type of TOUCH_EVENTS) window.addEventListener(type, onTouch, { capture: true, passive: true })
    const tick = () => {
      follow(neko.client.transport.element as HTMLVideoElement | null)
      setText(describe(neko, counters))
    }
    tick()
    const timer = window.setInterval(tick, 1000)
    return () => {
      clearInterval(timer)
      offMessage()
      follow(null)
      for (const type of TOUCH_EVENTS) window.removeEventListener(type, onTouch, { capture: true })
    }
  }, [neko])

  return (
    <pre
      className="pointer-events-none absolute top-2 left-2 z-30 m-0 max-h-[calc(100%-16px)] max-w-[calc(100%-16px)] overflow-hidden rounded-inner bg-black/70 px-2 py-1.5 font-mono text-[11px] leading-[1.35] max-[600px]:text-[9.5px] whitespace-pre-wrap text-white"
      data-testid="debug-overlay"
      aria-hidden
    >
      {text}
    </pre>
  )
}
