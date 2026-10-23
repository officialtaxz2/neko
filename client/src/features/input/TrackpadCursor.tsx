import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectControlling } from '@m1k1o/neko'
import { cn } from '@/lib/utils'
import { useNeko } from '@/state/provider'
import { useTouch } from './useTouch'

// the fork: the trackpad's pointer (core/src/overlay.ts, setTrackpad), a mint ring drawn into the
// overlay over the picture, smaller while a finger is on the pad. Only for the host whose input
// flows, and not when the viewer hid it (the remote desktop's own pointer then shows the way).
export function TrackpadCursor() {
  const { app, client, overlay } = useNeko()
  const touch = useTouch()
  const { on, hidden } = useStore(
    app,
    useShallow((s) => ({ on: s.settings.trackpad_mode, hidden: s.settings.trackpad_cursor_hidden })),
  )
  const active = useStore(client.store, (s) => selectControlling(s) && !s.control.locked)
  const [cursor, setCursor] = useState(() => overlay.trackpadCursor)
  useEffect(() => overlay.onTrackpad(setCursor), [overlay])
  if (!on || hidden || !touch || !active) return null
  return createPortal(
    <div
      className={cn(
        'pointer-events-none absolute z-20 size-3 -translate-1/2 rounded-full border-[1.5px] border-mint bg-[rgba(12,13,18,0.45)] shadow-[0_0_8px_rgba(38,230,180,0.4)] transition-[scale,box-shadow] duration-150 ease-fluid',
        'after:absolute after:top-1/2 after:left-1/2 after:size-1 after:-translate-1/2 after:rounded-full after:bg-mint after:shadow-[0_0_4px_var(--color-mint)]',
        cursor.touching && 'scale-80 shadow-[0_0_12px_var(--color-mint)]',
      )}
      style={{ left: `${cursor.x * 100}%`, top: `${cursor.y * 100}%` }}
      data-testid="trackpad-cursor"
      data-touching={cursor.touching}
    />,
    overlay.wrap,
  )
}
