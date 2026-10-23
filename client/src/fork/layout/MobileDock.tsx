import { useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { useTranslation } from 'react-i18next'
import { DotsThreeIcon, PauseIcon, PlayIcon } from '@phosphor-icons/react'
import { useClient } from '@/state/provider'
import { ControlPill } from '@/fork/controls/ControlPill'
import { dockButton } from '@/fork/controls/styles'
import { MoreSheet } from '@/fork/tools/MoreSheet'
import { DockEmotes } from './DockEmotes'

// The phone's dock right under the picture, in place of the room bar: the control pill, pause, the
// emotes and More (the sheet with the sound and every tool; a dot on it while the sound is off).
export function MobileDock() {
  const { t } = useTranslation()
  const client = useClient()
  const { playable, playing, muted } = useStore(
    client.store,
    useShallow((s) => ({
      playable: s.video.playable,
      playing: s.video.playing,
      muted: s.video.muted || s.video.volume === 0,
    })),
  )
  const [open, setOpen] = useState(false)
  const more = useRef<HTMLButtonElement>(null)
  return (
    <div
      className="phone:flex hidden flex-none items-center gap-2 border-t border-glass-border bg-bar-room px-2.5 pt-2 pb-[calc(8px+env(safe-area-inset-bottom))]"
      data-testid="mobile-dock"
    >
      <div className="flex min-w-0 flex-1">
        <ControlPill compact />
      </div>
      <div className="flex flex-none items-center gap-1.5">
        <button
          type="button"
          className={dockButton}
          disabled={!playable}
          aria-label={playing ? t('fork:dock.pause') : t('fork:dock.play')}
          onClick={() => (playing ? client.pause() : client.play().catch(() => {}))}
        >
          {playing ? <PauseIcon className="size-4" /> : <PlayIcon className="size-4" />}
        </button>
        <DockEmotes />
        <button
          ref={more}
          type="button"
          className={dockButton}
          aria-label={t('fork:controls.more')}
          aria-haspopup="dialog"
          aria-expanded={open}
          data-testid="dock-more"
          onClick={() => setOpen(!open)}
        >
          <DotsThreeIcon className="size-5" weight="bold" />
          {muted && (
            <span
              className="absolute top-1.5 right-1.5 size-2 rounded-full bg-warn"
              data-testid="muted-dot"
              aria-hidden
            />
          )}
        </button>
      </div>
      {open && <MoreSheet onClose={() => setOpen(false)} opener={more.current} />}
    </div>
  )
}
